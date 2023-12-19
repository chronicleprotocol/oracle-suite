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

	"github.com/chronicleprotocol/oracle-suite/pkg/util/bn"
)

const balancerV2Precision = 18

var bnEther = _powX(10, balancerV2Precision)
var bnZero = bn.DecFixedPoint(0, 0)
var bnOne = bn.DecFixedPoint(1, 0)
var bnTwo = bn.DecFixedPoint(2, 0)  //nolint:unused
var bnFour = bn.DecFixedPoint(4, 0) //nolint:unused

func _powX(x, y int64) *bn.DecFixedPointNumber { //nolint:unparam
	return bn.DecFixedPoint(new(big.Int).Exp(big.NewInt(x), big.NewInt(y), nil), 0)
}

// Complement returns the complement of a value (1 - x), capped to 0 if x is larger than 1.
//
// Useful when computing the complement for values with some level of relative error, as it strips this error and
// prevents intermediate negative values.
func _complementFixed(x *bn.DecFixedPointNumber) *bn.DecFixedPointNumber {
	if x.Cmp(bnEther) < 0 {
		return bnEther.Sub(x)
	}
	return bnZero

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

// _divUpFixed inflates prec precision and divides the number y up.
// Reference: https://github.com/balancer/balancer-v2-monorepo/blob/master/pkg/solidity-utils/contracts/math/FixedPoint.sol#L83
func _divUpFixed(x, y *bn.DecFixedPointNumber, prec uint8) *bn.DecFixedPointNumber {
	if x.Prec() != 0 || y.Prec() != 0 {
		panic("only available for integer")
	}
	if x.Sign() == 0 {
		return x
	}

	inflated := _powX(10, int64(prec))
	// The traditional divUp formula is:
	// divUp(x, y) := (x + y - 1) / y
	// To avoid intermediate overflow in the addition, we distribute the division and get:
	// divUp(x, y) := (x - 1) / y + 1
	// Note that this requires x != 0, which we already tested for.
	return x.Mul(inflated).Sub(bnOne).DivPrec(y, 0).Add(bnOne)
}

func _divUpFixed18(x, y *bn.DecFixedPointNumber) *bn.DecFixedPointNumber {
	return _divUpFixed(x, y, balancerV2Precision)
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

// _mulDownFixed multiplies the number y and deflates prec precision
// Reference: https://github.com/balancer/balancer-v2-monorepo/blob/master/pkg/solidity-utils/contracts/math/FixedPoint.sol#L50
func _mulDownFixed(x, y *bn.DecFixedPointNumber, prec uint8) *bn.DecFixedPointNumber {
	if x.Prec() != 0 || y.Prec() != 0 {
		panic("only available for integer")
	}
	inflated := _powX(10, int64(prec))
	return x.Mul(y).DivPrec(inflated, 0)
}

func _mulDownFixed18(x, y *bn.DecFixedPointNumber) *bn.DecFixedPointNumber {
	return _mulDownFixed(x, y, balancerV2Precision)
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

func _mod(x, y *bn.DecFixedPointNumber) *bn.DecFixedPointNumber {
	if x.Prec() != 0 || y.Prec() != 0 {
		panic("only available for integer")
	}
	if x.Sign() == 0 {
		return x
	}

	z := new(big.Int).Mod(x.RawBigInt(), y.RawBigInt())
	return bn.DecFixedPoint(z, 0)
}

var X_OUT_OF_BOUNDS = fmt.Errorf("X_OUT_OF_BOUNDS")             //nolint:revive,stylecheck
var Y_OUT_OF_BOUNDS = fmt.Errorf("Y_OUT_OF_BOUNDS")             //nolint:revive,stylecheck
var PRODUCT_OUT_OF_BOUNDS = fmt.Errorf("PRODUCT_OUT_OF_BOUNDS") //nolint:revive,stylecheck
var ONE_18 = _powX(10, balancerV2Precision)                     //nolint:revive,stylecheck
var ONE_20 = _powX(10, 20)                                      //nolint:revive,gomnd,stylecheck
var ONE_36 = _powX(10, 36)                                      //nolint:revive,gomnd,stylecheck
var MAX_NATURAL_EXPONENT = bn.DecFixedPoint(130, 0).Mul(ONE_18) //nolint:revive,gomnd,stylecheck
var MIN_NATURAL_EXPONENT = bn.DecFixedPoint(-41, 0).Mul(ONE_18) //nolint:revive,gomnd,stylecheck
var LN_36_LOWER_BOUND = ONE_18.Sub(_powX(10, 17))               //nolint:revive,gomnd,stylecheck
var LN_36_UPPER_BOUND = ONE_18.Add(_powX(10, 17))               //nolint:revive,gomnd,stylecheck
var MAX_EXPONENT_BOUND = _powX(2, 255)                          //nolint:revive,gomnd,stylecheck
var MILD_EXPONENT_BOUND = _divDown(_powX(2, 254), ONE_20)       //nolint:revive,gomnd,stylecheck
var x0 = bn.DecFixedPoint("128000000000000000000", 0)
var a0 = bn.DecFixedPoint("38877084059945950922200000000000000000000000000000000000", 0)
var x1 = bn.DecFixedPoint("64000000000000000000", 0)
var a1 = bn.DecFixedPoint("6235149080811616882910000000", 0)

var x2 = bn.DecFixedPoint("3200000000000000000000", 0)
var a2 = bn.DecFixedPoint("7896296018268069516100000000000000", 0)
var x3 = bn.DecFixedPoint("1600000000000000000000", 0)
var a3 = bn.DecFixedPoint("888611052050787263676000000", 0)
var x4 = bn.DecFixedPoint("800000000000000000000", 0)
var a4 = bn.DecFixedPoint("298095798704172827474000", 0)
var x5 = bn.DecFixedPoint("400000000000000000000", 0)
var a5 = bn.DecFixedPoint("5459815003314423907810", 0)
var x6 = bn.DecFixedPoint("200000000000000000000", 0)
var a6 = bn.DecFixedPoint("738905609893065022723", 0)
var x7 = bn.DecFixedPoint("100000000000000000000", 0)
var a7 = bn.DecFixedPoint("271828182845904523536", 0)
var x8 = bn.DecFixedPoint("50000000000000000000", 0)
var a8 = bn.DecFixedPoint("164872127070012814685", 0)
var x9 = bn.DecFixedPoint("25000000000000000000", 0)
var a9 = bn.DecFixedPoint("128402541668774148407", 0)
var x10 = bn.DecFixedPoint("12500000000000000000", 0)
var a10 = bn.DecFixedPoint("113314845306682631683", 0)
var x11 = bn.DecFixedPoint("6250000000000000000", 0)
var a11 = bn.DecFixedPoint("106449445891785942956", 0)

// _pow calculate an exponentiation (x^y) with unsigned 18 decimal fixed point base and exponent.
// Reference: https://github.com/balancer/balancer-v2-monorepo/blob/master/pkg/solidity-utils/contracts/math/LogExpMath.sol#L93
func _pow(x, y *bn.DecFixedPointNumber) (*bn.DecFixedPointNumber, error) {
	// if (y == 0) {
	//	// We solve the 0^0 indetermination by making it equal one.
	//	return uint256(ONE_18);
	// }
	if y.Cmp(bnZero) == 0 {
		return ONE_18, nil
	}

	// if (x == 0) {
	//	return 0;
	// }
	if x.Cmp(bnZero) == 0 {
		return bnZero, nil
	}

	// Instead of computing x^y directly, we instead rely on the properties of logarithms and exponentiation to
	// arrive at that result. In particular, exp(ln(x)) = x, and ln(x^y) = y * ln(x). This means
	// x^y = exp(y * ln(x)).

	// The ln function takes a signed value, so we need to make sure x fits in the signed 256 bit range.
	// _require(x >> 255 == 0, Errors.X_OUT_OF_BOUNDS);
	// int256 x_int256 = int256(x);
	if x.Cmp(MAX_EXPONENT_BOUND) >= 0 {
		return nil, X_OUT_OF_BOUNDS
	}

	// We will compute y * ln(x) in a single step. Depending on the value of x, we can either use ln or ln_36. In
	// both cases, we leave the division by ONE_18 (due to fixed point multiplication) to the end.

	// This prevents y * ln(x) from overflowing, and at the same time guarantees y fits in the signed 256 bit range.
	// _require(y < MILD_EXPONENT_BOUND, Errors.Y_OUT_OF_BOUNDS);
	// int256 y_int256 = int256(y);
	if y.Cmp(MILD_EXPONENT_BOUND) >= 0 {
		return nil, Y_OUT_OF_BOUNDS
	}

	var logx_times_y *bn.DecFixedPointNumber //nolint:revive,stylecheck
	// if (LN_36_LOWER_BOUND < x_int256 && x_int256 < LN_36_UPPER_BOUND) {
	if LN_36_LOWER_BOUND.Cmp(x) < 0 && x.Cmp(LN_36_UPPER_BOUND) < 0 {
		//	int256 ln_36_x = _ln_36(x_int256);
		ln_36_x := _ln_36(x) //nolint:revive,stylecheck

		// ln_36_x has 36 decimal places, so multiplying by y_int256 isn't as straightforward, since we can't just
		// bring y_int256 to 36 decimal places, as it might overflow. Instead, we perform two 18 decimal
		// multiplications and add the results: one with the first 18 decimals of ln_36_x, and one with the
		// (downscaled) last 18 decimals.
		// logx_times_y = ((ln_36_x / ONE_18) * y_int256 + ((ln_36_x % ONE_18) * y_int256) / ONE_18);
		logx_times_y = ln_36_x.DivPrec(ONE_18, 0).Mul(y).
			Add(_mod(ln_36_x, ONE_18).Mul(y).DivPrec(ONE_18, 0))
	} else {
		// logx_times_y = _ln(x_int256) * y_int256;
		logx_times_y = _ln(x).Mul(y)
	}
	// logx_times_y /= ONE_18;
	logx_times_y = logx_times_y.DivPrec(ONE_18, 0)

	// Finally, we compute exp(y * ln(x)) to arrive at x^y
	// _require(
	//	MIN_NATURAL_EXPONENT <= logx_times_y && logx_times_y <= MAX_NATURAL_EXPONENT,
	//	Errors.PRODUCT_OUT_OF_BOUNDS
	// );
	if logx_times_y.Cmp(MIN_NATURAL_EXPONENT) <= 0 && logx_times_y.Cmp(MAX_NATURAL_EXPONENT) <= 0 {
		return nil, PRODUCT_OUT_OF_BOUNDS
	}

	// return uint256(exp(logx_times_y));
	return _exp(logx_times_y)
}

// _exp is a natural exponentiation (e^x) with signed 18 decimal fixed point exponent.
// Reference: https://github.com/balancer/balancer-v2-monorepo/blob/master/pkg/solidity-utils/contracts/math/LogExpMath.sol#L146
func _exp(x *bn.DecFixedPointNumber) (*bn.DecFixedPointNumber, error) {
	// _require(x >= MIN_NATURAL_EXPONENT && x <= MAX_NATURAL_EXPONENT, Errors.INVALID_EXPONENT);
	if x.Cmp(MIN_NATURAL_EXPONENT) < 0 || x.Cmp(MAX_NATURAL_EXPONENT) > 0 {
		return nil, fmt.Errorf("INVALID_EXPONENT")
	}

	// if (x < 0) {
	if x.Cmp(bnZero) < 0 {
		// We only handle positive exponents: e^(-x) is computed as 1 / e^x. We can safely make x positive since it
		// fits in the signed 256 bit range (as it is larger than MIN_NATURAL_EXPONENT).
		// Fixed point division requires multiplying by ONE_18.
		// return ((ONE_18 * ONE_18) / exp(-x));
		ret, err := _exp(x.Neg())
		if err != nil {
			return nil, err
		}
		return ONE_18.Mul(ONE_18).DivPrec(ret, 0), nil
	}

	// First, we use the fact that e^(x+y) = e^x * e^y to decompose x into a sum of powers of two, which we call x_n,
	// where x_n == 2^(7 - n), and e^x_n = a_n has been precomputed. We choose the first x_n, x0, to equal 2^7
	// because all larger powers are larger than MAX_NATURAL_EXPONENT, and therefore not present in the
	// decomposition.
	// At the end of this process we will have the product of all e^x_n = a_n that apply, and the remainder of this
	// decomposition, which will be lower than the smallest x_n.
	// exp(x) = k_0 * a_0 * k_1 * a_1 * ... + k_n * a_n * exp(remainder), where each k_n equals either 0 or 1.
	// We mutate x by subtracting x_n, making it the remainder of the decomposition.

	// The first two a_n (e^(2^7) and e^(2^6)) are too large if stored as 18 decimal numbers, and could cause
	// intermediate overflows. Instead we store them as plain integers, with 0 decimals.
	// Additionally, x0 + x1 is larger than MAX_NATURAL_EXPONENT, which means they will not both be present in the
	// decomposition.

	// For each x_n, we test if that term is present in the decomposition (if x is larger than it), and if so deduct
	// it and compute the accumulated product.

	var firstAN *bn.DecFixedPointNumber
	switch {
	case x.Cmp(x0) >= 0:
		x = x.Sub(x0)
		firstAN = a0
	case x.Cmp(x1) >= 0:
		x = x.Sub(x1)
		firstAN = a1
	default:
		firstAN = bnOne
	}

	// We now transform x into a 20 decimal fixed point number, to have enhanced precision when computing the
	// smaller terms.
	// x *= 100;
	x = x.Mul(bn.DecFixedPoint(100, 0))

	// `product` is the accumulated product of all a_n (except a0 and a1), which starts at 20 decimal fixed point
	// one. Recall that fixed point multiplication requires dividing by ONE_20.
	// int256 product = ONE_20;
	product := ONE_20

	// if (x >= x2) {
	//	x -= x2;
	//	product = (product * a2) / ONE_20;
	// }
	if x.Cmp(x2) >= 0 {
		x = x.Sub(x2)
		product = product.Mul(a2).DivPrec(ONE_20, 0)
	}
	if x.Cmp(x3) >= 0 {
		x = x.Sub(x3)
		product = product.Mul(a3).DivPrec(ONE_20, 0)
	}
	if x.Cmp(x4) >= 0 {
		x = x.Sub(x4)
		product = product.Mul(a4).DivPrec(ONE_20, 0)
	}
	if x.Cmp(x5) >= 0 {
		x = x.Sub(x5)
		product = product.Mul(a5).DivPrec(ONE_20, 0)
	}
	if x.Cmp(x6) >= 0 {
		x = x.Sub(x6)
		product = product.Mul(a6).DivPrec(ONE_20, 0)
	}
	if x.Cmp(x7) >= 0 {
		x = x.Sub(x7)
		product = product.Mul(a7).DivPrec(ONE_20, 0)
	}
	if x.Cmp(x8) >= 0 {
		x = x.Sub(x8)
		product = product.Mul(a8).DivPrec(ONE_20, 0)
	}
	if x.Cmp(x9) >= 0 {
		x = x.Sub(x9)
		product = product.Mul(a9).DivPrec(ONE_20, 0)
	}

	// x10 and x11 are unnecessary here since we have high enough precision already.

	// Now we need to compute e^x, where x is small (in particular, it is smaller than x9). We use the Taylor series
	// expansion for e^x: 1 + x + (x^2 / 2!) + (x^3 / 3!) + ... + (x^n / n!).

	// int256 seriesSum = ONE_20; // The initial one in the sum, with 20 decimal places.
	// int256 term; // Each term in the sum, where the nth term is (x^n / n!).
	var seriesSum = ONE_20
	var term = x                    // The first term is simply x.
	seriesSum = seriesSum.Add(term) // seriesSum += term;

	// Each term (x^n / n!) equals the previous one times x, divided by n. Since x is a fixed point number,
	// multiplying by it requires dividing by ONE_20, but dividing by the non-fixed point n values does not.

	// term = ((term * x) / ONE_20) / 2
	// seriesSum += term
	for i := 2; i <= 12; i++ {
		term = term.Mul(x).DivPrec(ONE_20, 0).DivPrec(bn.DecFixedPoint(i, 0), 0)
		seriesSum = seriesSum.Add(term)
	}

	// 12 Taylor terms are sufficient for 18 decimal precision.

	// We now have the first a_n (with no decimals), and the product of all other a_n present, and the Taylor
	// approximation of the exponentiation of the remainder (both with 20 decimals). All that remains is to multiply
	// all three (one 20 decimal fixed point multiplication, dividing by ONE_20, and one integer multiplication),
	// and then drop two digits to return an 18 decimal value.

	// return (((product * seriesSum) / ONE_20) * firstAN) / 100;
	return product.Mul(seriesSum).DivPrec(ONE_20, 0).Mul(firstAN).DivPrec(bn.DecFixedPoint(100, 0), 0), nil
}

// _ln is an internal natural logarithm (ln(a)) with signed 18 decimal fixed point argument.
// Reference: https://github.com/balancer/balancer-v2-monorepo/blob/master/pkg/solidity-utils/contracts/math/LogExpMath.sol#L326
func _ln(a *bn.DecFixedPointNumber) *bn.DecFixedPointNumber {
	// if (a < ONE_18) {
	if a.Cmp(ONE_18) < 0 {
		// Since ln(a^k) = k * ln(a), we can compute ln(a) as ln(a) = ln((1/a)^(-1)) = - ln((1/a)). If a is less
		// than one, 1/a will be greater than one, and this if statement will not be entered in the recursive call.
		// Fixed point division requires multiplying by ONE_18.
		// return (-_ln((ONE_18 * ONE_18) / a));
		return _ln(ONE_18.Mul(ONE_18).DivPrec(a, 0)).Neg()
	}

	// First, we use the fact that ln^(a * b) = ln(a) + ln(b) to decompose ln(a) into a sum of powers of two, which
	// we call x_n, where x_n == 2^(7 - n), which are the natural logarithm of precomputed quantities a_n (that is,
	// ln(a_n) = x_n). We choose the first x_n, x0, to equal 2^7 because the exponential of all larger powers cannot
	// be represented as 18 fixed point decimal numbers in 256 bits, and are therefore larger than a.
	// At the end of this process we will have the sum of all x_n = ln(a_n) that apply, and the remainder of this
	// decomposition, which will be lower than the smallest a_n.
	// ln(a) = k_0 * x_0 + k_1 * x_1 + ... + k_n * x_n + ln(remainder), where each k_n equals either 0 or 1.
	// We mutate a by subtracting a_n, making it the remainder of the decomposition.

	// For reasons related to how `exp` works, the first two a_n (e^(2^7) and e^(2^6)) are not stored as fixed point
	// numbers with 18 decimals, but instead as plain integers with 0 decimals, so we need to multiply them by
	// ONE_18 to convert them to fixed point.
	// For each a_n, we test if that term is present in the decomposition (if a is larger than it), and if so divide
	// by it and compute the accumulated sum.

	// int256 sum = 0;
	var sum = bnZero
	// if (a >= a0 * ONE_18) {
	if a.Cmp(a0.Mul(ONE_18)) >= 0 {
		// a /= a0; // Integer, not fixed point division
		a = a.DivPrec(a0, 0)
		// sum += x0;
		sum = sum.Add(x0)
	}

	// if (a >= a1 * ONE_18) {
	if a.Cmp(a1.Mul(ONE_18)) >= 0 {
		// a /= a1; // Integer, not fixed point division
		a = a.DivPrec(a1, 0)
		// sum += x1;
		sum = sum.Add(x1)
	}

	// All other a_n and x_n are stored as 20 digit fixed point numbers, so we convert the sum and a to this format.
	// sum *= 100;
	// a *= 100;
	sum = sum.Mul(bn.DecFixedPoint(100, 0))
	a = a.Mul(bn.DecFixedPoint(100, 0))

	// Because further a_n are  20 digit fixed point numbers, we multiply by ONE_20 when dividing by them.

	// if (a >= a2) {
	//	a = (a * ONE_20) / a2;
	//	sum += x2;
	// }
	if a.Cmp(a2) >= 0 {
		a = a.Mul(ONE_20).DivPrec(a2, 0)
		sum = sum.Add(x2)
	}

	// if (a >= a3) {
	//	a = (a * ONE_20) / a3;
	//	sum += x3;
	// }
	if a.Cmp(a3) >= 0 {
		a = a.Mul(ONE_20).DivPrec(a3, 0)
		sum = sum.Add(x3)
	}

	// if (a >= a4) {
	//	a = (a * ONE_20) / a4;
	//	sum += x4;
	// }
	if a.Cmp(a4) >= 0 {
		a = a.Mul(ONE_20).DivPrec(a4, 0)
		sum = sum.Add(x4)
	}

	// if (a >= a5) {
	//	a = (a * ONE_20) / a4;
	//	sum += x5;
	// }
	if a.Cmp(a5) >= 0 {
		a = a.Mul(ONE_20).DivPrec(a5, 0)
		sum = sum.Add(x5)
	}

	// if (a >= a6) {
	//	a = (a * ONE_20) / a6;
	//	sum += x6;
	// }
	if a.Cmp(a6) >= 0 {
		a = a.Mul(ONE_20).DivPrec(a6, 0)
		sum = sum.Add(x6)
	}

	// if (a >= a7) {
	//	a = (a * ONE_20) / a7;
	//	sum += x7;
	// }
	if a.Cmp(a7) >= 0 {
		a = a.Mul(ONE_20).DivPrec(a7, 0)
		sum = sum.Add(x7)
	}

	// if (a >= a8) {
	//	a = (a * ONE_20) / a8;
	//	sum += x8;
	// }
	if a.Cmp(a8) >= 0 {
		a = a.Mul(ONE_20).DivPrec(a8, 0)
		sum = sum.Add(x8)
	}

	// if (a >= a9) {
	//	a = (a * ONE_20) / a9;
	//	sum += x9;
	// }
	if a.Cmp(a9) >= 0 {
		a = a.Mul(ONE_20).DivPrec(a9, 0)
		sum = sum.Add(x9)
	}

	// if (a >= a10) {
	//	a = (a * ONE_20) / a10;
	//	sum += x10;
	// }
	if a.Cmp(a10) >= 0 {
		a = a.Mul(ONE_20).DivPrec(a10, 0)
		sum = sum.Add(x10)
	}

	// if (a >= a11) {
	//	a = (a * ONE_20) / a11;
	//	sum += x11;
	// }
	if a.Cmp(a11) >= 0 {
		a = a.Mul(ONE_20).DivPrec(a11, 0)
		sum = sum.Add(x11)
	}

	// a is now a small number (smaller than a_11, which roughly equals 1.06). This means we can use a Taylor series
	// that converges rapidly for values of `a` close to one - the same one used in ln_36.
	// Let z = (a - 1) / (a + 1).
	// ln(a) = 2 * (z + z^3 / 3 + z^5 / 5 + z^7 / 7 + ... + z^(2 * n + 1) / (2 * n + 1))

	// Recall that 20 digit fixed point division requires multiplying by ONE_20, and multiplication requires
	// division by ONE_20.
	// int256 z = ((a - ONE_20) * ONE_20) / (a + ONE_20);
	z := a.Sub(ONE_20).Mul(ONE_20).DivPrec(a.Add(ONE_20), 0)
	// int256 z_squared = (z * z) / ONE_20;
	z_squared := z.Mul(z).DivPrec(ONE_20, 0) //nolint:revive,stylecheck

	// num is the numerator of the series: the z^(2 * n + 1) term
	// int256 num = z;
	num := z

	// seriesSum holds the accumulated sum of each term in the series, starting with the initial z
	// int256 seriesSum = num;
	seriesSum := num

	// In each step, the numerator is multiplied by z^2
	for i := 3; i <= 11; i += 2 {
		// num = (num * z_squared) / ONE_20;
		// seriesSum += num / 3;
		num = num.Mul(z_squared).DivPrec(ONE_20, 0)
		seriesSum = seriesSum.Add(num.DivPrec(bn.DecFixedPoint(i, 0), 0))
	}

	// 6 Taylor terms are sufficient for 36 decimal precision.

	// Finally, we multiply by 2 (non fixed point) to compute ln(remainder)
	// seriesSum *= 2;
	seriesSum = seriesSum.Mul(bn.DecFixedPoint(2, 0))

	// We now have the sum of all x_n present, and the Taylor approximation of the logarithm of the remainder (both
	// with 20 decimals). All that remains is to sum these two, and then drop two digits to return a 18 decimal
	// value.
	// return (sum + seriesSum) / 100;
	return sum.Add(seriesSum).DivPrec(bn.DecFixedPoint(100, 0), 0)
}

// _ln36 is an internal high precision (36 decimal places) natural logarithm (ln(x)) with signed 18 decimal fixed point argument,
// for x close to one.
// Reference: https://github.com/balancer/balancer-v2-monorepo/blob/master/pkg/solidity-utils/contracts/math/LogExpMath.sol#L466
func _ln_36(x *bn.DecFixedPointNumber) *bn.DecFixedPointNumber { //nolint:revive,stylecheck
	// Since ln(1) = 0, a value of x close to one will yield a very small result, which makes using 36 digits
	// worthwhile.

	// First, we transform x to a 36 digit fixed point value.
	x = x.Mul(ONE_18)

	// We will use the following Taylor expansion, which converges very rapidly. Let z = (x - 1) / (x + 1).
	// ln(x) = 2 * (z + z^3 / 3 + z^5 / 5 + z^7 / 7 + ... + z^(2 * n + 1) / (2 * n + 1))

	// Recall that 36 digit fixed point division requires multiplying by ONE_36, and multiplication requires
	// division by ONE_36.
	z := x.Sub(ONE_36).Mul(ONE_36).DivPrec(x.Add(ONE_36), 0)
	z_squared := z.Mul(z).DivPrec(ONE_36, 0) //nolint:revive,stylecheck

	// num is the numerator of the series: the z^(2 * n + 1) term
	var num = z

	// seriesSum holds the accumulated sum of each term in the series, starting with the initial z
	var seriesSum = num

	// In each step, the numerator is multiplied by z^2
	for i := 3; i <= 15; i += 2 {
		// num = (num * z_squared) / ONE_36;
		// seriesSum += num / 3;
		num = num.Mul(z_squared).DivPrec(ONE_36, 0)
		seriesSum = seriesSum.Add(num.DivPrec(bn.DecFixedPoint(i, 0), 0))
	}

	// 8 Taylor terms are sufficient for 36 decimal precision.

	// All that remains is multiplying by 2 (non fixed point).
	return seriesSum.Mul(bn.DecFixedPoint(2, 0))
}

// PowUpFixed returns x^y, assuming both are fixed point numbers, rounding up.
// The result is guaranteed to not be below the true value (that is, the error function expected - actual is always negative).
// Reference: https://github.com/balancer/balancer-v2-monorepo/blob/master/pkg/solidity-utils/contracts/math/FixedPoint.sol#L132
func _powUpFixed(x, y *bn.DecFixedPointNumber, prec uint8) *bn.DecFixedPointNumber {
	// Optimize for when y equals 1.0, 2.0 or 4.0, as those are very simple to implement and occur often in 50/50
	// and 80/20 Weighted Pools
	one := bn.DecFixedPoint(1, 0).Mul(_powX(10, balancerV2Precision))
	two := bn.DecFixedPoint(2, 0).Mul(_powX(10, balancerV2Precision))
	four := bn.DecFixedPoint(4, 0).Mul(_powX(10, balancerV2Precision))

	const MAX_POW_RELATIVE_ERROR = 10000 //nolint:revive,stylecheck

	switch {
	case y.Cmp(one) == 0:
		return x
	case y.Cmp(two) == 0:
		return _mulUpFixed(x, x, prec)
	case y.Cmp(four) == 0:
		square := _mulUpFixed(x, x, prec)
		return _mulUpFixed(square, square, prec)
	default:
		//	uint256 raw = LogExpMath.pow(x, y);
		//	uint256 maxError = add(mulUp(raw, MAX_POW_RELATIVE_ERROR), 1);
		raw, _ := _pow(x, y)
		// uint256 internal constant MAX_POW_RELATIVE_ERROR = 10000; // 10^(-14)
		maxPowRelativeError := bn.DecFixedPoint(MAX_POW_RELATIVE_ERROR, 0)
		//	uint256 maxError = add(mulUp(raw, MAX_POW_RELATIVE_ERROR), 1);
		maxError := _mulUpFixed(raw, maxPowRelativeError, prec).Add(one)
		return raw.Add(maxError)
	}
}

func _powUpFixed18(x, y *bn.DecFixedPointNumber) *bn.DecFixedPointNumber {
	return _powUpFixed(x, y, balancerV2Precision)
}

// PowDownFixed returns x^y, assuming both are fixed point numbers, rounding down.
// The result is guaranteed to not be above the true value (that is, the error function expected - actual is always positive).
// Reference: https://github.com/balancer/balancer-v2-monorepo/blob/master/pkg/solidity-utils/contracts/math/FixedPoint.sol#L106
func _powDownFixed(x, y *bn.DecFixedPointNumber, prec uint8) *bn.DecFixedPointNumber { //nolint:unused
	// Optimize for when y equals 1.0, 2.0 or 4.0, as those are very simple to implement and occur often in 50/50
	// and 80/20 Weighted Pools
	one := bn.DecFixedPoint(1, 0).Mul(_powX(10, balancerV2Precision))
	two := bn.DecFixedPoint(2, 0).Mul(_powX(10, balancerV2Precision))
	four := bn.DecFixedPoint(4, 0).Mul(_powX(10, balancerV2Precision))

	const MAX_POW_RELATIVE_ERROR = 10000 //nolint:revive,stylecheck

	switch {
	case y.Cmp(one) == 0:
		return x
	case y.Cmp(two) == 0:
		return _mulDownFixed(x, x, prec)
	case y.Cmp(four) == 0:
		square := _mulDownFixed(x, x, prec)
		return _mulDownFixed(square, square, prec)
	default:
		//	uint256 raw = LogExpMath.pow(x, y);
		//	uint256 maxError = add(mulUp(raw, MAX_POW_RELATIVE_ERROR), 1);
		raw, _ := _pow(x, y)
		// uint256 internal constant MAX_POW_RELATIVE_ERROR = 10000; // 10^(-14)
		maxPowRelativeError := bn.DecFixedPoint(MAX_POW_RELATIVE_ERROR, 0)
		maxError := _mulUpFixed(raw, maxPowRelativeError, prec).Add(one)
		if raw.Cmp(maxError) < 0 {
			return bnZero
		}
		return raw.Sub(maxError)
	}
}

func _powDownFixed18(x, y *bn.DecFixedPointNumber) *bn.DecFixedPointNumber { //nolint:unused
	return _powDownFixed(x, y, balancerV2Precision)
}
