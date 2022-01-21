package memory

import (
	"crypto/sha256"
	"sync"
	"time"

	"github.com/chronicleprotocol/oracle-suite/pkg/transport/messages"
)

type Memory struct {
	mu sync.Mutex

	ttl    time.Duration // Message TTL.
	events map[[sha256.Size]byte][]*messages.Event

	// Variables used for message garbage collector.
	gccount int // Increases every time a message is added.
	gcevery int // Specifies every how many messages the garbage collector should be called.
}

// New returns a new instance of Memory. The ttl argument specifies how long
// the message should be kept in storage.
func New(ttl time.Duration) *Memory {
	return &Memory{
		ttl:     ttl,
		events:  map[[sha256.Size]byte][]*messages.Event{},
		gcevery: 100,
	}
}

// Add implements the store.Storage interface.
func (m *Memory) Add(msg *messages.Event) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	h := hash(msg.Type, msg.Index)
	if _, ok := m.events[h]; !ok {
		m.events[h] = nil
	}
	m.events[h] = append(m.events[h], msg)
	m.gc()
	return nil
}

// Get implements the store.Storage interface.
func (m *Memory) Get(typ string, idx []byte) ([]*messages.Event, error) {
	m.mu.Lock()
	defer m.mu.Unlock()
	return m.events[hash(typ, idx)], nil
}

// Garbage Collector removes expired messages.
func (m *Memory) gc() {
	m.gccount++
	if m.gccount%m.gcevery != 0 {
		return
	}
	for h, events := range m.events {
		// Count number of expired messages:
		expired := 0
		for _, event := range events {
			if time.Since(event.Date) > m.ttl {
				expired++
			}
		}
		// Delete expired messages:
		if expired == len(m.events[h]) {
			// If all messages with the same hash are expired.
			delete(m.events, h)
		} else if expired > 0 {
			// If only some messages are expired.
			var es []*messages.Event
			for _, event := range events {
				if time.Since(event.Date) <= m.ttl {
					es = append(es, event)
				}
			}
			m.events[h] = es
		}
	}
}

func hash(typ string, index []byte) [sha256.Size]byte {
	return sha256.Sum256(append([]byte(typ), index...))
}
