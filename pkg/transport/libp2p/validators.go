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

package libp2p

import (
	"context"
	"reflect"
	"time"

	"github.com/libp2p/go-libp2p-core/peer"
	pubsub "github.com/libp2p/go-libp2p-pubsub"

	"github.com/chronicleprotocol/oracle-suite/pkg/ethereum"
	"github.com/chronicleprotocol/oracle-suite/pkg/log"
	"github.com/chronicleprotocol/oracle-suite/pkg/transport"
	"github.com/chronicleprotocol/oracle-suite/pkg/transport/libp2p/crypto/ethkey"
	"github.com/chronicleprotocol/oracle-suite/pkg/transport/libp2p/internal"
	"github.com/chronicleprotocol/oracle-suite/pkg/transport/messages"
)

func messageValidator(topics map[string]transport.Message, logger log.Logger) internal.Options {
	return func(n *internal.Node) error {
		// Validator actually have two roles in the libp2p: it unmarshalls messages
		// and then validates them. Unmarshalled message is stored in the
		// ValidatorData field which was created for this purpose:
		// https://github.com/libp2p/go-libp2p-pubsub/pull/231
		n.AddValidator(func(ctx context.Context, topic string, id peer.ID, psMsg *pubsub.Message) pubsub.ValidationResult {
			if typ, ok := topics[topic]; ok {
				typRefl := reflect.TypeOf(typ).Elem()
				msg := reflect.New(typRefl).Interface().(transport.Message)
				err := msg.UnmarshallBinary(psMsg.Data)
				if err != nil {
					feedAddr := ethkey.PeerIDToAddress(psMsg.GetFrom())
					logger.
						WithField("peerID", psMsg.GetFrom().String()).
						WithField("from", feedAddr).
						Warn("The message has been rejected, unable to unmarshall")
					return pubsub.ValidationReject
				}
				psMsg.ValidatorData = msg
				return pubsub.ValidationAccept
			}
			return pubsub.ValidationIgnore // should never happen
		})
		return nil
	}
}

func feederValidator(feeders []ethereum.Address, logger log.Logger) internal.Options {
	return func(n *internal.Node) error {
		n.AddValidator(func(ctx context.Context, topic string, id peer.ID, psMsg *pubsub.Message) pubsub.ValidationResult {
			feedAddr := ethkey.PeerIDToAddress(psMsg.GetFrom())
			feedAllowed := false
			for _, addr := range feeders {
				if addr == feedAddr {
					feedAllowed = true
					break
				}
			}
			if !feedAllowed {
				logger.
					WithField("peerID", psMsg.GetFrom().String()).
					WithField("from", feedAddr).
					Warn("The message has been ignored, the feeder is not allowed to send messages")
				return pubsub.ValidationIgnore
			}
			return pubsub.ValidationAccept
		})
		return nil
	}
}

// eventValidator adds a validator for event messages.
func eventValidator(logger log.Logger) internal.Options {
	return func(n *internal.Node) error {
		n.AddValidator(func(ctx context.Context, topic string, id peer.ID, psMsg *pubsub.Message) pubsub.ValidationResult {
			eventMsg, ok := psMsg.ValidatorData.(*messages.Event)
			if !ok {
				return pubsub.ValidationAccept
			}
			feedAddr := ethkey.PeerIDToAddress(psMsg.GetFrom())
			// Check when message was created, ignore if older than 5 min, reject if older than 10 min:
			if time.Since(eventMsg.MessageDate) > 5*time.Minute {
				logger.
					WithField("peerID", psMsg.GetFrom().String()).
					WithField("from", feedAddr.String()).
					Warn("The event message has been rejected, the message is older than 5 min")
				if time.Since(eventMsg.MessageDate) > 10*time.Minute {
					return pubsub.ValidationReject
				}
				return pubsub.ValidationIgnore
			}
			return pubsub.ValidationAccept
		})
		return nil
	}
}

// priceValidator adds a validator for price messages. The validator checks if
// the price message is valid, and if the price is not older than 5 min.
func priceValidator(signer ethereum.Signer, logger log.Logger) internal.Options {
	return func(n *internal.Node) error {
		n.AddValidator(func(ctx context.Context, topic string, id peer.ID, psMsg *pubsub.Message) pubsub.ValidationResult {
			priceMsg, ok := psMsg.ValidatorData.(*messages.Price)
			if !ok {
				return pubsub.ValidationAccept
			}
			// Check is a message signature is valid and extract author's address:
			priceFrom, err := priceMsg.Price.From(signer)
			wat := priceMsg.Price.Wat
			age := priceMsg.Price.Age.UTC().Format(time.RFC3339)
			val := priceMsg.Price.Val.String()
			if err != nil {
				logger.
					WithError(err).
					WithField("peerID", psMsg.GetFrom().String()).
					WithField("wat", wat).
					WithField("age", age).
					WithField("val", val).
					Warn("The price message has been rejected, invalid signature")
				return pubsub.ValidationReject
			}
			// The libp2p message should be created by the same person who signs the price message:
			if ethkey.AddressToPeerID(*priceFrom) != psMsg.GetFrom() {
				logger.
					WithField("peerID", psMsg.GetFrom().String()).
					WithField("from", priceFrom.String()).
					WithField("wat", wat).
					WithField("age", age).
					WithField("val", val).
					Warn("The price message has been rejected, the message and price signatures do not match")
				return pubsub.ValidationReject
			}
			// Check when message was created, ignore if older than 5 min, reject if older than 10 min:
			if time.Since(priceMsg.Price.Age) > 5*time.Minute {
				logger.
					WithField("peerID", psMsg.GetFrom().String()).
					WithField("from", priceFrom.String()).
					WithField("wat", wat).
					WithField("age", age).
					WithField("val", val).
					Warn("The price message has been rejected, the message is older than 5 min")
				if time.Since(priceMsg.Price.Age) > 10*time.Minute {
					return pubsub.ValidationReject
				}
				return pubsub.ValidationIgnore
			}
			return pubsub.ValidationAccept
		})
		return nil
	}
}
