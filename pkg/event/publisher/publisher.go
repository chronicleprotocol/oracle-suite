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
	"errors"

	"github.com/chronicleprotocol/oracle-suite/pkg/log"
	"github.com/chronicleprotocol/oracle-suite/pkg/log/null"
	"github.com/chronicleprotocol/oracle-suite/pkg/transport"
	"github.com/chronicleprotocol/oracle-suite/pkg/transport/messages"
)

const LoggerTag = "EVENT_PUBLISHER"

// EventPublisher collects event messages from event providers and publishes
// them using transport interface.
//
// An event message could be anything Oracle could sign.
type EventPublisher struct {
	ctx    context.Context
	waitCh chan error

	signers   []Signer
	listeners []EventProvider
	transport transport.Transport
	log       log.Logger
}

// Config is the configuration for the EventPublisher.
type Config struct {
	Listeners []EventProvider
	// Signer is a list of Signers used to sign events.
	Signers []Signer
	// Transport is used to send events to the Oracle network.
	Transport transport.Transport
	// Logger is a current logger interface used by the EventPublisher. The Logger
	// helps to monitor asynchronous processes.
	Logger log.Logger
}

// EventProvider providers events to EventPublisher.
type EventProvider interface {
	Start(ctx context.Context) error
	Events() chan *messages.Event
}

type Signer interface {
	Sign(event *messages.Event) (bool, error)
}

// New returns a new instance of the EventPublisher struct.
func New(cfg Config) (*EventPublisher, error) {
	if cfg.Transport == nil {
		return nil, errors.New("transport must not be nil")
	}
	if cfg.Logger == nil {
		cfg.Logger = null.New()
	}
	return &EventPublisher{
		waitCh:    make(chan error),
		transport: cfg.Transport,
		listeners: cfg.Listeners,
		signers:   cfg.Signers,
		log:       cfg.Logger.WithField("tag", LoggerTag),
	}, nil
}

// Start implements the supervisor.Service interface.
func (l *EventPublisher) Start(ctx context.Context) error {
	if l.ctx != nil {
		return errors.New("service can be started only once")
	}
	if ctx == nil {
		return errors.New("context must not be nil")
	}
	l.log.Infof("Starting")
	l.ctx = ctx
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

// Wait implements the supervisor.Service interface.
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

func (l *EventPublisher) broadcast(evt *messages.Event) {
	if !l.sign(evt) {
		return
	}
	l.log.
		WithFields(log.Fields{
			"id":          evt.ID,
			"type":        evt.Type,
			"index":       evt.Index,
			"eventDate":   evt.EventDate,
			"messageDate": evt.MessageDate,
			"data":        evt.Data,
			"signatures":  evt.Signatures,
			"from":        l.transport.ID(),
		}).
		Info("Event published")
	err := l.transport.Broadcast(messages.EventV1MessageName, evt)
	if err != nil {
		l.log.
			WithError(err).
			WithFields(log.Fields{
				"id":   evt.ID,
				"type": evt.Type,
				"from": l.transport.ID(),
			}).
			Error("Unable to publish the event")
	}
}

func (l *EventPublisher) sign(evt *messages.Event) bool {
	var signed bool
	for _, s := range l.signers {
		ok, err := s.Sign(evt)
		if !ok {
			continue
		}
		if err != nil {
			l.log.
				WithError(err).
				WithFields(log.Fields{
					"id":   evt.ID,
					"type": evt.Type,
					"from": l.transport.ID(),
				}).
				Error("Unable to sign the event")
			continue
		}
		signed = true
	}
	if !signed {
		l.log.
			WithFields(log.Fields{
				"id":   evt.ID,
				"type": evt.Type,
				"from": l.transport.ID(),
			}).
			Warn("There are no signers that supports the event")
	}
	return signed
}

func (l *EventPublisher) contextCancelHandler() {
	defer func() { close(l.waitCh) }()
	defer l.log.Info("Stopped")
	<-l.ctx.Done()
}
