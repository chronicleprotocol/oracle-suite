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

package chain

import (
	"bytes"
	"context"
	"errors"

	"github.com/chronicleprotocol/oracle-suite/pkg/transport"
	"github.com/chronicleprotocol/oracle-suite/pkg/util/chanutil"
)

// Chain is a transport implementation that chains multiple transports
// together.
type Chain struct {
	ctx    context.Context
	waitCh <-chan error
	ts     []transport.Transport
}

// New creates a new Chain instance.
func New(ts ...transport.Transport) *Chain {
	fi := chanutil.NewFanIn[error]()
	for _, t := range ts {
		_ = fi.Add(t.Wait())
	}
	fi.AutoClose()
	return &Chain{
		waitCh: fi.Chan(),
		ts:     ts,
	}
}

// ID implements the transport.Transport interface.
func (m *Chain) ID() []byte {
	var ids [][]byte
	for _, t := range m.ts {
		ids = append(ids, t.ID())
	}
	return bytes.Join(ids, []byte{','})
}

// Broadcast implements the transport.Transport interface.
func (m *Chain) Broadcast(topic string, message transport.Message) error {
	var err error
	for _, t := range m.ts {
		if bErr := t.Broadcast(topic, message); bErr != nil {
			err = bErr // TODO(mdobak): Collect all errors.
		}
	}
	return err
}

// Messages implements the transport.Transport interface.
func (m *Chain) Messages(topic string) <-chan transport.ReceivedMessage {
	fi := chanutil.NewFanIn[transport.ReceivedMessage]()
	for _, t := range m.ts {
		_ = fi.Add(t.Messages(topic))
	}
	fi.AutoClose()
	return fi.Chan()
}

// Start implements the transport.Transport interface.
func (m *Chain) Start(ctx context.Context) error {
	if m.ctx != nil {
		return errors.New("service can be started only once")
	}
	if ctx == nil {
		return errors.New("context must not be nil")
	}
	m.ctx = ctx
	for _, t := range m.ts {
		if err := t.Start(ctx); err != nil {
			return err
		}
	}
	return nil
}

// Wait implements the transport.Transport interface.
func (m *Chain) Wait() <-chan error {
	return m.waitCh
}
