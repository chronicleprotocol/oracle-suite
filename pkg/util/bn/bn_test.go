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
	"testing"

	"github.com/stretchr/testify/assert"
)

func Test_countFractionalDigits(t *testing.T) {
	tests := []struct {
		num  string
		want int
	}{
		{"0", 0},
		{"1", 0},
		{"0.0", 0},
		{"0.1", 1},
		{"0.01", 2},
		{"0.001", 3},
		{"1.001", 3},
		{"1.0010", 3},
	}
	for _, tt := range tests {
		t.Run(tt.num, func(t *testing.T) {
			p, ok := stringNumberDecimalPrecision(tt.num)
			assert.Equal(t, uint32(tt.want), p)
			assert.True(t, ok)
		})
	}
}
