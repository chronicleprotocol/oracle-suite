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
	"sync"
	"time"

	"github.com/libp2p/go-libp2p-core/peer"
	pubsub "github.com/libp2p/go-libp2p-pubsub"
	"golang.org/x/time/rate"

	"github.com/chronicleprotocol/oracle-suite/pkg/transport/libp2p/internal/sets"

	"github.com/chronicleprotocol/oracle-suite/pkg/log"
)

// RateLimiterConfig is a configuration for the RateLimiter option.
type RateLimiterConfig struct {
	// BytesPerSecond is the maximum rate of messages in bytes/s that can be
	// created by a single peer.
	BytesPerSecond float64
	// BurstSize is a burst value in bytes for a messages created from a singe
	// peer.
	BurstSize int
	// RelayBytesPerSecond is the maximum rate of messages in bytes/s that can
	// be relayed by a single relay.
	RelayBytesPerSecond float64
	// RelayBurstSize is a burst value in bytes for a messages relayed by
	// a singe peer.
	RelayBurstSize int
}

type rateLimiter struct {
	mu sync.Mutex

	bytesPerSecond float64
	burstSize      int
	peerLimiters   map[peer.ID]*peerLimiter

	// gcTTL is a time since last message after which peer will be removed
	// by the gc method.
	gcTTL time.Duration
}

func newRateLimiter(bytesPerSecond float64, burstSize int) *rateLimiter {
	return &rateLimiter{
		bytesPerSecond: bytesPerSecond,
		burstSize:      burstSize,
		peerLimiters:   make(map[peer.ID]*peerLimiter),
		// time required to fill an empty token bucket
		gcTTL: time.Second * time.Duration(float64(burstSize)/bytesPerSecond),
	}
}

type peerLimiter struct {
	limiter *rate.Limiter
	lastMsg time.Time // lastMsg is a time since last message.
}

// peerLimiter creates or returns previously created limiter for a given peer.
func (p *rateLimiter) peerLimiter(id peer.ID) *peerLimiter {
	if _, ok := p.peerLimiters[id]; !ok {
		p.peerLimiters[id] = &peerLimiter{
			limiter: rate.NewLimiter(rate.Limit(p.bytesPerSecond), p.burstSize),
			lastMsg: time.Now(),
		}
	}
	return p.peerLimiters[id]
}

// allow checks if a message of a given size can be received from a given peer.
func (p *rateLimiter) allow(id peer.ID, msgSize int) bool {
	p.mu.Lock()
	defer p.mu.Unlock()
	prl := p.peerLimiter(id)
	prl.lastMsg = time.Now()
	return prl.limiter.AllowN(prl.lastMsg, msgSize)
}

// gc removes inactive peers.
func (p *rateLimiter) gc() {
	p.mu.Lock()
	defer p.mu.Unlock()
	for id, pl := range p.peerLimiters {
		if time.Since(pl.lastMsg) >= p.gcTTL {
			delete(p.peerLimiters, id)
		}
	}
}

// RateLimiter limits the number of bytes which is allowed to receive from
// the network using the token bucket algorithm:
// https://en.wikipedia.org/wiki/Token_bucket
//
// bytesPerSecond is the maximum number of bytes/s that can be
// received from a single peer.
//
// burstSize is a burst value in bytes applied for a messages received
// from a singe peer.
func RateLimiter(cfg RateLimiterConfig) Options {
	return func(n *Node) error {
		// Rate limiter for message relays:
		relayRL := newRateLimiter(cfg.RelayBytesPerSecond, cfg.RelayBurstSize)
		// Rate limiter for message authors:
		msgRL := newRateLimiter(cfg.BytesPerSecond, cfg.BurstSize)
		n.AddValidator(func(ctx context.Context, topic string, id peer.ID, msg *pubsub.Message) pubsub.ValidationResult {
			if n.Host().ID() == id {
				return pubsub.ValidationAccept
			}
			// The order of the following checks is important because each
			// "allow" call uses "tokens" and message authors should not
			// be penalized for misbehaving relays.
			if !relayRL.allow(id, len(msg.Data)) {
				n.tsLog.get().
					WithFields(log.Fields{
						"topic":              topic,
						"peerID":             msg.GetFrom().String(),
						"receivedFromPeerID": msg.ReceivedFrom.String(),
					}).
					Debug("The message has been rejected, rate limit for relay exceeded")
				return pubsub.ValidationReject
			}
			if !msgRL.allow(msg.GetFrom(), len(msg.Data)) {
				n.tsLog.get().
					WithFields(log.Fields{
						"topic":              topic,
						"peerID":             msg.GetFrom().String(),
						"receivedFromPeerID": msg.ReceivedFrom.String(),
					}).
					Debug("The message has been rejected, rate limit for message author exceeded")
				return pubsub.ValidationReject
			}
			return pubsub.ValidationAccept
		})
		n.AddNodeEventHandler(sets.NodeEventHandlerFunc(func(event interface{}) {
			if _, ok := event.(sets.NodeStartedEvent); ok {
				go func() {
					t := time.NewTicker(time.Minute)
					defer t.Stop()
					for {
						select {
						case <-n.ctx.Done():
							return
						case <-t.C:
							relayRL.gc()
						}
					}
				}()
			}
		}))
		return nil
	}
}
