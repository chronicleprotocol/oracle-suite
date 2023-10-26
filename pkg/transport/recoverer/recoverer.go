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

package recoverer

import (
	"context"
	"fmt"

	"github.com/chronicleprotocol/oracle-suite/pkg/log"
	"github.com/chronicleprotocol/oracle-suite/pkg/log/null"
	"github.com/chronicleprotocol/oracle-suite/pkg/supervisor"
	"github.com/chronicleprotocol/oracle-suite/pkg/transport"
)

// Recoverer is a transport wrapper that handles panics that occur in the
// underlying transport.
type Recoverer struct {
	t transport.Service
	l log.Logger
}

// New creates a new Recoverer transport.
func New(t transport.Service, l log.Logger) *Recoverer {
	if t == nil {
		panic("t cannot be nil")
	}
	if l == nil {
		l = null.New()
	}
	return &Recoverer{t: t, l: l}
}

// Start implements the transport.Transport interface.
func (r *Recoverer) Start(ctx context.Context) error {
	return r.t.Start(ctx)
}

// Wait implements the transport.Transport interface.
func (r *Recoverer) Wait() <-chan error {
	return r.t.Wait()
}

// Broadcast implements the transport.Transport interface.
func (r *Recoverer) Broadcast(topic string, message transport.Message) (err error) {
	defer func() {
		if rec := recover(); rec != nil {
			r.l.
				WithField("panic", rec).
				WithAdvice("This is a critical bug and must be investigated").
				Error("Recovered from panic")
			err = fmt.Errorf("recovered from panic: %v", rec)
		}
	}()
	return r.t.Broadcast(topic, message)
}

// Messages implements the transport.Transport interface.
func (r *Recoverer) Messages(topic string) <-chan transport.ReceivedMessage {
	defer func() {
		if rec := recover(); rec != nil {
			r.l.
				WithField("panic", rec).
				WithAdvice("This is a critical bug and must be investigated").
				Error("Recovered from panic")
		}
	}()
	return r.t.Messages(topic)
}

// ServiceName implements the supervisor.WithName interface.
func (r *Recoverer) ServiceName() string {
	return fmt.Sprintf("Recoverer(%s)", supervisor.ServiceName(r.t))
}
