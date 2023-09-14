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

package feed

import (
	"context"
	"encoding/json"
	"errors"

	"github.com/chronicleprotocol/oracle-suite/pkg/datapoint"
	"github.com/chronicleprotocol/oracle-suite/pkg/log/null"
	"github.com/chronicleprotocol/oracle-suite/pkg/transport/messages"
	"github.com/chronicleprotocol/oracle-suite/pkg/util/sliceutil"
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
	log    log.Logger

	dataProvider datapoint.Provider
	dataModels   []string
	signers      []datapoint.Signer
	hooks        []Hook
	transport    transport.Service
	interval     *timeutil.Ticker
}

// Config is the configuration for the Feed.
type Config struct {
	// DataModels is a list of data models handled by the Feed.
	DataModels []string

	// DataProvider is a data provider which is used to fetch data points.
	DataProvider datapoint.Provider

	// Signers is a list of signers used to sign data points.
	//
	// If none of the provided signers can sign the data point, it will be
	// skipped.
	Signers []datapoint.Signer

	// Hooks is a list of hooks that will be called before broadcasting
	// data points.
	Hooks []Hook

	// Transport is an implementation of transport used to send data points to
	// the network.
	Transport transport.Service

	// Interval describes how often data points should be sent to the network.
	Interval *timeutil.Ticker

	// Logger is a current logger interface used by the Feed.
	// If nil, null logger will be used.
	Logger log.Logger
}

type Hook interface {
	BeforeSign(ctx context.Context, dp *datapoint.Point) error
	BeforeBroadcast(ctx context.Context, dp *datapoint.Point) error
}

// New creates a new instance of the Feed.
func New(cfg Config) (*Feed, error) {
	if cfg.DataModels == nil {
		return nil, errors.New("data models must not be nil")
	}
	if cfg.DataProvider == nil {
		return nil, errors.New("data provider must not be nil")
	}
	if cfg.Transport == nil {
		return nil, errors.New("transport must not be nil")
	}
	if len(cfg.DataModels) == 0 {
		return nil, errors.New("at least one data model must be provided")
	}
	if len(cfg.Signers) == 0 {
		return nil, errors.New("at least one signer must be provided")
	}
	if cfg.Logger == nil {
		cfg.Logger = null.New()
	}
	f := &Feed{
		waitCh:       make(chan error),
		log:          cfg.Logger.WithField("tag", LoggerTag),
		dataProvider: cfg.DataProvider,
		dataModels:   cfg.DataModels,
		signers:      cfg.Signers,
		hooks:        cfg.Hooks,
		transport:    cfg.Transport,
		interval:     cfg.Interval,
	}
	return f, nil
}

// Start implements the supervisor.Service interface.
func (f *Feed) Start(ctx context.Context) error {
	if f.ctx != nil {
		return errors.New("service can be started only once")
	}
	if ctx == nil {
		return errors.New("context must not be nil")
	}
	f.ctx = ctx
	f.log.
		WithFields(log.Fields{
			"dataModels": f.dataModels,
			"interval":   f.interval.Duration(),
		}).
		Debug("Starting")
	f.interval.Start(f.ctx)
	go f.broadcasterRoutine()
	go f.contextCancelHandler()
	return nil
}

// Wait implements the supervisor.Service interface.
func (f *Feed) Wait() <-chan error {
	return f.waitCh
}

// broadcast sends data point to the network.
func (f *Feed) broadcast(model string, point datapoint.Point) {
	found := false
	for _, signer := range f.signers {
		if !signer.Supports(f.ctx, point) {
			continue
		}
		found = true

		// BeforeSign hook.
		for _, hook := range f.hooks {
			if err := hook.BeforeSign(f.ctx, &point); err != nil {
				f.log.
					WithError(err).
					WithFields(datapoint.PointLogFields(point)).
					Error("BeforeBroadcast hook failed")
				return
			}
		}

		// Sign data point.
		sig, err := signer.Sign(f.ctx, model, point)
		if err != nil {
			f.log.
				WithError(err).
				WithFields(datapoint.PointLogFields(point)).
				Error("Unable to sign data point")
		}

		// BeforeBroadcast hook.
		for _, hook := range f.hooks {
			if err := hook.BeforeBroadcast(f.ctx, &point); err != nil {
				f.log.
					WithError(err).
					WithFields(datapoint.PointLogFields(point)).
					Error("BeforeBroadcast hook failed")
				return
			}
		}

		// Broadcast data point.
		msg := &messages.DataPoint{
			Model:          model,
			Point:          point,
			ECDSASignature: *sig,
		}
		if err := f.transport.Broadcast(messages.DataPointV1MessageName, msg); err != nil {
			f.log.
				WithError(err).
				WithFields(messages.DataPointMessageLogFields(*msg)).
				Error("Unable to broadcast data point")
		} else {
			f.log.
				WithFields(messages.DataPointMessageLogFields(*msg)).
				Info("Data point broadcast")
		}
	}
	if !found {
		f.log.
			WithField("model", model).
			WithFields(datapoint.PointLogFields(point)).
			Warn("Unable to find signer for data point")
	}
}

func (f *Feed) broadcasterRoutine() {
	for {
		select {
		case <-f.ctx.Done():
			return
		case <-f.interval.TickCh():
			// Fetch data points from the data provider.
			points, err := f.dataProvider.DataPoints(
				f.ctx,
				sliceutil.Intersect(
					f.dataProvider.ModelNames(f.ctx),
					f.dataModels,
				)...,
			)
			if err != nil {
				f.log.
					WithError(err).
					Error("Unable to update data points")
				continue
			}

			// Send data points to the network.
			for model, point := range points {
				if err := point.Validate(); err != nil {
					if log.IsLevel(f.log, log.Debug) {
						trace, _ := json.Marshal(point)
						f.log.
							WithError(err).
							WithFields(datapoint.PointLogFields(point)).
							WithField("trace", string(trace)).
							Debug("Invalid data point trace")
					}
					f.log.
						WithError(err).
						WithFields(datapoint.PointLogFields(point)).
						Error("Unable to broadcast data point, data point is invalid")
					continue
				}
				f.broadcast(model, point)
			}
		}
	}
}

func (f *Feed) contextCancelHandler() {
	defer func() { close(f.waitCh) }()
	defer f.log.Info("Stopped")
	<-f.ctx.Done()
}
