//  Copyright (C) 2020 Maker Ecosystem Growth Holdings, INC.
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

package publisher

import (
	"context"
	"encoding/hex"
	"errors"

	"github.com/chronicleprotocol/oracle-suite/pkg/log"
	"github.com/chronicleprotocol/oracle-suite/pkg/transport"
	"github.com/chronicleprotocol/oracle-suite/pkg/transport/messages"
)

const LoggerTag = "EVENT_OBSERVER"

// EventPublisher collects event messages from listeners and publishes them
// using transport.
type EventPublisher struct {
	ctx    context.Context
	waitCh chan error

	signers   []Signer
	listeners []Listener
	transport transport.Transport
	log       log.Logger
}

// Config contains configuration parameters for EventPublisher.
type Config struct {
	Listeners []Listener
	// Signer is a list of Signers used to sign events.
	Signers []Signer
	// Transport is implementation of transport used to send events to relayers.
	Transport transport.Transport
	// Logger is a current logger interface used by the EventPublisher. The Logger
	// helps to monitor asynchronous processes.
	Logger log.Logger
}

type Listener interface {
	Start(ctx context.Context) error
	Events() chan *messages.Event
}

type Signer interface {
	Sign(event *messages.Event) (bool, error)
}

// New returns a new instance of the EventPublisher struct.
func New(ctx context.Context, cfg Config) (*EventPublisher, error) {
	if ctx == nil {
		return nil, errors.New("context must not be nil")
	}
	return &EventPublisher{
		ctx:       ctx,
		waitCh:    make(chan error),
		transport: cfg.Transport,
		listeners: cfg.Listeners,
		signers:   cfg.Signers,
		log:       cfg.Logger.WithField("tag", LoggerTag),
	}, nil
}

func (l *EventPublisher) Start() error {
	l.log.Infof("Starting")
	l.listenerLoop()
	for _, lis := range l.listeners {
		err := lis.Start(l.ctx)
		if err != nil {
			return err
		}
	}
	go l.contextCancelHandler()
	return nil
}

// Wait waits until the context is canceled or until an error occurs.
func (l *EventPublisher) Wait() chan error {
	return l.waitCh
}

func (l *EventPublisher) listenerLoop() {
	for _, li := range l.listeners {
		li := li
		go func() {
			for {
				select {
				case <-l.ctx.Done():
					return
				case e := <-li.Events():
					l.broadcast(e)
				}
			}
		}()
	}
}

func (l *EventPublisher) broadcast(event *messages.Event) {
	if !l.sign(event) {
		return
	}
	l.log.
		WithField("id", hex.EncodeToString(event.ID)).
		WithField("type", event.Type).
		WithField("index", hex.EncodeToString(event.Index)).
		Info("Event broadcast")
	err := l.transport.Broadcast(messages.EventMessageName, event)
	if err != nil {
		l.log.
			WithError(err).
			WithField("id", hex.EncodeToString(event.ID)).
			WithField("type", event.Type).
			WithField("index", hex.EncodeToString(event.Index)).
			Error("Unable to broadcast the event")
	}
}

func (l *EventPublisher) sign(event *messages.Event) bool {
	var signed bool
	for _, s := range l.signers {
		ok, err := s.Sign(event)
		if !ok {
			continue
		}
		if err != nil {
			l.log.
				WithError(err).
				WithField("id", hex.EncodeToString(event.ID)).
				WithField("type", event.Type).
				WithField("index", hex.EncodeToString(event.Index)).
				Error("Unable to sign the event")
			continue
		}
		signed = true
	}
	if !signed {
		l.log.
			WithField("id", hex.EncodeToString(event.ID)).
			WithField("type", event.Type).
			WithField("index", hex.EncodeToString(event.Index)).
			Warn("There are no signers that supports the event")
	}
	return signed
}

func (l *EventPublisher) contextCancelHandler() {
	defer func() { l.waitCh <- nil }()
	defer l.log.Info("Stopped")
	<-l.ctx.Done()
}
