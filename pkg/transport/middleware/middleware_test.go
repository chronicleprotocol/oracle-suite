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

package middleware

import (
	"context"
	"testing"

	"github.com/stretchr/testify/require"

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

func TestMiddleware_Broadcast(t *testing.T) {
	ctx, ctxCancel := context.WithCancel(context.Background())
	defer ctxCancel()

	l := local.New([]byte("test"), 1, map[string]transport.Message{"foo": (*testMsg)(nil)})
	m := New(l)
	require.NoError(t, m.Start(ctx))

	m.Use(BroadcastMiddlewareFunc(func(_ context.Context, next BroadcastFunc) BroadcastFunc {
		return func(topic string, message transport.Message) error {
			// Modify the message before broadcasting.
			message.(*testMsg).Val = "bar"
			err := next(topic, message)
			// Modify the message after broadcasting. This should not affect the
			// message received later.
			message.(*testMsg).Val = "baz"
			return err
		}
	}))

	require.NoError(t, m.Broadcast("foo", &testMsg{Val: "bar"}))
	require.Equal(t, "bar", (<-m.Messages("foo")).Message.(*testMsg).Val)
}

func TestMiddleware_Broadcast_Order(t *testing.T) {
	ctx, ctxCancel := context.WithCancel(context.Background())
	defer ctxCancel()

	l := local.New([]byte("test"), 1, map[string]transport.Message{"foo": (*testMsg)(nil)})
	m := New(l)
	require.NoError(t, m.Start(ctx))

	var order []int

	m.Use(BroadcastMiddlewareFunc(func(_ context.Context, next BroadcastFunc) BroadcastFunc {
		return func(topic string, message transport.Message) error {
			order = append(order, 1)
			err := next(topic, message)
			order = append(order, 4)
			return err
		}
	}))

	m.Use(BroadcastMiddlewareFunc(func(_ context.Context, next BroadcastFunc) BroadcastFunc {
		return func(topic string, message transport.Message) error {
			order = append(order, 2)
			err := next(topic, message)
			order = append(order, 3)
			return err
		}
	}))

	require.NoError(t, m.Broadcast("foo", &testMsg{Val: "bar"}))
	require.Equal(t, []int{1, 2, 3, 4}, order)
	<-m.Messages("foo")
}
