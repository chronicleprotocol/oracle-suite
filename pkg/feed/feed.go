package feed

import (
	"context"
	"errors"

	"github.com/defiweb/go-eth/wallet"

	"github.com/chronicleprotocol/oracle-suite/pkg/data"
	"github.com/chronicleprotocol/oracle-suite/pkg/log/null"
	"github.com/chronicleprotocol/oracle-suite/pkg/transport/messages"
	"github.com/chronicleprotocol/oracle-suite/pkg/util/timeutil"

	"github.com/chronicleprotocol/oracle-suite/pkg/log"
	"github.com/chronicleprotocol/oracle-suite/pkg/transport"
)

const LoggerTag = "FEED"

// Feed is a service which periodically fetches prices and then sends them to
// the Oracle network using transport layer.
type Feed struct {
	ctx    context.Context
	waitCh chan error

	dataProvider data.Provider
	dataModels   []string
	handlers     []DataPointHandler
	signer       wallet.Key
	transport    transport.Transport
	interval     *timeutil.Ticker
	log          log.Logger
}

type DataPointHandler interface {
	Supports(data.Point) bool
	Handle(model string, point data.Point) (*messages.Event, error)
}

// Config is the configuration for the Feed.
type Config struct {
	DataModels []string

	DataProvider data.Provider

	Handlers []DataPointHandler

	// Transport is an implementation of transport used to send prices to
	// the network.
	Transport transport.Transport

	// Interval describes how often we should send prices to the network.
	Interval *timeutil.Ticker

	// Logger is a current logger interface used by the Feed.
	Logger log.Logger
}

// New creates a new instance of the Feed.
func New(cfg Config) (*Feed, error) {
	if cfg.DataModels == nil {
		return nil, errors.New("data provider must not be nil")
	}
	if cfg.Transport == nil {
		return nil, errors.New("transport must not be nil")
	}
	if cfg.Logger == nil {
		cfg.Logger = null.New()
	}
	g := &Feed{
		waitCh:       make(chan error),
		dataProvider: cfg.DataProvider,
		dataModels:   cfg.DataModels,
		handlers:     cfg.Handlers,
		transport:    cfg.Transport,
		interval:     cfg.Interval,
		log:          cfg.Logger.WithField("tag", LoggerTag),
	}
	return g, nil
}

// Start implements the supervisor.Service interface.
func (g *Feed) Start(ctx context.Context) error {
	if g.ctx != nil {
		return errors.New("service can be started only once")
	}
	if ctx == nil {
		return errors.New("context must not be nil")
	}
	g.log.Infof("Starting")
	g.ctx = ctx
	g.interval.Start(g.ctx)
	go g.broadcasterRoutine()
	go g.contextCancelHandler()
	return nil
}

// Wait implements the supervisor.Service interface.
func (g *Feed) Wait() <-chan error {
	return g.waitCh
}

// broadcast sends price for single pair to the network. This method uses
// current price from the Provider, so it must be updated beforehand.
func (g *Feed) broadcast(model string, point data.Point) {
	handlerFound := false
	for _, handler := range g.handlers {
		if !handler.Supports(point) {
			continue
		}
		handlerFound = true
		event, err := handler.Handle(model, point)
		if err != nil {
			g.log.
				WithError(err).
				WithField("dataPoint", point).
				Warn("Unable to handle data point")
		}
		if err := g.transport.Broadcast(messages.EventV1MessageName, event); err != nil {
			g.log.
				WithError(err).
				WithField("dataPoint", point).
				Warn("Unable to broadcast data point")
		}
		g.log.
			WithField("dataPoint", point).
			Info("Data point broadcast")
	}
	if !handlerFound {
		g.log.
			WithField("dataPoint", point).
			Warn("Unable to find handler for data point")
	}
}

func (g *Feed) broadcasterRoutine() {
	for {
		select {
		case <-g.ctx.Done():
			return
		case <-g.interval.TickCh():
			// Update all data points.
			points, err := g.dataProvider.DataPoints(g.ctx, g.dataModels...)
			if err != nil {
				g.log.
					WithError(err).
					Warn("Unable to update data points")
				continue
			}

			// Send data points to the network.
			for model, point := range points {
				g.broadcast(model, point)
			}
		}
	}
}

func (g *Feed) contextCancelHandler() {
	defer func() { close(g.waitCh) }()
	defer g.log.Info("Stopped")
	<-g.ctx.Done()
}
