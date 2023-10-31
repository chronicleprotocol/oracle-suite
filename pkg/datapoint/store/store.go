//  Copyright (C) 2021-2023 Chronicle Labs, Inc.
//
//  This program is free software: you can redistribute it and/or modify
//  it under the terms of the GNU Affero General Public License as
//  published by the Free Software Foundation, either version 3 of the
//  License, or (at your option) any later version.
//
//  This program is distributed in the hope that it will be useful,
//  but WITHOUT ANY WARRANTY; without even the implied warranty of
//  MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.  See the
//  GNU Affero General Public License for more details.
//
//  You should have received a copy of the GNU Affero General Public License
//  along with this program.  If not, see <http://www.gnu.org/licenses/>.

package store

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"time"

	"github.com/defiweb/go-eth/types"

	"github.com/chronicleprotocol/oracle-suite/pkg/contract/chronicle"
	"github.com/chronicleprotocol/oracle-suite/pkg/datapoint"
	"github.com/chronicleprotocol/oracle-suite/pkg/datapoint/value"
	"github.com/chronicleprotocol/oracle-suite/pkg/log"
	"github.com/chronicleprotocol/oracle-suite/pkg/log/null"
	"github.com/chronicleprotocol/oracle-suite/pkg/transport"
	"github.com/chronicleprotocol/oracle-suite/pkg/transport/messages"
	"github.com/chronicleprotocol/oracle-suite/pkg/util/bn"
)

const LoggerTag = "DATA_POINT_STORE"

// DataPointProvider is an interface which provides latest data points from
// feeds.
type DataPointProvider interface {
	// LatestFrom returns the latest data point from a given address.
	LatestFrom(ctx context.Context, from types.Address, model string) (StoredDataPoint, bool, error)

	// Latest returns the latest data points from all addresses.
	Latest(ctx context.Context, model string) (map[types.Address]StoredDataPoint, error)
}

// Storage is underlying storage implementation for the Store.
//
// It must be thread-safe.
type Storage interface {
	// Add adds a data point to the store.
	//
	// Adding a data point with a timestamp older than the latest data point
	// for the same address and model will be ignored.
	Add(ctx context.Context, point StoredDataPoint) error

	// LatestFrom returns the latest data point from a given address.
	LatestFrom(ctx context.Context, from types.Address, model string) (point StoredDataPoint, ok bool, err error)

	// Latest returns the latest data points from all addresses.
	Latest(ctx context.Context, model string) (points map[types.Address]StoredDataPoint, err error)
}

// StoredDataPoint is a struct which represents a data point stored in the
// Store.
type StoredDataPoint struct {
	Model     string
	DataPoint datapoint.Point
	From      types.Address
	Signature types.Signature
}

func StoredDataPointLogFields(o StoredDataPoint) log.Fields {
	f := log.Fields{
		"point.model":     o.Model,
		"point.from":      o.From.String(),
		"point.signature": o.Signature.String(),
	}
	for k, v := range datapoint.PointLogFields(o.DataPoint) {
		f[k] = v
	}
	return f
}

// Store stores latest data points from feeds.
type Store struct {
	ctx    context.Context
	waitCh chan error
	log    log.Logger

	storage    Storage
	transport  transport.Service
	models     []string
	recoverers []datapoint.Recoverer
}

// Config is the configuration for Storage.
type Config struct {
	// Storage is the storage implementation.
	Storage Storage

	// Transport is an implementation of transport used to fetch prices from feeds.
	Transport transport.Service

	// Models is the list of models which are supported by the store.
	Models []string

	// Recoverers is the list of recoverers which are used to recover the
	// feed's address from the data point.
	Recoverers []datapoint.Recoverer

	// Logger is a current logger interface used by the Store.
	// The Logger is required to monitor asynchronous processes.
	Logger log.Logger
}

// New creates a new Store.
func New(cfg Config) (*Store, error) {
	if cfg.Logger == nil {
		cfg.Logger = null.New()
	}
	if cfg.Storage == nil {
		return nil, errors.New("storage must not be nil")
	}
	if cfg.Transport == nil {
		return nil, errors.New("transport must not be nil")
	}
	s := &Store{
		waitCh:     make(chan error),
		log:        cfg.Logger.WithField("tag", LoggerTag),
		storage:    cfg.Storage,
		transport:  cfg.Transport,
		models:     cfg.Models,
		recoverers: cfg.Recoverers,
	}
	return s, nil
}

// Start implements the supervisor.Service interface.
func (p *Store) Start(ctx context.Context) error {
	if p.ctx != nil {
		return errors.New("service can be started only once")
	}
	if ctx == nil {
		return errors.New("context must not be nil")
	}
	p.log.Info("Starting")
	p.ctx = ctx
	go p.dataPointCollectorRoutine()
	go p.logDataPointsRoutine()
	go p.contextCancelHandler()
	return nil
}

// Wait implements the supervisor.Service interface.
func (p *Store) Wait() <-chan error {
	return p.waitCh
}

