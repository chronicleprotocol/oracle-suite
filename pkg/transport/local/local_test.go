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

package local

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/chronicleprotocol/oracle-suite/pkg/transport"
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

func TestLocal_Broadcast(t *testing.T) {
	ctx, ctxCancel := context.WithCancel(context.Background())
	defer ctxCancel()

	l := New([]byte("test"), 1, map[string]transport.Message{"foo": (*testMsg)(nil)})
	_ = l.Start(ctx)

	// Valid message:
	vm := &testMsg{Val: "bar"}
	assert.NoError(t, l.Broadcast("foo", vm))
}

func TestLocal_Messages(t *testing.T) {
	ctx, ctxCancel := context.WithCancel(context.Background())
	defer ctxCancel()

	l := New([]byte("test"), 1, map[string]transport.Message{"foo": (*testMsg)(nil)})
	_ = l.Start(ctx)

	// Valid message:
	assert.NoError(t, l.Broadcast("foo", &testMsg{Val: "bar"}))
	assert.Equal(t, &testMsg{Val: "bar"}, (<-l.Messages("foo")).Message)
}
