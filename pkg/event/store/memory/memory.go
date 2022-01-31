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

package memory

import (
	"crypto/sha256"
	"sync"
	"time"

	"github.com/chronicleprotocol/oracle-suite/pkg/transport/messages"
)

type Memory struct {
	mu sync.RWMutex

	ttl   time.Duration // Message TTL.
	index map[[sha256.Size]byte]map[[sha256.Size]byte]*messages.Event

	// Variables used for message garbage collector.
	gccount int // Increases every time a message is added.
	gcevery int // Specifies every how many messages the garbage collector should be called.
}

type evtPtr struct {
	ptr *messages.Event
}

// New returns a new instance of Memory. The ttl argument specifies how long
// the message should be kept in storage.
func New(ttl time.Duration) *Memory {
	return &Memory{
		ttl:     ttl,
		index:   map[[sha256.Size]byte]map[[sha256.Size]byte]*messages.Event{},
		gcevery: 100,
	}
}

// Add implements the store.Storage interface.
func (m *Memory) Add(author []byte, msg *messages.Event) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	hi := hashIndex(msg.Type, msg.Index)
	hu := hashUnique(author, msg.ID)
	if _, ok := m.index[hi]; !ok {
		m.index[hi] = map[[32]byte]*messages.Event{}
	}
	evt, ok := m.index[hi][hu]
	if !ok || (ok && evt.Date.Before(msg.Date)) {
		m.index[hi][hu] = msg
		m.gc()
	}
	return nil
}

// Get implements the store.Storage interface.
func (m *Memory) Get(typ string, idx []byte) ([]*messages.Event, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()
	hi := hashIndex(typ, idx)
	if _, ok := m.index[hi]; ok {
		var evts []*messages.Event
		for _, evt := range m.index[hi] {
			evts = append(evts, evt)
		}
		return evts, nil
	}
	return nil, nil
}

// Garbage Collector removes expired messages.
func (m *Memory) gc() {
	m.gccount++
	if m.gccount%m.gcevery != 0 {
		return
	}
	for hi, evts := range m.index {
		// Count number of expired messages:
		expired := 0
		for _, evt := range evts {
			if time.Since(evt.Date) > m.ttl {
				expired++
			}
		}
		// Delete expired messages:
		if expired == len(m.index[hi]) {
			// If all messages with the same hash are expired.
			delete(m.index, hi)
		} else if expired > 0 {
			// If only some messages are expired.
			for ha, evt := range evts {
				if time.Since(evt.Date) >= m.ttl {
					delete(m.index[hi], ha)
				}
			}
		}
	}
}

func hashUnique(author []byte, id []byte) [sha256.Size]byte {
	return sha256.Sum256(append(author, id...))
}

func hashIndex(typ string, index []byte) [sha256.Size]byte {
	return sha256.Sum256(append([]byte(typ), index...))
}
