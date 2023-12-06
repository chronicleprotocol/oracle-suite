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
	"math/big"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/chronicleprotocol/oracle-suite/pkg/util/bn"
)

func expectEqualWithError(actual *bn.DecFloatPointNumber, expected *bn.DecFloatPointNumber, error *bn.DecFloatPointNumber) bool {
	acceptedError := expected.Mul(error)
	if acceptedError.Cmp(bn.DecFloatPoint(0)) > 0 {
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
	var MAX_X = bn.DecFloatPoint(new(big.Int).Exp(big.NewInt(2), big.NewInt(255), nil)).Sub(bn.DecFloatPoint(1))
	var MAX_Y = bn.DecFloatPoint(
		new(big.Int).Div(
			new(big.Int).Exp(big.NewInt(2), big.NewInt(254), nil),
			new(big.Int).Exp(big.NewInt(10), big.NewInt(20), nil),
		),
	).Sub(bn.DecFloatPoint(1))

	tests := []struct {
		name     string
		base     *bn.DecFloatPointNumber
		exponent *bn.DecFloatPointNumber
		result   *bn.DecFloatPointNumber
		delta    *bn.DecFloatPointNumber
		error    error
	}{
		{
			name:     "exponent zero, handles base zero",
			base:     bn.DecFloatPoint(0),
			exponent: bn.DecFloatPoint(0),
			result:   bn.DecFloatPoint(1).Inflate(balancerV2Precision),
			delta:    bn.DecFloatPoint(0),
		},
		{
			name:     "exponent zero, handles base one",
			base:     bn.DecFloatPoint(1),
			exponent: bn.DecFloatPoint(0),
			result:   bn.DecFloatPoint(1).Inflate(balancerV2Precision),
			delta:    bn.DecFloatPoint(0),
		},
		{
			name:     "exponent zero, handles base greater than one",
			base:     bn.DecFloatPoint(10),
			exponent: bn.DecFloatPoint(0),
			result:   bn.DecFloatPoint(1).Inflate(balancerV2Precision),
			delta:    bn.DecFloatPoint(0),
		},
		{
			name:     "base zero, handles exponent zero",
			base:     bn.DecFloatPoint(0),
			exponent: bn.DecFloatPoint(0),
			result:   bn.DecFloatPoint(1).Inflate(balancerV2Precision),
			delta:    bn.DecFloatPoint(0),
		},
		{
			name:     "base zero, handles exponent one",
			base:     bn.DecFloatPoint(0),
			exponent: bn.DecFloatPoint(1),
			result:   bn.DecFloatPoint(0),
			delta:    bn.DecFloatPoint(0),
		},
		{
			name:     "base zero, handles exponent greater than one",
			base:     bn.DecFloatPoint(0),
			exponent: bn.DecFloatPoint(10),
			result:   bn.DecFloatPoint(0),
			delta:    bn.DecFloatPoint(0),
		},
		{
			name:     "base one, handles exponent zero",
			base:     bn.DecFloatPoint(1),
			exponent: bn.DecFloatPoint(0),
			result:   bn.DecFloatPoint(1).Inflate(balancerV2Precision),
			delta:    bn.DecFloatPoint(0),
		},
		{
			name:     "base one, handles exponent one",
			base:     bn.DecFloatPoint(1),
			exponent: bn.DecFloatPoint(1),
			result:   bn.DecFloatPoint(1).Inflate(balancerV2Precision),
			delta:    bn.DecFloatPoint(0.000000000001),
		},
		{
			name:     "base one, handles exponent greater than one",
			base:     bn.DecFloatPoint(1),
			exponent: bn.DecFloatPoint(10),
			result:   bn.DecFloatPoint(1).Inflate(balancerV2Precision),
			delta:    bn.DecFloatPoint(0.000000000001),
		},
		{
			name:     "decimals, handles decimals properly",
			base:     bn.DecFloatPoint(2).Inflate(balancerV2Precision),
			exponent: bn.DecFloatPoint(4).Inflate(balancerV2Precision),
			result:   bn.DecFloatPoint(16).Inflate(balancerV2Precision),
			delta:    bn.DecFloatPoint(0.000000000001),
		},
		{
			name:     "max values, cannot handle a base greater than 2^255 - 1",
			base:     MAX_X.Add(bn.DecFloatPoint(1)),
			exponent: bn.DecFloatPoint(1),
			error:    fmt.Errorf("X_OUT_OF_BOUNDS"),
		},
		{
			name:     "max values, cannot handle an exponent greater than (2^254/1e20) - 1",
			base:     bn.DecFloatPoint(1),
			exponent: MAX_Y.Add(bn.DecFloatPoint(1)),
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
	valuesPow4 := []*bn.DecFloatPointNumber{
		bn.DecFloatPoint(0.0007),
		bn.DecFloatPoint(0.0022),
		bn.DecFloatPoint(0.093),
		bn.DecFloatPoint(2.9),
		bn.DecFloatPoint(13.3),
		bn.DecFloatPoint(450.8),
		bn.DecFloatPoint(1550.3339),
		bn.DecFloatPoint(69039.11),
		bn.DecFloatPoint(7834839.432),
		bn.DecFloatPoint(83202933.5433),
		bn.DecFloatPoint(9983838318.4),
		bn.DecFloatPoint(15831567871.1),
	}

	valuesPow2 := append(append([]*bn.DecFloatPointNumber{
		bn.DecFloatPoint(8e-9),
		bn.DecFloatPoint(0.0000013),
		bn.DecFloatPoint(0.000043),
	}, valuesPow4...), []*bn.DecFloatPointNumber{
		bn.DecFloatPoint(8382392893832.1),
		bn.DecFloatPoint(38859321075205.1),
		bn.DecFloatPoint("848205610278492.2383"),
		bn.DecFloatPoint("371328129389320282.3783289"),
	}...)

	valuesPow1 := append(append([]*bn.DecFloatPointNumber{
		bn.DecFloatPoint(1.7e-18),
		bn.DecFloatPoint(1.7e-15),
		bn.DecFloatPoint(1.7e-11),
	}, valuesPow2...), []*bn.DecFloatPointNumber{
		bn.DecFloatPoint("701847104729761867823532.139"),
		bn.DecFloatPoint("175915239864219235419349070.947"),
	}...)

	tests := []struct {
		name   string
		values []*bn.DecFloatPointNumber
		pow    *bn.DecFloatPointNumber
	}{
		{
			name:   "non-fractional pow 1",
			values: valuesPow1,
			pow:    bn.DecFloatPoint(1),
		},
		{
			name:   "non-fractional pow 2",
			values: valuesPow2,
			pow:    bn.DecFloatPoint(2),
		},
		{
			name:   "non-fractional pow 4",
			values: valuesPow4,
			pow:    bn.DecFloatPoint(4),
		},
	}

	for _, tt := range tests {
		for _, x := range tt.values {
			t.Run(tt.name+":"+x.String(), func(t *testing.T) {
				pow := tt.pow
				EXPECTED_RELATIVE_ERROR := bn.DecFloatPoint(1e-14)
				result, err := _pow(x.Inflate(balancerV2Precision), pow.Inflate(balancerV2Precision))
				require.NoError(t, err)
				x2 := x.Inflate(balancerV2Precision)
				pow2 := pow.Inflate(balancerV2Precision)
				assert.True(t, expectEqualWithError(_powDownFixed(x2, pow2, balancerV2Precision), result, EXPECTED_RELATIVE_ERROR))
				assert.True(t, expectEqualWithError(_powUpFixed(x2, pow2, balancerV2Precision), result, EXPECTED_RELATIVE_ERROR))
			})
		}
	}
}