// LatestFrom implements the DataPointProvider interface.
func (p *Store) LatestFrom(ctx context.Context, from types.Address, model string) (StoredDataPoint, bool, error) {
	return p.storage.LatestFrom(ctx, from, model)
}

// Latest implements the DataPointProvider interface.
func (p *Store) Latest(ctx context.Context, model string) (map[types.Address]StoredDataPoint, error) {
	return p.storage.Latest(ctx, model)
}

func (p *Store) collectDataPoint(point *messages.DataPoint) {
	for _, recoverer := range p.recoverers {
		if recoverer.Supports(p.ctx, point.Point) {
			from, err := recoverer.Recover(p.ctx, point.Model, point.Point, point.ECDSASignature)
			if err != nil {
				p.log.
					WithError(err).
					WithFields(log.Fields{
						"model": point.Model,
						"from":  from,
					}).
					WithFields(datapoint.PointLogFields(point.Point)).
					WithAdvice("This is a sign of a misbehaving feed or a serious bug in the feed software").
					Error("Unable to recover address from the data point")
				return
			}
			sdp := StoredDataPoint{
				Model:     point.Model,
				DataPoint: point.Point,
				From:      *from,
				Signature: point.ECDSASignature,
			}
			if err := p.storage.Add(p.ctx, sdp); err != nil {
				p.log.
					WithError(err).
					WithFields(StoredDataPointLogFields(sdp)).
					Error("Unable to add data point to the storage")
				return
			}
			p.log.
				WithFields(StoredDataPointLogFields(sdp)).
				Debug("Data point collected")
			return
		}
	}
	p.log.
		WithField("model", point.Model).
		WithFields(datapoint.PointLogFields(point.Point)).
		WithAdvice("This is probably caused by misconfigured feed or an error in the data model").
		Error("Unable to find recoverer for the data point")
}

func (p *Store) shouldCollect(model string) bool {
	for _, a := range p.models {
		if a == model {
			return true
		}
	}
	return false
}

func (p *Store) handlePointMessage(msg transport.ReceivedMessage) {
	if msg.Error != nil {
		p.log.
			WithError(msg.Error).
			WithAdvice("Ignore if occurs occasionally, especially if it is related to temporary network issues").
			Error("Unable to receive a message from the transport layer")
		return
	}
	point, ok := msg.Message.(*messages.DataPoint)
	if !ok {
		p.log.
			WithFields(transport.ReceivedMessageFields(msg)).
			WithField("type", fmt.Sprintf("%T", msg.Message)).
			WithAdvice("This is a bug and must be investigated").
			Error("Unexpected value returned from the transport layer")
		return
	}
	if !p.shouldCollect(point.Model) {
		p.log.
			WithFields(transport.ReceivedMessageFields(msg)).
			WithField("model", point.Model).
			Debug("Data point rejected, model is not supported")
		return
	}
	p.collectDataPoint(point)
}

// handleLegacyPriceMessage handles legacy price messages and converts them to
// data points. This is temporary solution until the price messages are
// completely removed.
//
// TODO: Remove this method when the price messages are removed.
func (p *Store) handleLegacyPriceMessage(msg transport.ReceivedMessage) {
	if msg.Error != nil {
		p.log.
			WithError(msg.Error).
			WithAdvice("Ignore if occurs occasionally, especially if it is related to temporary network issues").
			Error("Unable to receive a message from the transport layer")
		return
	}
	price, ok := msg.Message.(*messages.Price)
	if !ok {
		p.log.
			WithFields(transport.ReceivedMessageFields(msg)).
			WithField("type", fmt.Sprintf("%T", msg.Message)).
			WithAdvice("This is a bug and must be investigated").
			Error("Unexpected value returned from the transport layer")
		return
	}
	trace := make(map[string]string)
	_ = json.Unmarshal(price.Trace, &trace)
	point := &messages.DataPoint{
		Model: price.Price.Wat,
		Point: datapoint.Point{
			Value: value.Tick{
				Pair:  findPairForLegacyPrice(price.Price.Wat),
				Price: bn.DecFixedPointFromRawBigInt(price.Price.Val, chronicle.MedianPricePrecision).DecFloatPoint(),
			},
			Time:      price.Price.Age,
			SubPoints: nil,
			Meta: map[string]any{
				"legacy": true,
				"trace":  trace,
			},
		},
		ECDSASignature: price.Price.Sig,
	}
	if !p.shouldCollect(point.Model) {
		p.log.
			WithFields(transport.ReceivedMessageFields(msg)).
			WithField("model", point.Model).
			Debug("Data point rejected, model is not supported")
		return
	}
	p.collectDataPoint(point)
}

