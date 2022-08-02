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
	"sync"

	"github.com/chronicleprotocol/oracle-suite/pkg/transport"
)

var ErrNotSubscribed = errors.New("topic is not subscribed")

// Local is a simple implementation of the transport.Transport interface
// using local channels.
type Local struct {
	mu  sync.RWMutex
	ctx context.Context

	id     []byte
	waitCh chan error
	subs   map[string]*subscription
}

type subscription struct {
	// typ is the structure type to which the message must be unmarshalled.
	typ reflect.Type
	// rawMsgs is a channel used to broadcast raw message data.
	rawMsgs chan []byte
	// msgs is a channel used to broadcast unmarshalled messages.
	msgs chan transport.ReceivedMessage
}

// New returns a new instance of the Local structure. The created transport
// can as many unread messages before it is blocked as defined in the queue
// arg. The list of supported subscriptions must be given as a map in the
// topics argument, where the key is the subscription's topic name, and the
// map value is the message type messages given as a nil pointer,
// e.g: (*Message)(nil).
func New(id []byte, queue int, topics map[string]transport.Message) *Local {
	l := &Local{
		id:     id,
		waitCh: make(chan error),
		subs:   make(map[string]*subscription),
	}
	for topic, typ := range topics {
		sub := &subscription{
			typ:     reflect.TypeOf(typ).Elem(),
			rawMsgs: make(chan []byte, queue),
			msgs:    make(chan transport.ReceivedMessage),
		}
		l.subs[topic] = sub
		go l.unmarshallRoutine(sub)
	}
	return l
}

// Start implements the transport.Transport interface.
func (l *Local) Start(ctx context.Context) error {
	if l.ctx != nil {
		return errors.New("service can be started only once")
	}
	if ctx == nil {
		return errors.New("context must not be nil")
	}
	l.ctx = ctx
	go l.contextCancelHandler()
	return nil
}

// Wait implements the transport.Transport interface.
func (l *Local) Wait() chan error {
	return l.waitCh
}

// ID implements the transport.Transport interface.
func (l *Local) ID() []byte {
	return l.id
}

// Broadcast implements the transport.Transport interface.
func (l *Local) Broadcast(topic string, message transport.Message) error {
	if sub, ok := l.subs[topic]; ok {
		b, err := message.MarshallBinary()
		if err != nil {
			return err
		}
		sub.rawMsgs <- b
		return nil
	}
	return ErrNotSubscribed
}

// Messages implements the transport.Transport interface.
func (l *Local) Messages(topic string) chan transport.ReceivedMessage {
	l.mu.RLock()
	defer l.mu.RUnlock()
	if sub, ok := l.subs[topic]; ok {
		return sub.msgs
	}
	return nil
}

func (l *Local) unmarshallRoutine(sub *subscription) {
	for {
		msg, ok := <-sub.rawMsgs
		if !ok {
			return
		}
		l.mu.RLock()
		message := reflect.New(sub.typ).Interface().(transport.Message)
		err := message.UnmarshallBinary(msg)
		sub.msgs <- transport.ReceivedMessage{
			Message: message,
			Author:  l.id,
			Error:   err,
		}
		l.mu.RUnlock()
	}
}

// contextCancelHandler handles context cancellation.
func (l *Local) contextCancelHandler() {
	defer func() { close(l.waitCh) }()
	<-l.ctx.Done()
	l.mu.Lock()
	defer l.mu.Unlock()
	for _, sub := range l.subs {
		close(sub.msgs)
	}
	l.subs = nil
}
