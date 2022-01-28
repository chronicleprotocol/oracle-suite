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

package store

import (
	"context"
	"errors"

	"github.com/chronicleprotocol/oracle-suite/pkg/log"
	"github.com/chronicleprotocol/oracle-suite/pkg/transport"
	"github.com/chronicleprotocol/oracle-suite/pkg/transport/messages"
)

const LoggerTag = "EVENTSTORE"

// EventStore listens for event messages using the transport and stores
// them for later use.
type EventStore struct {
	ctx       context.Context
	storage   Storage
	transport transport.Transport
	log       log.Logger
	waitCh    chan error
}

// Config contains configuration parameters for EventStore.
type Config struct {
	Storage   Storage
	Transport transport.Transport
	Log       log.Logger
}

type Storage interface {
	// Add adds a message to the store. If the message already exists, no error
	// will be returned. Method is thread safe.
	Add(author []byte, msg *messages.Event) error
	// Get returns a message form the store. If the message does not exist,
	// nil will be returned. Method is thread safe.
	Get(typ string, idx []byte) ([]*messages.Event, error)
}

// New returns a new instance of the EventStore struct.
func New(ctx context.Context, cfg Config) (*EventStore, error) {
	if ctx == nil {
		return nil, errors.New("context must not be nil")
	}
	return &EventStore{
		ctx:       ctx,
		storage:   cfg.Storage,
		transport: cfg.Transport,
		log:       cfg.Log.WithField("tag", LoggerTag),
		waitCh:    make(chan error),
	}, nil
}

func (e *EventStore) Start() error {
	e.log.Info("Starting")
	go e.eventCollectorRoutine()
	go e.contextCancelHandler()
	return nil
}

// Wait waits until the context is canceled or until an error occurs.
func (e *EventStore) Wait() chan error {
	return e.waitCh
}

func (e *EventStore) Events(typ string, idx []byte) ([]*messages.Event, error) {
	return e.storage.Get(typ, idx)
}

func (e *EventStore) eventCollectorRoutine() {
	for {
		select {
		case <-e.ctx.Done():
			return
		case msg := <-e.transport.Messages(messages.EventMessageName):
			if msg.Error != nil {
				e.log.
					WithError(msg.Error).
					Warn("Unable to read events from the transport layer")
				continue
			}
			evtMsg, ok := msg.Message.(*messages.Event)
			if !ok {
				e.log.Error("Unexpected value returned from the transport layer")
				continue
			}
			err := e.storage.Add(msg.Author, evtMsg)
			if err != nil {
				e.log.
					WithError(msg.Error).
					Warn("Unable to store the event")
				continue
			}
		}
	}
}

// contextCancelHandler handles context cancellation.
func (e *EventStore) contextCancelHandler() {
	defer func() { e.waitCh <- nil }()
	defer e.log.Info("Stopped")
	<-e.ctx.Done()
}
