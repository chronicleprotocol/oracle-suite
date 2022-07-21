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
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/chronicleprotocol/oracle-suite/pkg/log/null"
	"github.com/chronicleprotocol/oracle-suite/pkg/transport"
	"github.com/chronicleprotocol/oracle-suite/pkg/transport/local"
	"github.com/chronicleprotocol/oracle-suite/pkg/transport/messages"
)

func TestEventStore(t *testing.T) {
	ctx, cancelFunc := context.WithCancel(context.Background())
	tra := local.New([]byte("test"), 1, map[string]transport.Message{messages.EventV1MessageName: (*messages.Event)(nil)})

	mem := NewMemoryStorage(time.Minute)
	evs, err := New(Config{
		EventTypes: []string{"test"},
		Storage:    mem,
		Transport:  tra,
		Logger:     null.New(),
	})
	require.NoError(t, err)

	require.NoError(t, tra.Start(ctx))
	require.NoError(t, evs.Start(ctx))
	defer func() {
		cancelFunc()
		require.NoError(t, <-evs.Wait())
		require.NoError(t, <-tra.Wait())
	}()

	event := &messages.Event{
		Type:        "test",
		ID:          []byte("test"),
		Index:       []byte("idx"),
		EventDate:   time.Now(),
		MessageDate: time.Now(),
		Data:        map[string][]byte{"test": []byte("test")},
		Signatures:  map[string]messages.EventSignature{"sig_key": {Signer: []byte("val"), Signature: []byte("val")}},
	}
	require.NoError(t, tra.Broadcast(messages.EventV1MessageName, event))

	time.Sleep(100 * time.Millisecond)

	events, err := evs.Events(context.Background(), "test", []byte("idx"))
	require.NoError(t, err)

	require.Len(t, events, 1)
	assert.Equal(t, event.Type, events[0].Type)
	assert.Equal(t, event.ID, events[0].ID)
	assert.Equal(t, event.Index, events[0].Index)
	assert.Equal(t, event.EventDate.Unix(), events[0].EventDate.Unix())
	assert.Equal(t, event.MessageDate.Unix(), events[0].MessageDate.Unix())
	assert.Equal(t, event.Data, events[0].Data)
	assert.Equal(t, event.Signatures, events[0].Signatures)
}
