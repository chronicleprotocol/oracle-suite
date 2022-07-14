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

package internal

import (
	"context"
	"errors"

	"github.com/libp2p/go-libp2p-core/peer"
	pubsub "github.com/libp2p/go-libp2p-pubsub"

	"github.com/chronicleprotocol/oracle-suite/pkg/transport/libp2p/internal/sets"
)

var ErrNilMessage = errors.New("message is nil")

type Subscription struct {
	ctx       context.Context
	ctxCancel context.CancelFunc

	topic          *pubsub.Topic
	sub            *pubsub.Subscription
	teh            *pubsub.TopicEventHandler
	validatorSet   *sets.ValidatorSet
	eventHandler   sets.PubSubEventHandler
	messageHandler sets.MessageHandler

	// msgCh is used to send a notification about a new message, it's
	// returned by the Transport.Messages function.
	msgCh chan *pubsub.Message
}

func newSubscription(node *Node, topic string) (*Subscription, error) {
	var err error
	ctx, ctxCancel := context.WithCancel(node.ctx)
	s := &Subscription{
		ctx:            ctx,
		ctxCancel:      ctxCancel,
		validatorSet:   node.validatorSet,
		eventHandler:   node.pubSubEventHandlerSet,
		messageHandler: node.messageHandlerSet,
		msgCh:          make(chan *pubsub.Message),
	}
	err = node.pubSub.RegisterTopicValidator(topic, s.validator(topic))
	if err != nil {
		return nil, err
	}
	s.topic, err = node.PubSub().Join(topic)
	if err != nil {
		return nil, err
	}
	s.sub, err = s.topic.Subscribe()
	if err != nil {
		return nil, err
	}
	s.teh, err = s.topic.EventHandler()
	if err != nil {
		return nil, err
	}
	go s.messageLoop()
	go s.eventLoop()
	return s, err
}

func (s *Subscription) Publish(msg []byte) error {
	if msg == nil {
		return ErrNilMessage
	}
	s.messageHandler.Published(s.topic.String(), msg)
	return s.topic.Publish(s.ctx, msg)
}

func (s *Subscription) Next() chan *pubsub.Message {
	return s.msgCh
}

func (s *Subscription) validator(topic string) pubsub.ValidatorEx {
	return func(ctx context.Context, id peer.ID, psMsg *pubsub.Message) pubsub.ValidationResult {
		vr := s.validatorSet.Validator(topic)(ctx, id, psMsg)
		s.messageHandler.Received(topic, psMsg, vr)
		return vr
	}
}

func (s *Subscription) messageLoop() {
	defer close(s.msgCh)
	func() {
		for {
			msg, err := s.sub.Next(s.ctx)
			if err != nil {
				// The only time when an error may be returned here is
				// when the subscription is canceled.
				return
			}
			s.msgCh <- msg
		}
	}()
}

func (s *Subscription) eventLoop() {
	func() {
		for {
			pe, err := s.teh.NextPeerEvent(s.ctx)
			if err != nil {
				// The only time when an error may be returned here is
				// when the subscription is canceled.
				return
			}
			s.eventHandler.Handle(s.topic.String(), pe)
		}
	}()
}

func (s *Subscription) close() error {
	s.ctxCancel()
	s.teh.Cancel()
	s.sub.Cancel()
	return s.topic.Close()
}
