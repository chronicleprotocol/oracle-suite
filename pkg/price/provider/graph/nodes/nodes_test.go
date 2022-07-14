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

package nodes

import (
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/chronicleprotocol/oracle-suite/pkg/price/provider"
)

func TestDetectCycles(t *testing.T) {
	p := provider.Pair{Base: "A", Quote: "B"}

	// Non cyclic graph:
	nonCyclic := NewMedianAggregatorNode(p, 0)
	nonCyclicC1 := NewOriginNode(OriginPair{Origin: "a", Pair: p}, 0, 0)
	nonCyclicC2 := NewOriginNode(OriginPair{Origin: "b", Pair: p}, 0, 0)
	nonCyclicC3 := NewMedianAggregatorNode(p, 0)
	nonCyclic.AddChild(nonCyclicC1)
	nonCyclic.AddChild(nonCyclicC2)
	nonCyclic.AddChild(nonCyclicC3)
	nonCyclicC3.AddChild(nonCyclicC1)
	nonCyclicC3.AddChild(nonCyclicC2)

	// Cyclic graph:
	cyclic := NewMedianAggregatorNode(p, 0)
	cyclicC1 := NewOriginNode(OriginPair{Origin: "a", Pair: p}, 0, 0)
	cyclicC2 := NewOriginNode(OriginPair{Origin: "b", Pair: p}, 0, 0)
	cyclicC3 := NewMedianAggregatorNode(p, 0)
	cyclic.AddChild(cyclicC1)
	cyclic.AddChild(cyclicC2)
	cyclic.AddChild(cyclicC3)
	cyclicC3.AddChild(cyclicC1)
	cyclicC3.AddChild(cyclic)

	// Graph with references to the same aggregator nodes:
	r := NewMedianAggregatorNode(p, 0)
	c1 := NewMedianAggregatorNode(p, 0)
	c2 := NewMedianAggregatorNode(p, 0)
	r.AddChild(c1)
	r.AddChild(c2)
	r.AddChild(nonCyclic)
	c1.AddChild(nonCyclic)
	c2.AddChild(nonCyclic)
	c2.AddChild(cyclic)

	assert.Len(t, DetectCycle(nonCyclic), 0)
	assert.Equal(t, []Node{cyclic, cyclicC3}, DetectCycle(cyclic))
	assert.Equal(t, []Node{r, c2, cyclic, cyclicC3}, DetectCycle(r))
}
