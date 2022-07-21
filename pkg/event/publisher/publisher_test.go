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

package publisher

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

type testListener struct{ ch chan *messages.Event }
type testSigner struct{}

func (t *testListener) Start(_ context.Context) error {
	return nil
}

func (t *testListener) Events() chan *messages.Event {
	return t.ch
}

func (t testSigner) Sign(event *messages.Event) (bool, error) {
	event.Signatures["test"] = messages.EventSignature{
		Signer:    []byte("signer"),
		Signature: []byte("signature"),
	}
	return true, nil
}

func TestEventPublisher(t *testing.T) {
	ctx, cancelFunc := context.WithCancel(context.Background())
	loc := local.New([]byte("test"), 10, map[string]transport.Message{messages.EventV1MessageName: (*messages.Event)(nil)})
	lis := &testListener{ch: make(chan *messages.Event, 10)}
	sig := &testSigner{}

	pub, err := New(Config{
		Listeners: []EventProvider{lis},
		Signers:   []Signer{sig},
		Transport: loc,
		Logger:    null.New(),
	})
	require.NoError(t, err)

	require.NoError(t, loc.Start(ctx))
	require.NoError(t, pub.Start(ctx))
	defer func() {
		cancelFunc()
		require.NoError(t, <-loc.Wait())
		require.NoError(t, <-pub.Wait())
	}()

	msg1 := &messages.Event{
		Type:        "event1",
		ID:          []byte("id1"),
		Index:       []byte("idx1"),
		EventDate:   time.Unix(1, 0),
		MessageDate: time.Unix(1, 0),
		Data:        map[string][]byte{"data_key": []byte("val")},
		Signatures:  map[string]messages.EventSignature{"sig_key": {Signer: []byte("val"), Signature: []byte("val")}},
	}
	msg2 := &messages.Event{
		Type:        "event2",
		ID:          []byte("id2"),
		Index:       []byte("idx2"),
		EventDate:   time.Unix(2, 0),
		MessageDate: time.Unix(2, 0),
		Data:        map[string][]byte{"data_key": []byte("val")},
		Signatures:  map[string]messages.EventSignature{"sig_key": {Signer: []byte("val"), Signature: []byte("val")}},
	}
	lis.ch <- msg1
	lis.ch <- msg2

	time.Sleep(100 * time.Millisecond)

	rMsg1 := <-loc.Messages(messages.EventV1MessageName)
	rMsg2 := <-loc.Messages(messages.EventV1MessageName)

	assert.Equal(t, []byte("signer"), rMsg1.Message.(*messages.Event).Signatures["test"].Signer)
	assert.Equal(t, []byte("signer"), rMsg2.Message.(*messages.Event).Signatures["test"].Signer)
	assert.Equal(t, []byte("signature"), rMsg1.Message.(*messages.Event).Signatures["test"].Signature)
	assert.Equal(t, []byte("signature"), rMsg2.Message.(*messages.Event).Signatures["test"].Signature)
	// This test relies on us passing the same message instances, so the values
	// added by the signer will be visible in all objects, but this behavior is
	// not required.
	assert.Equal(t, msg1, rMsg1.Message.(*messages.Event))
	assert.Equal(t, msg2, rMsg2.Message.(*messages.Event))
}
