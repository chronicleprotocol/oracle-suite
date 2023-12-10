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

import "github.com/chronicleprotocol/oracle-suite/pkg/util/bn"

const balancerV2Precision = 18

var bnEther = bn.DecFloatPoint(1).Inflate(balancerV2Precision)
var bnZero = bn.DecFloatPoint(0)
var bnOne = bn.DecFloatPoint(1)
var bnTwo = bn.DecFloatPoint(2)

// Complement returns the complement of a value (1 - x), capped to 0 if x is larger than 1.
//
// Useful when computing the complement for values with some level of relative error, as it strips this error and
// prevents intermediate negative values.
func _complementFixed(x *bn.DecFloatPointNumber) *bn.DecFloatPointNumber {
	if x.Cmp(bnEther) < 0 {
		return bnEther.Sub(x)
	}
	return bnZero
}

// _divUp divides the number y up and return the result.
// Reference: https://github.com/balancer/balancer-v2-monorepo/blob/master/pkg/solidity-utils/contracts/math/Math.sol#L102
func _divUp(x, y *bn.DecFloatPointNumber) *bn.DecFloatPointNumber {
	if x.Prec() != 0 || y.Prec() != 0 {
		panic("only available for integer")
	}
	if x.Sign() == 0 {
		return x
	}
	// 1 + (a - 1) / b
	//return x.Sub(bnOne).Add(bnOne)
	return x.Sub(bnOne).DivPrec(y, uint32(x.Prec())).Add(bnOne)
}

// _divDown divides the number y down and return the result.
// Reference: https://github.com/balancer/balancer-v2-monorepo/blob/master/pkg/solidity-utils/contracts/math/Math.sol#L97
func _divDown(x, y *bn.DecFloatPointNumber) *bn.DecFloatPointNumber {
	if x.Prec() != 0 || y.Prec() != 0 {
		panic("only available for integer")
	}
	if x.Sign() == 0 {
		return x
	}
	return x.DivPrec(y, uint32(x.Prec()))
}

// _divUpFixed inflates prec precision and divides the number y up.
// Reference: https://github.com/balancer/balancer-v2-monorepo/blob/master/pkg/solidity-utils/contracts/math/FixedPoint.sol#L83
func _divUpFixed(x, y *bn.DecFloatPointNumber, prec uint8) *bn.DecFloatPointNumber {
	if x.Sign() == 0 {
		return x
	}

	// The traditional divUp formula is:
	// divUp(x, y) := (x + y - 1) / y
	// To avoid intermediate overflow in the addition, we distribute the division and get:
	// divUp(x, y) := (x - 1) / y + 1
	// Note that this requires x != 0, which we already tested for.
	return x.Inflate(prec).Sub(bnOne).DivPrec(y, uint32(x.Prec())).Add(bnOne)
}

func _divUpFixed18(x, y *bn.DecFloatPointNumber) *bn.DecFloatPointNumber {
	return _divUpFixed(x, y, balancerV2Precision)
}

// _divDownFixed inflates prec precision and divides the number y down
// Reference: https://github.com/balancer/balancer-v2-monorepo/blob/master/pkg/solidity-utils/contracts/math/FixedPoint.sol#L74
func _divDownFixed(x, y *bn.DecFloatPointNumber, prec uint8) *bn.DecFloatPointNumber {
	if x.Sign() == 0 {
		return x
	}
	return x.Inflate(prec).DivPrec(y, uint32(x.Prec()))
}

func _divDownFixed18(x, y *bn.DecFloatPointNumber) *bn.DecFloatPointNumber {
	return _divDownFixed(x, y, balancerV2Precision)
}

// _mulDownFixed multiplies the number y and deflates prec precision
// Reference: https://github.com/balancer/balancer-v2-monorepo/blob/master/pkg/solidity-utils/contracts/math/FixedPoint.sol#L50
func _mulDownFixed(x, y *bn.DecFloatPointNumber, prec uint8) *bn.DecFloatPointNumber {
	return x.Mul(y).Deflate(prec)
}

func _mulDownFixed18(x, y *bn.DecFloatPointNumber) *bn.DecFloatPointNumber {
	return _mulDownFixed(x, y, balancerV2Precision)
}

// _mulUpFixed multiplies the number y up and deflates prec precision
// Reference: https://github.com/balancer/balancer-v2-monorepo/blob/master/pkg/solidity-utils/contracts/math/FixedPoint.sol#L57
func _mulUpFixed(x, y *bn.DecFloatPointNumber, prec uint8) *bn.DecFloatPointNumber {
	// The traditional divUp formula is:
	// divUp(x, y) := (x + y - 1) / y
	// To avoid intermediate overflow in the addition, we distribute the division and get:
	// divUp(x, y) := (x - 1) / y + 1
	// Note that this requires x != 0, if x == 0 then the result is zero

	ret := x.Mul(y)
	if ret.Sign() == 0 {
		return ret
	}
	return ret.Sub(bnOne).Deflate(prec).Add(bnOne)
}

func _mulUpFixed18(x, y *bn.DecFloatPointNumber) *bn.DecFloatPointNumber {
	return _mulUpFixed(x, y, balancerV2Precision)
}
