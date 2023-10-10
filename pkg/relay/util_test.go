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

package relay

import (
	"math"
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/chronicleprotocol/oracle-suite/pkg/util/bn"
)

func TestCalculateSpread(t *testing.T) {
	tests := []struct {
		name     string
		new, old *bn.DecFloatPointNumber
		expected float64
	}{
		{"calculateSpread(150, 100)", bn.DecFloatPoint(150), bn.DecFloatPoint(100), 50},
		{"calculateSpread(50, 100)", bn.DecFloatPoint(50), bn.DecFloatPoint(100), 50},
		{"calculateSpread(100, 100)", bn.DecFloatPoint(100), bn.DecFloatPoint(100), 0},
		{"calculateSpread(100, 0)", bn.DecFloatPoint(100), bn.DecFloatPoint(0), math.Inf(1)},
		{"calculateSpread(-100, 0)", bn.DecFloatPoint(-100), bn.DecFloatPoint(0), math.Inf(1)},
		{"calculateSpread(0, 0)", bn.DecFloatPoint(0), bn.DecFloatPoint(0), math.Inf(1)},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := calculateSpread(tt.new, tt.old)
			assert.Equal(t, tt.expected, got)
		})
	}
}

func TestCalculateMedian(t *testing.T) {
	tests := []struct {
		name     string
		prices   []*bn.DecFloatPointNumber
		expected *bn.DecFloatPointNumber
	}{
		{
			"calculateMedian([])",
			[]*bn.DecFloatPointNumber{},
			bn.DecFloatPoint(0),
		},
		{
			"calculateMedian([1, 3, 2])",
			[]*bn.DecFloatPointNumber{
				bn.DecFloatPoint(1),
				bn.DecFloatPoint(3),
				bn.DecFloatPoint(2),
			},
			bn.DecFloatPoint(2),
		},
		{
			"calculateMedian([1, 4, 3, 2])",
			[]*bn.DecFloatPointNumber{
				bn.DecFloatPoint(1),
				bn.DecFloatPoint(4),
				bn.DecFloatPoint(3),
				bn.DecFloatPoint(2),
			},
			bn.DecFloatPoint(2.5),
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := calculateMedian(tt.prices)
			assert.Equal(t, tt.expected, got)
		})
	}
}
