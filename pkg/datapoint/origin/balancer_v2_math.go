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
