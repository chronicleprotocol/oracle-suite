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
	"github.com/stretchr/testify/require"

	"github.com/chronicleprotocol/oracle-suite/pkg/transport/messages"
)

func TestMain(m *testing.M) {
	rand.Seed(time.Now().Unix())
	os.Exit(m.Run())
}

func TestRedis_Add(t *testing.T) {
	ok, cfg := getConfig()
	if !ok {
		t.Skip()
		return
	}
	typ := strconv.Itoa(rand.Int())
	author := strconv.Itoa(rand.Int())
	r, err := NewRedisStorage(cfg)
	require.NoError(t, err)
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

	isNew, err := r.Add(context.Background(), []byte(author+"1"), e1)
	assert.True(t, isNew)
	assert.NoError(t, err)

	isNew, err = r.Add(context.Background(), []byte(author+"2"), e2)
	assert.True(t, isNew)
	assert.NoError(t, err)

	isNew, err = r.Add(context.Background(), []byte(author+"3"), e3) // different index
	assert.True(t, isNew)
	assert.NoError(t, err)

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
	author := strconv.Itoa(rand.Int())
	r, err := NewRedisStorage(cfg)
	require.NoError(t, err)
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
		MessageDate: time.Now().Add(time.Second * 1),
		EventDate:   time.Now().Add(time.Second * 1),
		Data:        map[string][]byte{"test": []byte("test2")},
		Signatures:  map[string]messages.EventSignature{},
	}

	isNew, err := r.Add(context.Background(), []byte(author), e1)
	assert.True(t, isNew)
	assert.NoError(t, err)

	// Replace if never
	isNew, err = r.Add(context.Background(), []byte(author), e2)
	assert.False(t, isNew)
	assert.NoError(t, err)

	es, err := r.Get(context.Background(), typ, []byte("idx"))
	assert.NoError(t, err)
	assert.ElementsMatch(t, eventsToByteSlices([]*messages.Event{e2}), eventsToByteSlices(es))

	// Keep previous if older
	isNew, err = r.Add(context.Background(), []byte(author), e1)
	assert.False(t, isNew)
	assert.NoError(t, err)
	es, err = r.Get(context.Background(), typ, []byte("idx"))
	assert.NoError(t, err)
	assert.ElementsMatch(t, eventsToByteSlices([]*messages.Event{e2}), eventsToByteSlices(es))
}

func TestRedis_memoryLimit(t *testing.T) {
	ok, cfg := getConfig()
	cfg.MemoryLimit = 60 // 60 is enough for one message
	if !ok {
		t.Skip()
		return
	}
	typ := strconv.Itoa(rand.Int())
	author := strconv.Itoa(rand.Int())
	r, err := NewRedisStorage(cfg)
	require.NoError(t, err)

	e1 := &messages.Event{
		Type:        typ,
		ID:          []byte("test"),
		Index:       []byte("idx1"),
		MessageDate: time.Now(),
		EventDate:   time.Now(),
		Data:        map[string][]byte{"test": []byte("test")},
		Signatures:  map[string]messages.EventSignature{},
	}
	e2 := &messages.Event{
		Type:        typ,
		ID:          []byte("test"),
		Index:       []byte("idx2"),
		MessageDate: time.Now(),
		EventDate:   time.Now(),
		Data:        map[string][]byte{"test": []byte("test")},
		Signatures:  map[string]messages.EventSignature{},
	}

	isNew, err := r.Add(context.Background(), []byte(author), e1)
	assert.True(t, isNew)
	assert.NoError(t, err)

	isNew, err = r.Add(context.Background(), []byte(author), e2)
	assert.False(t, isNew)
	assert.Error(t, err)
}

func getConfig() (bool, Config) {
	addr := os.Getenv("TEST_REDIS_ADDR")
	pass := os.Getenv("TEST_REDIS_PASS")
	db, _ := strconv.Atoi(os.Getenv("TEST_REDIS_DB"))
	if len(addr) == 0 {
		return false, Config{}
	}
	return true, Config{
		MemoryLimit: 0,
		TTL:         time.Minute,
		Address:     addr,
		Password:    pass,
		DB:          db,
	}
}

func eventsToByteSlices(es []*messages.Event) (r [][]byte) {
	for _, e := range es {
		b, _ := e.MarshallBinary()
		r = append(r, b)
	}
	return
}
