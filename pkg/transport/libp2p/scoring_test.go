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
	"math"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestScoreParams_calculate(t *testing.T) {
	p := &scoreParams{
		p1Score:              100,
		p2Score:              200,
		p3Score:              -300,
		p3bScore:             -400,
		p4Score:              -500,
		p1Length:             5 * time.Minute,
		p2Length:             10 * time.Minute,
		p3Length:             15 * time.Minute,
		p3bLength:            20 * time.Minute,
		p4Length:             25 * time.Minute,
		minMessagesPerSecond: 1,
		maxMessagesPerSecond: 10,
		maxInvalidMessages:   5,
	}
	pc, err := (p).calculate()
	require.NoError(t, err)

	// P₁
	assert.InDelta(t, p.p1Length.Seconds(), pc.TimeInMeshCap, 0.01)
	assert.InDelta(t, p.p1Score, pc.TimeInMeshCap*pc.TimeInMeshWeight, 0.01)
	// P₂
	assert.InDelta(t, p.p2Length.Seconds()/decayInterval.Seconds(), math.Log(decayToZero/pc.FirstMessageDeliveriesCap)/math.Log(pc.FirstMessageDeliveriesDecay), 0.01)
	assert.InDelta(t, p.p2Score, pc.FirstMessageDeliveriesCap*pc.FirstMessageDeliveriesWeight, 0.01)
	// P₃
	assert.InDelta(t, p.p3Length.Seconds()/decayInterval.Seconds(), math.Log(decayToZero/pc.MeshMessageDeliveriesThreshold)/math.Log(pc.MeshMessageDeliveriesDecay), 0.01)
	assert.InDelta(t, p.p3Score, pc.MeshMessageDeliveriesThreshold*pc.MeshMessageDeliveriesThreshold*pc.MeshMessageDeliveriesWeight, 0.01)
	// P₃b
	assert.InDelta(t, p.p3bLength.Seconds()/decayInterval.Seconds(), math.Log(decayToZero/pc.MeshMessageDeliveriesThreshold)/math.Log(pc.MeshFailurePenaltyDecay), 0.01)
	assert.InDelta(t, p.p3bScore, pc.MeshMessageDeliveriesThreshold*pc.MeshMessageDeliveriesThreshold*pc.MeshFailurePenaltyWeight, 0.01)
	// P₄
	assert.InDelta(t, p.p4Length.Seconds()/decayInterval.Seconds(), math.Log(decayToZero/p.maxInvalidMessages)/math.Log(pc.InvalidMessageDeliveriesDecay), 0.01)
	assert.InDelta(t, p.p4Score, p.maxInvalidMessages*p.maxInvalidMessages*pc.InvalidMessageDeliveriesWeight, 0.01)

	assert.InDelta(t, p.minMessagesPerSecond, pc.MeshMessageDeliveriesThreshold/p.p3Length.Seconds(), 0.01)
	assert.InDelta(t, p.maxMessagesPerSecond, pc.MeshMessageDeliveriesCap/p.p3Length.Seconds(), 0.01)
	assert.InDelta(t, p.maxInvalidMessages, decayToZero*math.Pow(pc.InvalidMessageDeliveriesDecay, p.p4Length.Seconds()/decayInterval.Seconds()*-1), 0.01)
}
