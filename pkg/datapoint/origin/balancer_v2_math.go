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
	"math/big"

	"github.com/chronicleprotocol/oracle-suite/pkg/util/bn"
)

const balancerV2Precision = 18

var bnZero = bn.DecFixedPoint(0, 0)
var bnOne = bn.DecFixedPoint(1, 0)
var bnTwo = bn.DecFixedPoint(2, 0)

func _powX(x, y int64) *bn.DecFixedPointNumber { //nolint:unparam
	return bn.DecFixedPoint(new(big.Int).Exp(big.NewInt(x), big.NewInt(y), nil), 0)
}

// _divUp divides the number y up and return the result.
// Reference: https://github.com/balancer/balancer-v2-monorepo/blob/master/pkg/solidity-utils/contracts/math/Math.sol#L102
func _divUp(x, y *bn.DecFixedPointNumber) *bn.DecFixedPointNumber {
	if x.Prec() != 0 || y.Prec() != 0 {
		panic("only available for integer")
	}
	if x.Sign() == 0 {
		return x
	}
	// 1 + (a - 1) / b
	return x.Sub(bnOne).DivPrec(y, 0).Add(bnOne)
}

// _divDown divides the number y down and return the result.
// Reference: https://github.com/balancer/balancer-v2-monorepo/blob/master/pkg/solidity-utils/contracts/math/Math.sol#L97
func _divDown(x, y *bn.DecFixedPointNumber) *bn.DecFixedPointNumber {
	if x.Prec() != 0 || y.Prec() != 0 {
		panic("only available for integer")
	}
	if x.Sign() == 0 {
		return x
	}
	return x.DivPrec(y, 0)
}

// _divDownFixed inflates prec precision and divides the number y down
// Reference: https://github.com/balancer/balancer-v2-monorepo/blob/master/pkg/solidity-utils/contracts/math/FixedPoint.sol#L74
func _divDownFixed(x, y *bn.DecFixedPointNumber, prec uint8) *bn.DecFixedPointNumber {
	if x.Prec() != 0 || y.Prec() != 0 {
		panic("only available for integer")
	}
	if x.Sign() == 0 {
		return x
	}
	inflated := _powX(10, int64(prec))
	return x.Mul(inflated).DivPrec(y, 0)
}

func _divDownFixed18(x, y *bn.DecFixedPointNumber) *bn.DecFixedPointNumber {
	return _divDownFixed(x, y, balancerV2Precision)
}

// _mulUpFixed multiplies the number y up and deflates prec precision
// Reference: https://github.com/balancer/balancer-v2-monorepo/blob/master/pkg/solidity-utils/contracts/math/FixedPoint.sol#L57
func _mulUpFixed(x, y *bn.DecFixedPointNumber, prec uint8) *bn.DecFixedPointNumber {
	if x.Prec() != 0 || y.Prec() != 0 {
		panic("only available for integer")
	}

	// The traditional divUp formula is:
	// divUp(x, y) := (x + y - 1) / y
	// To avoid intermediate overflow in the addition, we distribute the division and get:
	// divUp(x, y) := (x - 1) / y + 1
	// Note that this requires x != 0, if x == 0 then the result is zero

	ret := x.Mul(y)
	if ret.Sign() == 0 {
		return ret
	}
	inflated := _powX(10, int64(prec))
	return ret.Sub(bnOne).DivPrec(inflated, 0).Add(bnOne)
}

func _mulUpFixed18(x, y *bn.DecFixedPointNumber) *bn.DecFixedPointNumber {
	return _mulUpFixed(x, y, balancerV2Precision)
}
