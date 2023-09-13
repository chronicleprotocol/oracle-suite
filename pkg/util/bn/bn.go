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

package bn

import (
	"math"
	"math/big"
	"strings"

	"golang.org/x/exp/constraints"
)

var (
	intZero   = big.NewInt(0)
	intOne    = big.NewInt(1)
	intTen    = big.NewInt(10)
	floatHalf = big.NewFloat(0.5) //nolint:gomnd
	floatOne  = big.NewFloat(1)
)

func convertIntToDecFloatPoint(x *IntNumber) *DecFloatPointNumber {
	return &DecFloatPointNumber{x: convertIntToDecFixedPoint(x, 0)}
}

func convertFloatToDecFloatPoint(x *FloatNumber) *DecFloatPointNumber {
	prec := floatDecimalPrecision(x)
	if prec < 0 {
		return nil
	}
	if prec > math.MaxUint8 {
		prec = math.MaxUint8
	}
	return &DecFloatPointNumber{x: convertFloatToDecFixedPoint(x, uint8(prec))}
}

func convertDecFixedPointToDecFloatPoint(x *DecFixedPointNumber) *DecFloatPointNumber {
	return &DecFloatPointNumber{x: x}
}

func convertBigIntToDecFloatPoint(x *big.Int) *DecFloatPointNumber {
	return &DecFloatPointNumber{x: convertBigIntToDecFixedPoint(x, 0)}
}

func convertBigFloatToDecFloatPoint(x *big.Float) *DecFloatPointNumber {
	if x == nil || x.IsInf() {
		return nil
	}
	prec := bigFloatDecimalPrecision(x)
	if prec < 0 {
		return nil
	}
	if prec > math.MaxUint8 {
		prec = math.MaxUint8
	}
	return &DecFloatPointNumber{x: convertBigFloatToDecFixedPoint(x, uint8(prec))}
}

func convertInt64ToDecFloatPoint(x int64) *DecFloatPointNumber {
	return &DecFloatPointNumber{x: convertInt64ToDecFixedPoint(x, 0)}
}

func convertUint64ToDecFloatPoint(x uint64) *DecFloatPointNumber {
	return &DecFloatPointNumber{x: convertUint64ToDecFixedPoint(x, 0)}
}

func convertFloat64ToDecFloatPoint(x float64) *DecFloatPointNumber {
	if math.IsInf(x, 0) || math.IsNaN(x) {
		return nil
	}
	prec := float64DecimalPrecision(x)
	if prec < 0 {
		return nil
	}
	if prec > math.MaxUint8 {
		prec = math.MaxUint8
	}
	fixed := convertFloat64ToDecFixedPoint(x, uint8(prec))
	if fixed == nil {
		return nil
	}
	return &DecFloatPointNumber{x: fixed}
}

func convertStringToDecFloatPoint(x string) *DecFloatPointNumber {
	prec := stringNumberDecimalPrecision(x)
	if prec < 0 {
		return nil
	}
	if prec > math.MaxUint8 {
		prec = math.MaxUint8
	}
	fixed := convertStringToDecFixedPoint(x, uint8(prec))
	if fixed == nil {
		return nil
	}
	return &DecFloatPointNumber{x: fixed}
}

func convertIntToDecFixedPoint(x *IntNumber, n uint8) *DecFixedPointNumber {
	return &DecFixedPointNumber{x: x.Mul(pow10(n)).BigInt(), prec: n}
}

func convertFloatToDecFixedPoint(x *FloatNumber, n uint8) *DecFixedPointNumber {
	i := bigFloatToBigInt(new(big.Float).Mul(x.BigFloat(), new(big.Float).SetInt(pow10(n))))
	return &DecFixedPointNumber{x: i, prec: n}
}

func convertDecFloatPointToDecFixedPoint(x *DecFloatPointNumber, n uint8) *DecFixedPointNumber {
	return x.x.SetPrecision(n)
}

func convertBigIntToDecFixedPoint(x *big.Int, n uint8) *DecFixedPointNumber {
	return &DecFixedPointNumber{x: new(big.Int).Mul(x, pow10(n)), prec: n}
}

func convertBigFloatToDecFixedPoint(x *big.Float, n uint8) *DecFixedPointNumber {
	i := bigFloatToBigInt(new(big.Float).Mul(x, new(big.Float).SetInt(pow10(n))))
	return &DecFixedPointNumber{x: i, prec: n}
}

func convertInt64ToDecFixedPoint(x int64, n uint8) *DecFixedPointNumber {
	return &DecFixedPointNumber{x: new(big.Int).Mul(new(big.Int).SetInt64(x), pow10(n)), prec: n}
}

func convertUint64ToDecFixedPoint(x uint64, n uint8) *DecFixedPointNumber {
	return &DecFixedPointNumber{x: new(big.Int).Mul(new(big.Int).SetUint64(x), pow10(n)), prec: n}
}

func convertFloat64ToDecFixedPoint(x float64, n uint8) *DecFixedPointNumber {
	if math.IsInf(x, 0) || math.IsNaN(x) {
		return nil
	}
	i := bigFloatToBigInt(new(big.Float).Mul(big.NewFloat(x), new(big.Float).SetInt(pow10(n))))
	return &DecFixedPointNumber{x: i, prec: n}
}

