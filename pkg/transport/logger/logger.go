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
	"fmt"

	"github.com/chronicleprotocol/oracle-suite/pkg/log"
	"github.com/chronicleprotocol/oracle-suite/pkg/log/null"
	"github.com/chronicleprotocol/oracle-suite/pkg/supervisor"
	"github.com/chronicleprotocol/oracle-suite/pkg/transport"
	"github.com/chronicleprotocol/oracle-suite/pkg/util/chanutil"
)

// Logger logs all messages sent and received by the transport.
type Logger struct {
	t transport.Service
	l log.Logger
}

// New creates a new Logger transport.
func New(t transport.Service, l log.Logger) *Logger {
	if t == nil {
		panic("t cannot be nil")
	}
	if l == nil {
		l = null.New()
	}
	return &Logger{t: t, l: l}
}

// Start implements the transport.Transport interface.
func (r *Logger) Start(ctx context.Context) error {
	return r.t.Start(ctx)
}

// Wait implements the transport.Transport interface.
func (r *Logger) Wait() <-chan error {
	return r.t.Wait()
}

// Broadcast implements the transport.Transport interface.
func (r *Logger) Broadcast(topic string, msg transport.Message) error {
	if !log.IsLevel(r.l, log.Debug) {
		return r.t.Broadcast(topic, msg)
	}
	err := r.t.Broadcast(topic, msg)
	log := r.l.
		WithFields(log.Fields{
			"topic":   topic,
			"message": msg,
		})
	if err != nil {
		log.WithError(err)
	}
	log.Debug("Broadcast message")
	return err
}

// Messages implements the transport.Transport interface.
func (r *Logger) Messages(topic string) <-chan transport.ReceivedMessage {
	if !log.IsLevel(r.l, log.Debug) {
		return r.t.Messages(topic)
	}
	in := r.t.Messages(topic)
	if in == nil {
		// It is possible that the underlying transport does not support
		// given topic. In such case, it will return nil.
		return nil
	}
	fo := chanutil.NewFanOut(in)
	go func() {
		for msg := range fo.Chan() {
			r.l.
				WithFields(log.Fields{
					"topic":   topic,
					"message": msg,
				}).
				Debug("Received message")
		}
	}()
	return fo.Chan()
}

// ServiceName implements the supervisor.WithName interface.
func (r *Logger) ServiceName() string {
	return fmt.Sprintf("Logger(%s)", supervisor.ServiceName(r.t))
}
