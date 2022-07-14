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
	"errors"
	"math"
	"time"

	"github.com/libp2p/go-libp2p-core/peer"
	pubsub "github.com/libp2p/go-libp2p-pubsub"
)

// Peer scoring:
// https://github.com/libp2p/specs/blob/master/pubsub/gossipsub/gossipsub-v1.1.md#peer-scoring

const decayInterval = time.Minute
const decayToZero = 0.01

var thresholds = &pubsub.PeerScoreThresholds{
	// -4000 is sum of P₃ and P₃b for the "price" and "event" topics. It should
	// be equal to the lowest score a silent peer can get without receiving any
	// penalties other than P₃ and P₃b. Because only very few peers can produce
	// messages, some honest peers may have a score equal to this number.
	GossipThreshold:             -4000,
	PublishThreshold:            -4000,
	GraylistThreshold:           -4000,
	AcceptPXThreshold:           0,
	OpportunisticGraftThreshold: 0,
}

var peerScoreParams = &pubsub.PeerScoreParams{
	AppSpecificScore:            func(id peer.ID) float64 { return 0 },
	AppSpecificWeight:           1,
	IPColocationFactorWeight:    -10,
	IPColocationFactorThreshold: 2,
	BehaviourPenaltyWeight:      -1,
	BehaviourPenaltyThreshold:   1,
	BehaviourPenaltyDecay:       decay(1, time.Hour),
	DecayInterval:               decayInterval,
	DecayToZero:                 decayToZero,
	RetainScore:                 15 * time.Minute,
	Topics:                      make(map[string]*pubsub.TopicScoreParams),
}

func calculatePriceTopicScoreParams(cfg Config) (*pubsub.TopicScoreParams, error) {
	var maxPeers = float64(pubsub.GossipSubDhi)
	// Minimum and maximum expected number of feeders connected to the network:
	var minFeederCount = float64(len(cfg.FeedersAddrs)) / 2 // assume that 50% of feeders are offline
	var maxFeederCount = float64(len(cfg.FeedersAddrs))
	// Minimum and maximum expected number of messages to be received from a single peer in a mesh:
	var minMsgsPerSecond = (minFeederCount * minAssetPairs) / maxPeers / priceUpdateInterval.Seconds()
	var maxMsgsPerSecond = (maxFeederCount * maxAssetPairs) / priceUpdateInterval.Seconds()

	//nolint:gomnd
	return (&scoreParams{
		p1Score:              500,
		p2Score:              500,
		p3Score:              -1000,
		p3bScore:             -1000,
		p4Score:              -1000,
		p1Length:             15 * time.Minute,
		p2Length:             15 * time.Minute,
		p3Length:             15 * time.Minute,
		p3bLength:            15 * time.Minute,
		p4Length:             time.Hour,
		minMessagesPerSecond: minMsgsPerSecond,
		maxMessagesPerSecond: maxMsgsPerSecond,
		maxInvalidMessages:   maxInvalidMsgsPerHour,
	}).calculate()
}

func calculateEventTopicScoreParams(cfg Config) (*pubsub.TopicScoreParams, error) {
	// NOTE: The scoring parameters for events are just guesses at the moment, we will have to update them when we
	// know how many events we can expect.

	var maxPeers = float64(pubsub.GossipSubDhi)
	// Minimum and maximum expected number of feeders connected to the network:
	var minFeederCount = float64(len(cfg.FeedersAddrs)) / 2 // assume that 50% of feeders are offline
	var maxFeederCount = float64(len(cfg.FeedersAddrs))
	// Minimum and maximum expected number of messages to be received from a single peer in a mesh:
	var minMsgsPerSecond = minFeederCount * minEventsPerSecond / maxPeers
	var maxMsgsPerSecond = maxFeederCount * maxEventsPerSecond

	//nolint:gomnd
	return (&scoreParams{
		p1Score:              500,
		p2Score:              500,
		p3Score:              -1000,
		p3bScore:             -1000,
		p4Score:              -1000,
		p1Length:             120 * time.Minute,
		p2Length:             120 * time.Minute,
		p3Length:             120 * time.Minute,
		p3bLength:            15 * time.Minute,
		p4Length:             time.Hour,
		minMessagesPerSecond: minMsgsPerSecond,
		maxMessagesPerSecond: maxMsgsPerSecond,
		maxInvalidMessages:   maxInvalidMsgsPerHour,
	}).calculate()
}