// logDataPointsSince logs a short summary of data points collected since the
// given time.
func (p *Store) logDataPointsSince(since time.Time) {
	dataPointsLog := make(map[string]any)
	for _, model := range p.models {
		dataPoints, err := p.storage.Latest(p.ctx, model)
		if err != nil {
			p.log.
				WithError(err).
				Error("Failed to fetch latest data points")
			continue
		}
		for _, dp := range dataPoints {
			if dp.DataPoint.Time.After(since) {
				key := fmt.Sprintf("%s:%s", dp.From.String(), model)
				dataPointsLog[key] = dp.DataPoint.Value
			}
		}
	}
	p.log.
		WithField("dataPoints", dataPointsLog).
		Info("Collected data points in the last minute")
}

func (p *Store) logDataPointsRoutine() {
	lastSummaryTime := time.Time{}
	summaryInterval := time.NewTicker(1 * time.Minute)
	defer summaryInterval.Stop()
	for {
		select {
		case <-p.ctx.Done():
			return
		case t := <-summaryInterval.C:
			p.logDataPointsSince(lastSummaryTime)
			lastSummaryTime = t
		}
	}
}

func (p *Store) dataPointCollectorRoutine() {
	dataPointCh := p.transport.Messages(messages.DataPointV1MessageName)
	priceCh := p.transport.Messages(messages.PriceV0MessageName) //nolint:staticcheck
	for {
		select {
		case <-p.ctx.Done():
			return
		case msg := <-dataPointCh:
			p.handlePointMessage(msg)
		case msg := <-priceCh:
			p.handleLegacyPriceMessage(msg)
		}
	}
}

// contextCancelHandler handles context cancellation.
func (p *Store) contextCancelHandler() {
	defer func() { close(p.waitCh) }()
	defer p.log.Info("Stopped")
	<-p.ctx.Done()
}

func findPairForLegacyPrice(model string) value.Pair {
	if pair, ok := legacyPricePairs[model]; ok {
		return pair
	}
	// It is ok to return unknown pair here, because this value is currently
	// not used anywhere. Also, this should never happen because none of the
	// feeds broadcast prices other than the ones in the legacyPricePairs map.
	return value.Pair{Base: "UNKNOWN", Quote: "UNKNOWN"}
}

var legacyPricePairs = map[string]value.Pair{
	"AAVEUSD":   {Base: "AAVE", Quote: "USD"},
	"ARBUSD":    {Base: "ARB", Quote: "USD"},
	"AVAXUSD":   {Base: "AVAX", Quote: "USD"},
	"BNBUSD":    {Base: "BNB", Quote: "USD"},
	"BTCUSD":    {Base: "BTC", Quote: "USD"},
	"CRVUSD":    {Base: "CRV", Quote: "USD"},
	"DAIUSD":    {Base: "DAI", Quote: "USD"},
	"DSRRATE":   {Base: "DSR", Quote: "RATE"},
	"ETHBTC":    {Base: "ETH", Quote: "BTC"},
	"ETHUSD":    {Base: "ETH", Quote: "USD"},
	"FRAXUSD":   {Base: "FRAX", Quote: "USD"},
	"GNOUSD":    {Base: "GNO", Quote: "USD"},
	"IBTAUSD":   {Base: "IBTA", Quote: "USD"},
	"LDOUSD":    {Base: "LDO", Quote: "USD"},
	"LINKUSD":   {Base: "LINK", Quote: "USD"},
	"MATICUSD":  {Base: "MATIC", Quote: "USD"},
	"MKRETH":    {Base: "MKR", Quote: "ETH"},
	"MKRUSD":    {Base: "MKR", Quote: "USD"},
	"OPUSD":     {Base: "OP", Quote: "USD"},
	"RETHETH":   {Base: "RETH", Quote: "ETH"},
	"RETHUSD":   {Base: "RETH", Quote: "USD"},
	"SDAIDAI":   {Base: "SDAI", Quote: "DAI"},
	"SDAIETH":   {Base: "SDAI", Quote: "ETH"},
	"SDAIMATIC": {Base: "SDAI", Quote: "MATIC"},
	"SDAIUSD":   {Base: "SDAI", Quote: "USD"},
	"SNXUSD":    {Base: "SNX", Quote: "USD"},
	"SOLUSD":    {Base: "SOL", Quote: "USD"},
	"STETHETH":  {Base: "STETH", Quote: "ETH"},
	"STETHUSD":  {Base: "STETH", Quote: "USD"},
	"UNIUSD":    {Base: "UNI", Quote: "USD"},
	"USDCUSD":   {Base: "USDC", Quote: "USD"},
	"USDTUSD":   {Base: "USDT", Quote: "USD"},
	"WBTCUSD":   {Base: "WBTC", Quote: "USD"},
	"WSTETHETH": {Base: "WSTETH", Quote: "ETH"},
	"WSTETHUSD": {Base: "WSTETH", Quote: "USD"},
	"XTZUSD":    {Base: "XTZ", Quote: "USD"},
	"YFIUSD":    {Base: "YFI", Quote: "USD"},
	"MANAUSD":   {Base: "MANA", Quote: "USD"},
}
