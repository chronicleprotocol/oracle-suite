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
	"errors"

	"github.com/defiweb/go-eth/types"

	"github.com/chronicleprotocol/oracle-suite/pkg/datapoint"
	"github.com/chronicleprotocol/oracle-suite/pkg/log"
	"github.com/chronicleprotocol/oracle-suite/pkg/log/null"
	"github.com/chronicleprotocol/oracle-suite/pkg/transport"
	"github.com/chronicleprotocol/oracle-suite/pkg/transport/messages"
)

const LoggerTag = "DATA_POINT_STORE"

// Storage is underlying storage implementation for the Store.
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

// LogFields returns a set of log fields for the data point.
func (o StoredDataPoint) LogFields() log.Fields {
	f := log.Fields{
		"model":     o.Model,
		"from":      o.From.String(),
		"signature": o.Signature.String(),
	}
	for k, v := range o.DataPoint.LogFields() {
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
	go p.contextCancelHandler()
	return nil
}

// Wait implements the supervisor.Service interface.
func (p *Store) Wait() <-chan error {
	return p.waitCh
}

// LatestFrom returns the latest data point from a given address.
func (p *Store) LatestFrom(ctx context.Context, from types.Address, model string) (StoredDataPoint, bool, error) {
	return p.storage.LatestFrom(ctx, from, model)
}

// Latest returns the latest data points from all addresses.
func (p *Store) Latest(ctx context.Context, model string) (map[types.Address]StoredDataPoint, error) {
	return p.storage.Latest(ctx, model)
}

func (p *Store) collectDataPoint(point *messages.DataPoint) {
	for _, recoverer := range p.recoverers {
		if recoverer.Supports(p.ctx, point.Value) {
			from, err := recoverer.Recover(p.ctx, point.Model, point.Value, point.Signature)
			if err != nil {
				p.log.
					WithError(err).
					WithField("model", point.Model).
					WithField("from", from).
					WithFields(point.Value.LogFields()).
					Error("Unable to recover address")
			}
			sdp := StoredDataPoint{
				Model:     point.Model,
				DataPoint: point.Value,
				From:      *from,
				Signature: point.Signature,
			}
			if err := p.storage.Add(p.ctx, sdp); err != nil {
				p.log.
					WithError(err).
					WithFields(sdp.LogFields()).
					Error("Unable to add data point")
				return
			}
			p.log.
				WithFields(sdp.LogFields()).
				Info("Data point collected")
			return
		}
	}
	p.log.
		WithField("model", point.Model).
		WithFields(point.Value.LogFields()).
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

func (p *Store) dataPointCollectorRoutine() {
	dataPointCh := p.transport.Messages(messages.DataPointV1MessageName)
	for {
		select {
		case <-p.ctx.Done():
			return
		case msg := <-dataPointCh:
			p.handlePointMessage(msg)
		}
	}
}

// contextCancelHandler handles context cancellation.
func (p *Store) contextCancelHandler() {
	defer func() { close(p.waitCh) }()
	defer p.log.Info("Stopped")
	<-p.ctx.Done()
}

func (p *Store) handlePointMessage(msg transport.ReceivedMessage) {
	if msg.Error != nil {
		p.log.
			WithError(msg.Error).
			Error("Unable to receive message")
		return
	}
	point, ok := msg.Message.(*messages.DataPoint)
	if !ok {
		p.log.
			WithFields(msg.Fields()).
			Error("Unexpected value returned from the transport layer")
		return
	}
	if !p.shouldCollect(point.Model) {
		p.log.
			WithFields(msg.Fields()).
			WithField("model", point.Model).
			Warn("Data point rejected - model is not supported")
		return
	}
	p.collectDataPoint(point)
}
