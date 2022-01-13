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
	"errors"
	"reflect"

	"github.com/chronicleprotocol/oracle-suite/pkg/transport"
)

var ErrNotSubscribed = errors.New("topic is not subscribed")

// Local is a simple implementation of the transport.Transport interface
// using local channels.
type Local struct {
	ctx    context.Context
	waitCh chan error
	subs   map[string]subscription
}

type subscription struct {
	// typ is the structure type to which the message must be unmarshalled.
	typ reflect.Type
	// msgs is a channel used to broadcast raw message data.
	msgs chan []byte
	// status is a channel used to broadcast unmarshalled messages.
	status chan transport.ReceivedMessage
}

// New returns a new instance of the Local structure. The created transport
// can handle up to b unread messages before it is blocked. The list of
// supported subscriptions must be given as a map in s argument, where the
// key is the subscription's topic name, and the map value is the message type
// messages given as a nil pointer, e.g: (*Message)(nil).
func New(ctx context.Context, b int, s map[string]transport.Message) *Local {
	l := &Local{
		ctx:    ctx,
		waitCh: make(chan error),
		subs:   make(map[string]subscription),
	}
	for topic, typ := range s {
		l.subs[topic] = subscription{
			typ:    reflect.TypeOf(typ).Elem(),
			msgs:   make(chan []byte, b),
			status: make(chan transport.ReceivedMessage),
		}
	}
	return l
}

// Start implements the transport.Transport interface.
func (l *Local) Start() error {
	go l.contextCancelHandler()
	return nil
}

// Wait implements the transport.Transport interface.
func (l *Local) Wait() chan error {
	return l.waitCh
}

// Broadcast implements the transport.Transport interface.
func (l *Local) Broadcast(topic string, message transport.Message) error {
	if sub, ok := l.subs[topic]; ok {
		b, err := message.MarshallBinary()
		if err != nil {
			return err
		}
		sub.msgs <- b
		return nil
	}
	return ErrNotSubscribed
}

// Messages implements the transport.Transport interface.
func (l *Local) Messages(topic string) chan transport.ReceivedMessage {
	if sub, ok := l.subs[topic]; ok {
		go func() {
			select {
			case <-l.ctx.Done():
				return
			case msg := <-sub.msgs:
				message := reflect.New(sub.typ).Interface().(transport.Message)
				sub.status <- transport.ReceivedMessage{
					Message: message,
					Error:   message.UnmarshallBinary(msg),
				}
			}
		}()
		return sub.status
	}
	return nil
}

// contextCancelHandler handles context cancellation.
func (l *Local) contextCancelHandler() {
	defer func() { l.waitCh <- nil }()
	<-l.ctx.Done()
	for _, sub := range l.subs {
		close(sub.status)
	}
	l.subs = make(map[string]subscription)
}
