//  Copyright (C) 2021-2023 Chronicle Labs, Inc.
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

package chain

import (
	"context"
	"sync"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"

	"github.com/chronicleprotocol/oracle-suite/pkg/transport"
	"github.com/chronicleprotocol/oracle-suite/pkg/transport/local"
)

type testMsg struct {
	Val string
}

func (t *testMsg) MarshallBinary() ([]byte, error) {
	return []byte(t.Val), nil
}

func (t *testMsg) UnmarshallBinary(bytes []byte) error {
	t.Val = string(bytes)
	return nil
}

func TestChain_Broadcast(t *testing.T) {
	ctx, ctxCancel := context.WithTimeout(context.Background(), 1*time.Second)
	defer ctxCancel()

	l1 := local.New([]byte("test"), 1, map[string]transport.Message{"foo": (*testMsg)(nil)})
	l2 := local.New([]byte("test"), 1, map[string]transport.Message{"foo": (*testMsg)(nil)})

	l := New(l1, l2)
	_ = l.Start(ctx)

	tm := &testMsg{Val: "bar"}

	// Because two local transports are used, the message should be received
	// twice. Also, because each call to the Messages method must create a new
	// fan-out channel, the total number of messages received should be 4.
	m1 := l.Messages("foo")
	m2 := l.Messages("foo")

	assert.NoError(t, l.Broadcast("foo", tm))

	var wg sync.WaitGroup
	wg.Add(1)
	go func() {
		for c := 0; c < 4; c++ {
			select {
			case m, ok := <-m1:
				assert.True(t, ok)
				assert.Equal(t, tm, m.Message)
			case m, ok := <-m2:
				assert.True(t, ok)
				assert.Equal(t, tm, m.Message)
			case <-ctx.Done():
				return
			}
		}
		wg.Done()
	}()

	wg.Wait()
}

func TestChain_Wait(t *testing.T) {
	ctx, ctxCancel := context.WithTimeout(context.Background(), 1*time.Second)
	defer ctxCancel()

	l1 := local.New([]byte("test"), 1, map[string]transport.Message{"foo": (*testMsg)(nil)})
	l2 := local.New([]byte("test"), 1, map[string]transport.Message{"foo": (*testMsg)(nil)})

	l := New(l1, l2)
	_ = l.Start(ctx)

	ctxCancel()

	_, ok := <-l.Wait()
	assert.False(t, ok)
}
