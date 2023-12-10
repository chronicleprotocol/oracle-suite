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

package origin

import (
	"fmt"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/chronicleprotocol/oracle-suite/pkg/util/bn"
)

func expectEqualWithError(actual *bn.DecFixedPointNumber, expected *bn.DecFixedPointNumber, error *bn.DecFixedPointNumber) bool {
	acceptedError := expected.Mul(error)
	if acceptedError.Cmp(bnZero) > 0 {
		if actual.Cmp(expected.Sub(acceptedError)) < 0 {
			return false
		}
		if actual.Cmp(expected.Add(acceptedError)) > 0 {
			return false
		}
		return true
	} else {
		if actual.Cmp(expected.Sub(acceptedError)) > 0 {
			return false
		}
		if actual.Cmp(expected.Add(acceptedError)) < 0 {
			return false
		}
		return true
	}
}

// Reference: https://github.com/balancer/balancer-v2-monorepo/blob/master/pkg/solidity-utils/test/LogExpMath.test.ts#L9
func TestBalancerV2_ExpLog(t *testing.T) {
	var MAX_X = _powX(2, 255).Sub(bnOne)
	var MAX_Y = _powX(2, 254).DivPrec(_powX(10, 20), 0).Sub(bnOne)

	tests := []struct {
		name     string
		base     *bn.DecFixedPointNumber
		exponent *bn.DecFixedPointNumber
		result   *bn.DecFixedPointNumber
		delta    *bn.DecFixedPointNumber
		error    error
	}{
		{
			name:     "exponent zero, handles base zero",
			base:     bnZero,
			exponent: bnZero,
			result:   bnEther,
			delta:    bnZero,
		},
		{
			name:     "exponent zero, handles base one",
			base:     bnOne,
			exponent: bnZero,
			result:   bnEther,
			delta:    bnZero,
		},
		{
			name:     "exponent zero, handles base greater than one",
			base:     bn.DecFixedPoint(10, 0),
			exponent: bnZero,
			result:   bnEther,
			delta:    bnZero,
		},
		{
			name:     "base zero, handles exponent zero",
			base:     bnZero,
			exponent: bnZero,
			result:   bnEther,
			delta:    bnZero,
		},
		{
			name:     "base zero, handles exponent one",
			base:     bnZero,
			exponent: bnOne,
			result:   bnZero,
			delta:    bnZero,
		},
		{
			name:     "base zero, handles exponent greater than one",
			base:     bnZero,
			exponent: bn.DecFixedPoint(10, 0),
			result:   bnZero,
			delta:    bnZero,
		},
		{
			name:     "base one, handles exponent zero",
			base:     bnOne,
			exponent: bnZero,
			result:   bnEther,
			delta:    bnZero,
		},
		{
			name:     "base one, handles exponent one",
			base:     bnOne,
			exponent: bnOne,
			result:   bnEther,
			delta:    bn.DecFixedPoint(1, 12),
		},
		{
			name:     "base one, handles exponent greater than one",
			base:     bnOne,
			exponent: bn.DecFixedPoint(10, 0),
			result:   bnEther,
			delta:    bn.DecFixedPoint(1, 12),
		},
		{
			name:     "decimals, handles decimals properly",
			base:     _powX(2, balancerV2Precision),
			exponent: _powX(4, balancerV2Precision),
			result:   _powX(16, balancerV2Precision),
			delta:    bn.DecFixedPoint(1, 12),
		},
		{
			name:     "max values, cannot handle a base greater than 2^255 - 1",
			base:     MAX_X.Add(bnOne),
			exponent: bnOne,
			error:    fmt.Errorf("X_OUT_OF_BOUNDS"),
		},
		{
			name:     "max values, cannot handle an exponent greater than (2^254/1e20) - 1",
			base:     bnOne,
			exponent: MAX_Y.Add(bnOne),
			error:    fmt.Errorf("Y_OUT_OF_BOUNDS"),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := _pow(tt.base, tt.exponent)

			if tt.error != nil {
				assert.Equal(t, err, tt.error)
			} else {
				require.NoError(t, err)
				assert.True(t, expectEqualWithError(result, tt.result, tt.delta))
			}
		})
	}
}

// Reference: https://github.com/balancer/balancer-v2-monorepo/blob/master/pkg/solidity-utils/test/FixedPoint.test.ts#L9
func TestBalancerV2_PowUpdownFixed(t *testing.T) {
	valuesPow4 := []*bn.DecFixedPointNumber{
		bn.DecFixedPoint(7, 4),
		bn.DecFixedPoint(22, 4),
		bn.DecFixedPoint(93, 3),
		bn.DecFixedPoint(29, 1),
		bn.DecFixedPoint(133, 1),
		bn.DecFixedPoint(4508, 1),
		bn.DecFixedPoint(15503339, 4),
		bn.DecFixedPoint(6903911, 2),
		bn.DecFixedPoint(7834839432, 3),
		bn.DecFixedPoint(832029335433, 4),
		bn.DecFixedPoint(99838383184, 1),
		bn.DecFixedPoint(158315678711, 1),
	}

	valuesPow2 := append(append([]*bn.DecFixedPointNumber{
		bn.DecFixedPoint(8, 9),
		bn.DecFixedPoint(13, 7),
		bn.DecFixedPoint(43, 6),
	}, valuesPow4...), []*bn.DecFixedPointNumber{
		bn.DecFixedPoint(83823928938321, 1),
		bn.DecFixedPoint(3885932107520511, 1),
		bn.DecFixedPoint("8482056102784922383", 4),
		bn.DecFixedPoint("3713281293893202823783289", 7),
	}...)

	valuesPow1 := append(append([]*bn.DecFixedPointNumber{
		bn.DecFixedPoint(17, 18),
		bn.DecFixedPoint(17, 15),
		bn.DecFixedPoint(17, 11),
	}, valuesPow2...), []*bn.DecFixedPointNumber{
		bn.DecFixedPoint("701847104729761867823532139", 3),
		bn.DecFixedPoint("175915239864219235419349070947", 3),
	}...)

	tests := []struct {
		name   string
		values []*bn.DecFixedPointNumber
		pow    *bn.DecFixedPointNumber
	}{
		{
			name:   "non-fractional pow 1",
			values: valuesPow1,
			pow:    bnOne,
		},
		{
			name:   "non-fractional pow 2",
			values: valuesPow2,
			pow:    bnTwo,
		},
		{
			name:   "non-fractional pow 4",
			values: valuesPow4,
			pow:    bnFour,
		},
	}

	for _, tt := range tests {
		for _, x := range tt.values {
			t.Run(tt.name+":"+x.String(), func(t *testing.T) {
				pow := tt.pow
				EXPECTED_RELATIVE_ERROR := bn.DecFixedPoint(1, 14)
				result, err := _pow(x.Mul(bnEther).SetPrec(0), pow.Mul(bnEther).SetPrec(0))
				require.NoError(t, err)
				x2 := x.Mul(bnEther).SetPrec(0)
				pow2 := pow.Mul(bnEther).SetPrec(0)
				assert.True(t, expectEqualWithError(_powDownFixed(x2, pow2, balancerV2Precision), result, EXPECTED_RELATIVE_ERROR))
				assert.True(t, expectEqualWithError(_powUpFixed(x2, pow2, balancerV2Precision), result, EXPECTED_RELATIVE_ERROR))
			})
		}
	}
}