func convertStringToDecFixedPoint(x string, n uint8) *DecFixedPointNumber {
	if f, ok := new(big.Float).SetString(x); ok {
		i := bigFloatToBigInt(new(big.Float).Mul(f, new(big.Float).SetInt(pow10(n))))
		return &DecFixedPointNumber{x: i, prec: n}
	}
	return nil
}

func convertIntToFloat(x *IntNumber) *FloatNumber {
	return &FloatNumber{x: x.BigFloat()}
}

func convertDecFixedPointToFloat(x *DecFixedPointNumber) *FloatNumber {
	return &FloatNumber{x: x.BigFloat()}
}

func convertDecFloatPointToFloat(x *DecFloatPointNumber) *FloatNumber {
	return &FloatNumber{x: x.BigFloat()}
}

func convertBigIntToFloat(x *big.Int) *FloatNumber {
	return &FloatNumber{x: new(big.Float).SetInt(x)}
}

func convertBigFloatToFloat(x *big.Float) *FloatNumber {
	return &FloatNumber{x: x}
}

func convertInt64ToFloat(x int64) *FloatNumber {
	return &FloatNumber{x: new(big.Float).SetInt64(x)}
}

func convertUint64ToFloat(x uint64) *FloatNumber {
	return &FloatNumber{x: new(big.Float).SetUint64(x)}
}

func convertFloat64ToFloat(x float64) *FloatNumber {
	return &FloatNumber{x: big.NewFloat(x)}
}

func convertStringToFloat(x string) *FloatNumber {
	if f, ok := new(big.Float).SetString(x); ok {
		return &FloatNumber{x: f}
	}
	return nil
}

func convertFloatToInt(x *FloatNumber) *IntNumber {
	return &IntNumber{x: x.BigInt()}
}

func convertDecFixedPointToInt(x *DecFixedPointNumber) *IntNumber {
	return &IntNumber{x: x.BigInt()}
}

func convertDecFloatPointToInt(x *DecFloatPointNumber) *IntNumber {
	return &IntNumber{x: x.BigInt()}
}

func convertBigIntToInt(x *big.Int) *IntNumber {
	return &IntNumber{x: x}
}

func convertBigFloatToInt(x *big.Float) *IntNumber {
	return &IntNumber{x: bigFloatToBigInt(x)}
}

func convertInt64ToInt(x int64) *IntNumber {
	return &IntNumber{x: new(big.Int).SetInt64(x)}
}

func convertUint64ToInt(x uint64) *IntNumber {
	return &IntNumber{x: new(big.Int).SetUint64(x)}
}

func convertFloat64ToInt(x float64) *IntNumber {
	return &IntNumber{x: bigFloatToBigInt(big.NewFloat(x))}
}

func convertStringToInt(x string) *IntNumber {
	if i, ok := new(big.Int).SetString(x, 0); ok {
		return &IntNumber{x: i}
	}
	return nil
}

func convertBytesToInt(x []byte) *IntNumber {
	return &IntNumber{x: new(big.Int).SetBytes(x)}
}

func anyToInt64(x any) int64 {
	switch x := x.(type) {
	case int:
		return int64(x)
	case int8:
		return int64(x)
	case int16:
		return int64(x)
	case int32:
		return int64(x)
	case int64:
		return x
	}
	return 0
}

func anyToUint64(x any) uint64 {
	switch x := x.(type) {
	case uint:
		return uint64(x)
	case uint8:
		return uint64(x)
	case uint16:
		return uint64(x)
	case uint32:
		return uint64(x)
	case uint64:
		return x
	}
	return 0
}

func anyToFloat64(x any) float64 {
	switch x := x.(type) {
	case float32:
		return float64(x)
	case float64:
		return x
	}
	return 0
}

func floatDecimalPrecision(x *FloatNumber) int {
	return stringNumberDecimalPrecision(x.Text('f', -1))
}

func bigFloatDecimalPrecision(x *big.Float) int {
	return stringNumberDecimalPrecision(x.Text('f', -1))
}

func float64DecimalPrecision(x float64) int {
	return stringNumberDecimalPrecision(big.NewFloat(x).Text('f', -1))
}

func stringNumberDecimalPrecision(x string) int {
	if f, ok := new(big.Float).SetString(x); ok {
		s := f.Text('f', -1)
		if len(s) == 0 {
			return 0
		}
		d := strings.Index(s, ".")
		if d == -1 {
			return 0
		}
		z := len(s) - 1
		for z >= d && s[z] == '0' {
			z--
		}
		return z - d
	}
	if _, ok := new(big.Int).SetString(x, 0); ok {
		return 0
	}
	return -1
}

func bigFloatToBigInt(x *big.Float) *big.Int {
	i, acc := x.Int(nil)
	if acc == big.Exact {
		return i
	}
	f := x.Sub(x, new(big.Float).SetInt(i))
	if f.Cmp(floatHalf) >= 0 {
		i.Add(i, big.NewInt(1))
	}
	return i
}

func pow10[T constraints.Integer](n T) *big.Int {
	return new(big.Int).Exp(intTen, big.NewInt(int64(n)), nil)
}
