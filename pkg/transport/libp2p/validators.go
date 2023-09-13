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

package libp2p

import (
	"context"
	"encoding/hex"
	"reflect"
	"time"

	"github.com/defiweb/go-eth/crypto"
	"github.com/defiweb/go-eth/types"
	pubsub "github.com/libp2p/go-libp2p-pubsub"
	"github.com/libp2p/go-libp2p/core/peer"

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
						WithError(err).
						WithFields(log.Fields{
							"peerID":   psMsg.GetFrom().String(),
							"peerAddr": feedAddr.String(),
						}).
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

func feedValidator(feeds []types.Address, logger log.Logger) internal.Options {
	return func(n *internal.Node) error {
		if len(feeds) == 0 {
			return nil
		}
		n.AddValidator(func(ctx context.Context, topic string, id peer.ID, psMsg *pubsub.Message) pubsub.ValidationResult {
			from := ethkey.PeerIDToAddress(psMsg.GetFrom())
			if !feedAllowed(from, feeds) {
				logger.
					WithFields(log.Fields{
						"peerID":   psMsg.GetFrom().String(),
						"peerAddr": from.String(),
					}).
					Warn("Message ignored, feed is not allowed to send messages")
				return pubsub.ValidationIgnore
			}
			return pubsub.ValidationAccept
		})
		return nil
	}
}

func feedAllowed(addr types.Address, feeds []types.Address) bool {
	for _, f := range feeds {
		if f == addr {
			return true
		}
	}
	return false
}

// priceValidator adds a validator for price messages. The validator checks if
// the price message is valid, and if the price is not older than 5 min.
func priceValidator(logger log.Logger, recoverer crypto.Recoverer) internal.Options {
	return func(n *internal.Node) error {
		n.AddValidator(func(ctx context.Context, topic string, id peer.ID, psMsg *pubsub.Message) pubsub.ValidationResult {
			p, ok := psMsg.ValidatorData.(*messages.Price)
			if !ok {
				return pubsub.ValidationAccept
			}
			peerAddr := ethkey.PeerIDToAddress(psMsg.GetFrom())
			fields := log.Fields{
				"peerAddr": peerAddr.String(),
				"peerID":   psMsg.GetFrom().String(),
				"wat":      p.Price.Wat,
				"age":      p.Price.Age.UTC().Format(time.RFC3339),
				"val":      p.Price.Val.String(),
				"version":  p.Version,
				"hash":     hex.EncodeToString(p.Price.Hash().Bytes()),
				"V":        hex.EncodeToString(p.Price.Sig.V.Bytes()),
				"R":        hex.EncodeToString(p.Price.Sig.R.Bytes()),
				"S":        hex.EncodeToString(p.Price.Sig.S.Bytes()),
			}
			// Check is a message signature is valid and extract author's address:
			priceFrom, err := p.Price.From(recoverer)
			if err != nil {
				logger.
					WithError(err).
					WithFields(fields).
					Warn("Price message rejected, invalid signature")
				return pubsub.ValidationReject
			}
			// The libp2p message MUST be created by the same person who signs the price message.
			if *priceFrom != peerAddr {
				logger.
					WithField("from", *priceFrom).
					WithFields(fields).
					Warn("Price message rejected, the message and price signatures do not match")
				return pubsub.ValidationReject
			}
			// Check when message was created, ignore if older than 5 min, reject if older than 10 min:
			if time.Since(p.Price.Age) > 5*time.Minute {
				if time.Since(p.Price.Age) > 10*time.Minute {
					logger.
						WithFields(fields).
						Warn("Price message rejected, the message is older than 10 min")
					return pubsub.ValidationReject
				}
				logger.
					WithFields(fields).
					Warn("Price message ignored, the message is older than 5 min")
				return pubsub.ValidationIgnore
			}
			return pubsub.ValidationAccept
		})
		return nil
	}
}
