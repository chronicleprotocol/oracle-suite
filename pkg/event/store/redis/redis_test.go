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

package redis

import (
	"context"
	"math/rand"
	"os"
	"strconv"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"

	"github.com/chronicleprotocol/oracle-suite/pkg/transport/messages"
)

func TestRedis_Add(t *testing.T) {
	ok, cfg := getConfig()
	if !ok {
		t.Skip()
		return
	}
	typ := strconv.Itoa(rand.Int())
	r := New(cfg)
	e1 := &messages.Event{
		Type:        typ,
		ID:          []byte("test"),
		Index:       []byte("idx"),
		MessageDate: time.Now(),
		EventDate:   time.Now(),
		Data:        map[string][]byte{"test": []byte("test")},
		Signatures:  map[string]messages.EventSignature{},
	}
	e2 := &messages.Event{
		Type:        typ,
		ID:          []byte("test2"),
		Index:       []byte("idx"),
		MessageDate: time.Now(),
		EventDate:   time.Now(),
		Data:        map[string][]byte{"test": []byte("test2")},
		Signatures:  map[string]messages.EventSignature{},
	}
	e3 := &messages.Event{
		Type:        typ,
		ID:          []byte("test2"),
		Index:       []byte("idx2"),
		MessageDate: time.Now(),
		EventDate:   time.Now(),
		Data:        map[string][]byte{"test": []byte("test2")},
		Signatures:  map[string]messages.EventSignature{},
	}

	assert.NoError(t, r.Add(context.Background(), []byte("author1"), e1))
	assert.NoError(t, r.Add(context.Background(), []byte("author2"), e2))
	assert.NoError(t, r.Add(context.Background(), []byte("author3"), e3)) // different index

	es, err := r.Get(context.Background(), typ, []byte("idx"))
	assert.NoError(t, err)
	assert.ElementsMatch(t, eventsToByteSlices([]*messages.Event{e1, e2}), eventsToByteSlices(es))
}

func TestRedis_Add_replacePreviousEvent(t *testing.T) {
	ok, cfg := getConfig()
	if !ok {
		t.Skip()
		return
	}
	typ := strconv.Itoa(rand.Int())
	r := New(cfg)
	e1 := &messages.Event{
		Type:        typ,
		ID:          []byte("test"),
		Index:       []byte("idx"),
		MessageDate: time.Now(),
		EventDate:   time.Now(),
		Data:        map[string][]byte{"test": []byte("test")},
		Signatures:  map[string]messages.EventSignature{},
	}
	e2 := &messages.Event{
		Type:        typ,
		ID:          []byte("test"),
		Index:       []byte("idx"),
		MessageDate: time.Now().Add(time.Second * 5),
		EventDate:   time.Now().Add(time.Second * 5),
		Data:        map[string][]byte{"test": []byte("test2")},
		Signatures:  map[string]messages.EventSignature{},
	}

	assert.NoError(t, r.Add(context.Background(), []byte("author1"), e1))

	// Replace if never
	assert.NoError(t, r.Add(context.Background(), []byte("author1"), e2))
	es, err := r.Get(context.Background(), typ, []byte("idx"))
	assert.NoError(t, err)
	assert.ElementsMatch(t, eventsToByteSlices([]*messages.Event{e2}), eventsToByteSlices(es))

	// Keep previous if older
	assert.NoError(t, r.Add(context.Background(), []byte("author1"), e1))
	es, err = r.Get(context.Background(), typ, []byte("idx"))
	assert.NoError(t, err)
	assert.ElementsMatch(t, eventsToByteSlices([]*messages.Event{e2}), eventsToByteSlices(es))
}

func getConfig() (bool, Config) {
	addr := os.Getenv("TEST_REDIS_ADDR")
	pass := os.Getenv("TEST_REDIS_PASS")
	db, _ := strconv.Atoi(os.Getenv("TEST_REDIS_DB"))
	if len(addr) == 0 {
		return false, Config{}
	}
	return true, Config{
		TTL:      time.Minute,
		Address:  addr,
		Password: pass,
		DB:       db,
	}
}

func eventsToByteSlices(es []*messages.Event) (r [][]byte) {
	for _, e := range es {
		b, _ := e.MarshallBinary()
		r = append(r, b)
	}
	return
}