// scoreParams helps to calculate score parameters for libp2p's PubSub.
type scoreParams struct {
	p1Score              float64
	p2Score              float64
	p3Score              float64
	p3bScore             float64
	p4Score              float64
	p1Length             time.Duration
	p2Length             time.Duration
	p3Length             time.Duration
	p3bLength            time.Duration
	p4Length             time.Duration
	minMessagesPerSecond float64
	maxMessagesPerSecond float64
	maxInvalidMessages   float64
}

func (p *scoreParams) calculate() (*pubsub.TopicScoreParams, error) {
	if p.p1Score < 0 {
		return nil, errors.New("p1Score must be positive")
	}
	if p.p2Score < 0 {
		return nil, errors.New("p2Score must be positive")
	}
	if p.p3Score > 0 {
		return nil, errors.New("p3Score must be negative")
	}
	if p.p3bScore > 0 {
		return nil, errors.New("p3bScore must be negative")
	}
	if p.p4Score > 0 {
		return nil, errors.New("p4Score must be negative")
	}
	return &pubsub.TopicScoreParams{
		TopicWeight: 1,

		// P₁
		TimeInMeshWeight:  p.p1Score / p.p1Length.Seconds(),
		TimeInMeshQuantum: time.Second,
		TimeInMeshCap:     p.p1Length.Seconds(),

		// P₂
		FirstMessageDeliveriesWeight: p.p2Score / (p.minMessagesPerSecond * p.p2Length.Seconds()),
		FirstMessageDeliveriesDecay:  decay(p.minMessagesPerSecond*p.p2Length.Seconds(), p.p2Length),
		FirstMessageDeliveriesCap:    p.minMessagesPerSecond * p.p2Length.Seconds(),

		// P₃
		// The p3Length parameter specifies the decay time for minMessagesPerSecond. If the number of messages is higher
		// than the decay, time will be longer.
		MeshMessageDeliveriesWeight:     p.p3Score / math.Pow(p.minMessagesPerSecond*p.p3Length.Seconds(), 2),
		MeshMessageDeliveriesDecay:      decay(p.minMessagesPerSecond*p.p3Length.Seconds(), p.p3Length),
		MeshMessageDeliveriesCap:        p.maxMessagesPerSecond * p.p3Length.Seconds(),
		MeshMessageDeliveriesThreshold:  p.minMessagesPerSecond * p.p3Length.Seconds(),
		MeshMessageDeliveriesActivation: p.p3Length,
		MeshMessageDeliveriesWindow:     10 * time.Millisecond,

		// P₃b
		// Use of p3Length is not a mistake. The maximum score (before applying weight) will be equal to
		// MeshMessageDeliveriesThreshold, thus it should be calculated in the same way.
		MeshFailurePenaltyWeight: p.p3bScore / math.Pow(p.minMessagesPerSecond*p.p3Length.Seconds(), 2),
		MeshFailurePenaltyDecay:  decay(p.minMessagesPerSecond*p.p3Length.Seconds(), p.p3bLength),

		// P₄
		InvalidMessageDeliveriesWeight: p.p4Score / math.Pow(p.maxInvalidMessages, 2),
		InvalidMessageDeliveriesDecay:  decay(p.maxInvalidMessages, p.p4Length),
	}, nil
}

// decay calculates a decay parameter for a peer scoring. It finds a number X
// that satisfies the equation: from*X^(duration/decayInterval)=decayToZero.
// In other words, it finds a decay value for which a scoring will drop to
// the decayToZero value after a specified time.
func decay(from float64, duration time.Duration) float64 {
	return math.Pow(decayToZero/from, 1/(float64(duration)/float64(decayInterval)))
}
