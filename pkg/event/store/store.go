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
	"encoding/hex"
	"errors"

	"github.com/chronicleprotocol/oracle-suite/pkg/log"
	"github.com/chronicleprotocol/oracle-suite/pkg/log/null"
	"github.com/chronicleprotocol/oracle-suite/pkg/transport"
	"github.com/chronicleprotocol/oracle-suite/pkg/transport/messages"
)

const LoggerTag = "EVENT_STORE"

// EventStore listens for event messages using the transport and stores
// them for later use.
type EventStore struct {
	ctx        context.Context
	eventTypes []string
	storage    Storage
	transport  transport.Transport
	log        log.Logger
	waitCh     chan error
}

// Config is the configuration for the EventStore.
type Config struct {
	// EventTypes is a list of supported event types. Events of other types will
	// be ignored.
	EventTypes []string
	// Storage is the storage implementation.
	Storage Storage
	// Transport is a transport interface used to fetch events from Oracles.
	Transport transport.Transport
	// Logger is a current logger interface used by the EventStore.
	// The Logger is required to monitor asynchronous processes.
	Logger log.Logger
}

// Storage provides an interface to the event storage.
type Storage interface {
	// Add adds an event to the store. If the event already exists, it will be
	// updated if the MessageDate is newer. The first argument is true if the
	// event was added, false if it was replaced. The method is thread-safe.
	Add(ctx context.Context, author []byte, evt *messages.Event) (bool, error)
	// Get returns messages form the store for the given type and index. If the
	// message does not exist, nil will be returned. The method is thread-safe.
	Get(ctx context.Context, typ string, idx []byte) ([]*messages.Event, error)
}

// New returns a new instance of the EventStore struct.
func New(cfg Config) (*EventStore, error) {
	if cfg.Storage == nil {
		return nil, errors.New("storage must not be nil")
	}
	if cfg.Transport == nil {
		return nil, errors.New("transport must not be nil")
	}
	if cfg.Logger == nil {
		cfg.Logger = null.New()
	}
	return &EventStore{
		eventTypes: cfg.EventTypes,
		storage:    cfg.Storage,
		transport:  cfg.Transport,
		log:        cfg.Logger.WithField("tag", LoggerTag),
		waitCh:     make(chan error),
	}, nil
}

func (e *EventStore) Start(ctx context.Context) error {
	if e.ctx != nil {
		return errors.New("service can be started only once")
	}
	if ctx == nil {
		return errors.New("context must not be nil")
	}
	e.log.Info("Starting")
	e.ctx = ctx
	go e.eventCollectorRoutine()
	go e.contextCancelHandler()
	return nil
}

// Wait waits until the context is canceled or until an error occurs.
func (e *EventStore) Wait() chan error {
	return e.waitCh
}

// Events returns events for the given type and index. The method is thread-safe.
func (e *EventStore) Events(ctx context.Context, typ string, idx []byte) ([]*messages.Event, error) {
	return e.storage.Get(ctx, typ, idx)
}

func (e *EventStore) eventCollectorRoutine() {
	for {
		select {
		case <-e.ctx.Done():
			return
		case msg := <-e.transport.Messages(messages.EventV1MessageName):
			if msg.Error != nil {
				e.log.WithError(msg.Error).Error("Unable to read events from the transport layer")
				continue
			}
			evt, ok := msg.Message.(*messages.Event)
			if !ok {
				e.log.Error("Unexpected value returned from the transport layer")
				continue
			}
			if !e.isEventSupported(evt) {
				continue
			}
			isNew, err := e.storage.Add(e.ctx, msg.Author, evt)
			e.log.
				WithFields(log.Fields{
					"id":          hex.EncodeToString(evt.ID),
					"type":        evt.Type,
					"index":       hex.EncodeToString(evt.Index),
					"eventDate":   evt.EventDate,
					"messageDate": evt.MessageDate,
					"data":        evt.Data,
					"signatures":  evt.Signatures,
					"from":        msg.Author,
					"new":         isNew,
				}).
				Info("Event received")
			if err != nil {
				e.log.WithError(msg.Error).Error("Unable to store the event")
				continue
			}
		}
	}
}

func (e *EventStore) isEventSupported(evt *messages.Event) bool {
	for _, typ := range e.eventTypes {
		if typ == evt.Type {
			return true
		}
	}
	return false
}

// contextCancelHandler handles context cancellation.
func (e *EventStore) contextCancelHandler() {
	defer func() { close(e.waitCh) }()
	defer e.log.Info("Stopped")
	<-e.ctx.Done()
}
