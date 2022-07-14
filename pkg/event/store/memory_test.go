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
	"strconv"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"

	"github.com/chronicleprotocol/oracle-suite/pkg/transport/messages"
)

func TestMemory_Add(t *testing.T) {
	m := NewMemoryStorage(time.Minute)
	e1 := &messages.Event{
		Type:        "test",
		ID:          []byte("test"),
		Index:       []byte("idx"),
		MessageDate: time.Now(),
		EventDate:   time.Now(),
		Data:        map[string][]byte{"test": []byte("test")},
		Signatures:  map[string]messages.EventSignature{},
	}
	e2 := &messages.Event{
		Type:        "test",
		ID:          []byte("test2"),
		Index:       []byte("idx"),
		MessageDate: time.Now(),
		EventDate:   time.Now(),
		Data:        map[string][]byte{"test": []byte("test2")},
		Signatures:  map[string]messages.EventSignature{},
	}
	e3 := &messages.Event{
		Type:        "test",
		ID:          []byte("test2"),
		Index:       []byte("idx2"),
		MessageDate: time.Now(),
		EventDate:   time.Now(),
		Data:        map[string][]byte{"test": []byte("test2")},
		Signatures:  map[string]messages.EventSignature{},
	}

	isNew, err := m.Add(context.Background(), []byte("author1"), e1)
	assert.True(t, isNew)
	assert.NoError(t, err)

	isNew, err = m.Add(context.Background(), []byte("author2"), e2)
	assert.True(t, isNew)
	assert.NoError(t, err)

	isNew, err = m.Add(context.Background(), []byte("author3"), e3)
	assert.True(t, isNew)
	assert.NoError(t, err)

	es, err := m.Get(context.Background(), "test", []byte("idx"))
	assert.NoError(t, err)
	assert.ElementsMatch(t, []*messages.Event{e1, e2}, es)
}

func TestMemory_Add_replacePreviousEvent(t *testing.T) {
	m := NewMemoryStorage(time.Minute)
	e1 := &messages.Event{
		Type:        "test",
		ID:          []byte("test"),
		Index:       []byte("idx"),
		MessageDate: time.Unix(1, 0),
		EventDate:   time.Unix(1, 0),
		Data:        map[string][]byte{"test": []byte("test")},
		Signatures:  map[string]messages.EventSignature{},
	}
	e2 := &messages.Event{
		Type:        "test",
		ID:          []byte("test"),
		Index:       []byte("idx"),
		MessageDate: time.Unix(2, 0),
		EventDate:   time.Unix(2, 0),
		Data:        map[string][]byte{"test": []byte("test2")},
		Signatures:  map[string]messages.EventSignature{},
	}

	isNew, err := m.Add(context.Background(), []byte("author1"), e1)
	assert.True(t, isNew)
	assert.NoError(t, err)

	// Replace if never
	isNew, err = m.Add(context.Background(), []byte("author1"), e2)
	assert.False(t, isNew)
	assert.NoError(t, err)

	es, err := m.Get(context.Background(), "test", []byte("idx"))
	assert.NoError(t, err)
	assert.ElementsMatch(t, []*messages.Event{e2}, es)

	// Keep previous if older
	isNew, err = m.Add(context.Background(), []byte("author1"), e1)
	assert.False(t, isNew)
	assert.NoError(t, err)

	es, err = m.Get(context.Background(), "test", []byte("idx"))
	assert.NoError(t, err)
	assert.ElementsMatch(t, []*messages.Event{e2}, es)
}

func TestMemory_gc(t *testing.T) {
	m := NewMemoryStorage(time.Minute)
	_, err := m.Add(context.Background(), []byte("author"), &messages.Event{
		Type:        "test",
		ID:          []byte("test"),
		Index:       []byte("idx"),
		MessageDate: time.Now(),
		EventDate:   time.Now(),
		Data:        map[string][]byte{"test": []byte("test")},
		Signatures:  map[string]messages.EventSignature{},
	})
	assert.NoError(t, err)
	for i := 0; i < m.gcevery-1; i++ {
		e := &messages.Event{
			Type:        "test",
			ID:          []byte(strconv.Itoa(i)),
			Index:       []byte("idx"),
			MessageDate: time.Unix(0, 0),
			EventDate:   time.Unix(0, 0),
			Data:        map[string][]byte{"test": []byte("test")},
			Signatures:  map[string]messages.EventSignature{},
		}
		_, err := m.Add(context.Background(), []byte(strconv.Itoa(i)), e)
		assert.NoError(t, err)
	}

	es, err := m.Get(context.Background(), "test", []byte("idx"))
	assert.NoError(t, err)
	assert.Len(t, es, 1)
}

func TestMemory_gc_allExpired(t *testing.T) {
	m := NewMemoryStorage(time.Minute)
	for i := 0; i < m.gcevery; i++ {
		e := &messages.Event{
			Type:        "test",
			ID:          []byte(strconv.Itoa(i)),
			Index:       []byte("idx"),
			MessageDate: time.Unix(0, 0),
			EventDate:   time.Unix(0, 0),
			Data:        map[string][]byte{"test": []byte("test")},
			Signatures:  map[string]messages.EventSignature{},
		}
		_, err := m.Add(context.Background(), []byte(strconv.Itoa(i)), e)
		assert.NoError(t, err)
	}

	es, err := m.Get(context.Background(), "test", []byte("idx"))
	assert.NoError(t, err)
	assert.Len(t, es, 0)
}
