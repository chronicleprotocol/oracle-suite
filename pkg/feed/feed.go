package feed

import (
	"context"
	"errors"

	"github.com/chronicleprotocol/oracle-suite/pkg/data"
	"github.com/chronicleprotocol/oracle-suite/pkg/log/null"
	"github.com/chronicleprotocol/oracle-suite/pkg/transport/messages"
	"github.com/chronicleprotocol/oracle-suite/pkg/util/timeutil"

	"github.com/chronicleprotocol/oracle-suite/pkg/log"
	"github.com/chronicleprotocol/oracle-suite/pkg/transport"
)

const LoggerTag = "FEED"

// Feed is a service which periodically fetches data points and then sends them to
// the network using transport layer.
type Feed struct {
	ctx    context.Context
	waitCh chan error

	dataProvider data.Provider
	dataModels   []string
	handlers     []DataPointHandler
	transport    transport.Transport
	interval     *timeutil.Ticker
	log          log.Logger
}

// DataPointHandler converts data point to the event message.
type DataPointHandler interface {
	// Supports returns true if the handler supports given data point.
	Supports(data.Point) bool

	// Handle converts data point to the event message.
	Handle(model string, point data.Point) (*messages.Event, error)
}

// Config is the configuration for the Feed.
type Config struct {
	// DataModels is a list of data models handled by the Feed.
	DataModels []string

	// DataProvider is a data provider which is used to fetch data points.
	DataProvider data.Provider

	// Handlers is a list of handlers used to convert data points to the
	// event messages.
	Handlers []DataPointHandler

	// Transport is an implementation of transport used to send prices to
	// the network.
	Transport transport.Transport

	// Interval describes how often data points should be sent to the network.
	Interval *timeutil.Ticker

	// Logger is a current logger interface used by the Feed.
	// If nil, null logger will be used.
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

// broadcast sends data point to the network.
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
				Error("Unable to handle data point")
		}
		if err := g.transport.Broadcast(messages.EventV1MessageName, event); err != nil {
			g.log.
				WithError(err).
				WithField("dataPoint", point).
				Error("Unable to broadcast data point")
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
			// Fetch all data points from the provider to update them
			// at once.
			_, err := g.dataProvider.DataPoints(g.ctx, g.dataModels...)
			if err != nil {
				g.log.
					WithError(err).
					Error("Unable to update data points")
				continue
			}

			// Send data points to the network.
			for _, model := range g.dataModels {
				point, err := g.dataProvider.DataPoint(g.ctx, model)
				if err != nil {
					g.log.
						WithError(err).
						Error("Unable to get data points")
					continue
				}
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
