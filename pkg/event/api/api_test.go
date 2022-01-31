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

package api

import (
	"context"
	"fmt"
	"io"
	"net/http"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/chronicleprotocol/oracle-suite/pkg/event/store"
	"github.com/chronicleprotocol/oracle-suite/pkg/event/store/memory"
	"github.com/chronicleprotocol/oracle-suite/pkg/log/null"
	"github.com/chronicleprotocol/oracle-suite/pkg/transport"
	"github.com/chronicleprotocol/oracle-suite/pkg/transport/local"
	"github.com/chronicleprotocol/oracle-suite/pkg/transport/messages"
)

func TestEventAPI(t *testing.T) {
	ctx, cancelFunc := context.WithCancel(context.Background())
	loc := local.New(ctx, []byte("test"), 4, map[string]transport.Message{messages.EventMessageName: (*messages.Event)(nil)})
	mem := memory.New(time.Minute)
	evs, err := store.New(ctx, store.Config{
		Storage:   mem,
		Transport: loc,
		Log:       null.New(),
	})
	require.NoError(t, err)
	api, err := New(ctx, Config{
		EventStore: evs,
		Address:    "127.0.0.1:0",
		Logger:     null.New(),
	})
	require.NoError(t, err)

	require.NoError(t, loc.Start())
	require.NoError(t, evs.Start())
	require.NoError(t, api.Start())
	defer func() {
		cancelFunc()
		require.NoError(t, <-loc.Wait())
		require.NoError(t, <-evs.Wait())
		require.NoError(t, <-api.Wait())
	}()

	require.NoError(t, loc.Broadcast(messages.EventMessageName, &messages.Event{
		Type:       "event1",
		ID:         []byte("id1"),
		Index:      []byte("idx1"),
		Date:       time.Unix(1, 0),
		Data:       map[string][]byte{"data_key": []byte("val")},
		Signatures: map[string]messages.EventSignature{"sig_key": {Signer: []byte("val"), Signature: []byte("val")}},
	}))
	require.NoError(t, loc.Broadcast(messages.EventMessageName, &messages.Event{
		Type:       "event1",
		ID:         []byte("id2"),
		Index:      []byte("idx1"),
		Date:       time.Unix(2, 0),
		Data:       map[string][]byte{"data_key": []byte("val")},
		Signatures: map[string]messages.EventSignature{"sig_key": {Signer: []byte("val"), Signature: []byte("val")}},
	}))
	require.NoError(t, loc.Broadcast(messages.EventMessageName, &messages.Event{
		Type:       "event1",
		ID:         []byte("id3"),
		Index:      []byte("idx2"), // different index
		Date:       time.Unix(3, 0),
		Data:       map[string][]byte{"data_key": []byte("val")},
		Signatures: map[string]messages.EventSignature{"sig_key": {Signer: []byte("val"), Signature: []byte("val")}},
	}))
	require.NoError(t, loc.Broadcast(messages.EventMessageName, &messages.Event{
		Type:       "event2", // different type
		ID:         []byte("id4"),
		Index:      []byte("idx1"),
		Date:       time.Unix(4, 0),
		Data:       map[string][]byte{"data_key": []byte("val")},
		Signatures: map[string]messages.EventSignature{"sig_key": {Signer: []byte("val"), Signature: []byte("val")}},
	}))

	time.Sleep(time.Second)

	res, err := http.Get(fmt.Sprintf("http://%s?type=event1&index=%x", api.srv.Addr().String(), "idx1"))
	assert.NoError(t, err)
	assert.JSONEq(t, `[{"timestamp":1,"id":"696431","data":{"data_key":"76616c"},"signatures":{"sig_key":{"signer":"76616c","signature":"76616c"}}},{"timestamp":2,"id":"696432","data":{"data_key":"76616c"},"signatures":{"sig_key":{"signer":"76616c","signature":"76616c"}}}]`, read(res))

	res, err = http.Get(fmt.Sprintf("http://%s?type=event1&index=0x%x", api.srv.Addr().String(), "idx2"))
	assert.NoError(t, err)
	assert.JSONEq(t, `[{"timestamp":3,"id":"696433","data":{"data_key":"76616c"},"signatures":{"sig_key":{"signer":"76616c","signature":"76616c"}}}]`, read(res))

	res, err = http.Get(fmt.Sprintf("http://%s?type=event2&index=0x%x", api.srv.Addr().String(), "idx1"))
	assert.NoError(t, err)
	assert.JSONEq(t, `[{"timestamp":4,"id":"696434","data":{"data_key":"76616c"},"signatures":{"sig_key":{"signer":"76616c","signature":"76616c"}}}]`, read(res))
}

func read(res *http.Response) string {
	b, _ := io.ReadAll(res.Body)
	return string(b)
}
