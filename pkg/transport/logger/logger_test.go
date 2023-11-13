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

package logger

import (
	"context"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/chronicleprotocol/oracle-suite/pkg/log"
	"github.com/chronicleprotocol/oracle-suite/pkg/log/callback"
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

func TestLogger(t *testing.T) {
	ctx, ctxCancel := context.WithTimeout(context.Background(), 1*time.Second)
	defer ctxCancel()

	var (
		broadcastMessageLog bool
		receivedMessageLog  bool
	)

	logger := callback.New(log.Debug, func(level log.Level, fields log.Fields, msg string) {
		if level != log.Debug {
			return
		}
		if msg == "Broadcast message" {
			broadcastMessageLog = true
		}
		if msg == "Received message" {
			receivedMessageLog = true
		}
	})
	localTransport := local.New([]byte("test"), 1, map[string]transport.Message{"foo": (*testMsg)(nil)})
	loggerTransport := New(localTransport, logger)
	require.NoError(t, loggerTransport.Start(ctx))

	msgCh := loggerTransport.Messages("foo")
	msg := &testMsg{Val: "bar"}
	require.NoError(t, loggerTransport.Broadcast("foo", msg))
	recv := (<-msgCh).Message
	require.NotNil(t, recv)
	require.Equal(t, msg, recv.(*testMsg))
	assert.True(t, broadcastMessageLog)
	assert.True(t, receivedMessageLog)
}

func TestLogger_Bug_RandMessagesDrop(t *testing.T) {
	// This test is a regression test for a bug that caused random messages to
	// be dropped when using the logger middleware. The bug was caused by the
	// incorrect use of a fan-in channel instead of a fan-out channel.

	ctx, ctxCancel := context.WithTimeout(context.Background(), 1*time.Second)
	defer ctxCancel()

	logger := callback.New(log.Debug, func(level log.Level, fields log.Fields, log string) {})
	localTransport := local.New([]byte("test"), 1, map[string]transport.Message{"foo": (*testMsg)(nil)})
	loggerTransport := New(localTransport, logger)
	require.NoError(t, loggerTransport.Start(ctx))

	msgCh := loggerTransport.Messages("foo")
	for i := 0; i < 100; i++ {
		msg := &testMsg{Val: "bar"}
		require.NoError(t, loggerTransport.Broadcast("foo", msg))
		recv := (<-msgCh).Message
		require.NotNil(t, recv)
		require.Equal(t, msg, recv.(*testMsg))
	}
}
