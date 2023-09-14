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
	"crypto/rand"
	"math"
	"math/big"

	"github.com/chronicleprotocol/oracle-suite/pkg/util/bn"
)

// calculateSpread calculates the spread between given price and a median
// price. The spread is returned as percentage points.
//
// The spread is calculated as:
//
//	abs((new - old) / old * 100)
//
// If old is zero, the result is positive infinity.
func calculateSpread(new, old *bn.DecFloatPointNumber) float64 {
	if old.Sign() == 0 {
		return math.Inf(1)
	}
	spread, _ := new.Sub(old).Div(old).Mul(bn.DecFloatPoint(100)).Abs().BigFloat().Float64()
	return spread
}

// randomInts generates a slice of integers from 0 to n (exclusive), shuffled
// using crypto secure source.
func randomInts(n int) ([]int, error) {
	// Generate slice of integers from 0 to n.
	ints := make([]int, n)
	for i := range ints {
		ints[i] = i
	}
	// Shuffle using a crypto secure source.
	for i := range ints {
		j, err := cryptoRandInt(i, n)
		if err != nil {
			return nil, err
		}
		ints[i], ints[j] = ints[j], ints[i]
	}
	return ints, nil
}

// cryptoRandInt returns a crypto secure random int in [min, max).
func cryptoRandInt(min, max int) (int, error) {
	n, err := rand.Int(rand.Reader, big.NewInt(int64(max-min)))
	if err != nil {
		return 0, err
	}
	return int(n.Int64()) + min, nil
}
