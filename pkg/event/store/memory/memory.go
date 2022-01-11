package memory

import (
	"context"
	"crypto/md5"
	"errors"
	"sync"
	"time"

	"github.com/chronicleprotocol/oracle-suite/pkg/log"
	"github.com/chronicleprotocol/oracle-suite/pkg/transport"
	"github.com/chronicleprotocol/oracle-suite/pkg/transport/messages"
)

const LoggerTag = "EVENTSTORE"

type EventStore struct {
	ctx    context.Context
	mu     sync.Mutex
	doneCh chan struct{}

	events    map[[md5.Size]byte][]*messages.Event
	transport transport.Transport
	log       log.Logger
}

type Config struct {
	Transport transport.Transport
	Logger    log.Logger
}

func NewEventStore(ctx context.Context, cfg Config) (*EventStore, error) {
	if ctx == nil {
		return nil, errors.New("context must not be nil")
	}
	return &EventStore{
		ctx:       ctx,
		events:    map[[16]byte][]*messages.Event{},
		transport: cfg.Transport,
		log:       cfg.Logger.WithField("tag", LoggerTag),
	}, nil
}

func (m *EventStore) Start() error {
	m.log.Info("Starting")
	go m.collectorLoop()
	go m.contextCancelHandler()
	go m.cleanupOldEvents()
	return nil
}

func (m *EventStore) Wait() error {
	<-m.doneCh
	return nil
}

func (m *EventStore) Events(typ string, index []byte) []*messages.Event {
	return m.events[mapHash(typ, index)]
}

func (m *EventStore) collectorLoop() {
	for {
		select {
		case <-m.ctx.Done():
			return
		case msg := <-m.transport.Messages(messages.EventMessageName):
			if msg.Error != nil {
				m.log.
					WithError(msg.Error).
					Warn("Unable to read events from the transport")
				continue
			}
			event, ok := msg.Message.(*messages.Event)
			if !ok {
				m.log.Error("Unexpected value returned from transport layer")
				continue
			}
			h := mapHash(event.Type, event.Index)
			m.events[h] = append(m.events[h], event)
		}
	}
}

func (m *EventStore) cleanupOldEvents() {
	t := time.NewTicker(120 * time.Second)
	defer t.Stop()
	for {
		select {
		case <-m.ctx.Done():
			return
		case <-t.C:
			for h, events := range m.events {
				s := m.events[h]
				for n, event := range events {
					if time.Since(event.Date) > 7*24*time.Hour {
						s = removeMessage(s, n)
					}
				}
				if len(s) == 0 {
					delete(m.events, h)
				}
			}
		}
	}
}

func (m *EventStore) contextCancelHandler() {
	defer func() { close(m.doneCh) }()
	defer m.log.Info("Stopped")
	<-m.ctx.Done()
}

func removeMessage(s []*messages.Event, n int) []*messages.Event {
	return append(s[:n], s[n+1:]...)
}

func mapHash(typ string, index []byte) [md5.Size]byte {
	return md5.Sum(append([]byte(typ), index...))
}
