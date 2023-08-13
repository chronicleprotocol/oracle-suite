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

package local

import (
	"context"
	"errors"
	"reflect"
	"sync"

	"github.com/chronicleprotocol/oracle-suite/pkg/transport"
	"github.com/chronicleprotocol/oracle-suite/pkg/util/chanutil"
)

const TransportName = "local"

var ErrNotSubscribed = errors.New("topic is not subscribed")

// Local is a simple implementation of the transport.Transport interface
// using local channels.
type Local struct {
	*base
	author []byte
}

type base struct {
	mu     sync.RWMutex
	ctx    context.Context
	waitCh chan error
	subs   map[string]*subscription
}

type rawMsg struct {
	author []byte
	data   []byte
}

type subscription struct {
	// typ is the structure type to which the message must be unmarshalled.
	typ reflect.Type

	// rawMsgCh is a channel used to broadcast raw message data.
	rawMsgCh chan rawMsg

	// msgCh is a channel used to broadcast unmarshalled messages.
	msgCh chan transport.ReceivedMessage

	// msgFanOut is a fan-out demultiplexer for the msgCh channel.
	msgFanOut *chanutil.FanOut[transport.ReceivedMessage]
}

// New returns a new instance of the Local structure. The created transport can
// queue as many unread messages as defined in the queue argument. The list of
// supported subscriptions must be given as a map in the topics argument, where
// the key is the name of the subscription topic, and the value of the map is
// type of the message given as a nil pointer, e.g.: (*Message)(nil).
func New(author []byte, queue int, topics map[string]transport.Message) *Local {
	l := &Local{
		base: &base{
			waitCh: make(chan error),
			subs:   make(map[string]*subscription),
		},
		author: author,
	}
	for topic, typ := range topics {
		msgCh := make(chan transport.ReceivedMessage)
		sub := &subscription{
			typ:       reflect.TypeOf(typ).Elem(),
			rawMsgCh:  make(chan rawMsg, queue),
			msgCh:     msgCh,
			msgFanOut: chanutil.NewFanOut(msgCh),
		}
		l.subs[topic] = sub
		go l.unmarshallRoutine(sub)
	}
	return l
}

func (l *Local) WithAuthor(author []byte) *Local {
	return &Local{base: l.base, author: author}
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
func (l *Local) Wait() <-chan error {
	return l.waitCh
}

// Broadcast implements the transport.Transport interface.
func (l *Local) Broadcast(topic string, message transport.Message) error {
	l.mu.RLock()
	defer l.mu.RUnlock()
	if sub, ok := l.subs[topic]; ok {
		data, err := message.MarshallBinary()
		if err != nil {
			return err
		}
		sub.rawMsgCh <- rawMsg{author: l.author, data: data}
		return nil
	}
	return ErrNotSubscribed
}

// Messages implements the transport.Transport interface.
func (l *Local) Messages(topic string) <-chan transport.ReceivedMessage {
	l.mu.RLock()
	defer l.mu.RUnlock()
	if sub, ok := l.subs[topic]; ok {
		return sub.msgFanOut.Chan()
	}
	return nil
}

func (l *Local) unmarshallRoutine(sub *subscription) {
	for {
		rawMsg, ok := <-sub.rawMsgCh
		if !ok {
			return
		}
		l.mu.RLock()
		msg := reflect.New(sub.typ).Interface().(transport.Message)
		err := msg.UnmarshallBinary(rawMsg.data)
		sub.msgCh <- transport.ReceivedMessage{
			Message: msg,
			Author:  rawMsg.author,
			Error:   err,
			Meta:    transport.Meta{Transport: TransportName},
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
		close(sub.msgCh)
	}
	l.subs = nil
}
