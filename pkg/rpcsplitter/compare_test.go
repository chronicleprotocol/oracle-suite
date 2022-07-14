//  Copyright (C) 2020 Maker Ecosystem Growth Holdings, INC.
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

package rpcsplitter

import (
	"fmt"
	"testing"

	"github.com/stretchr/testify/assert"
)

type testComparable struct{ v interface{} }
type testStruct struct{ V interface{} }

func (t *testComparable) Compare(v interface{}) bool {
	return true
}

func Test_compare(t *testing.T) {
	testSlice := []int{1}
	testMap := map[int]int{1: 1}
	testVar := 1
	testVarPtr := &testVar
	tests := []struct {
		a     interface{}
		b     interface{}
		equal bool
	}{
		// Int
		{a: 1, b: 1, equal: true},
		{a: 1, b: 2, equal: false},
		// String
		{a: "a", b: "a", equal: true},
		{a: "a", b: "b", equal: false},
		// Array
		{a: [1]int{0}, b: [1]int{0}, equal: true},
		{a: [1]int{0}, b: [1]int{1}, equal: false},
		{a: [1]int{0}, b: [2]int{0, 1}, equal: false},
		{a: [2]int{0, 0}, b: [1]int{0}, equal: false},
		// Slice
		{a: []int{0}, b: []int{0}, equal: true},
		{a: []int{0}, b: []int{1}, equal: false},
		{a: []int{0}, b: []int{0, 1}, equal: false},
		{a: []int{0, 1}, b: []int{0}, equal: false},
		// Map
		{a: map[int]int{1: 1}, b: map[int]int{1: 1}, equal: true},
		{a: map[int]int{1: 1, 2: 2}, b: map[int]int{1: 1}, equal: false},
		{a: map[int]int{1: 1}, b: map[int]int{1: 1, 2: 2}, equal: false},
		{a: map[int]int{1: 1, 3: 2}, b: map[int]int{1: 1, 2: 2}, equal: false},
		{a: map[int]int{1: 1, 2: 2}, b: map[int]int{1: 1, 3: 2}, equal: false},
		{a: map[int]int{1: 1}, b: map[float64]int{1: 1}, equal: false},
		// Nil
		{a: nil, b: nil, equal: true},
		{a: 1, b: nil, equal: false},
		{a: nil, b: 1, equal: false},
		// Pointers
		{a: testSlice, b: &testSlice, equal: true},
		{a: &testSlice, b: testSlice, equal: true},
		{a: &testSlice, b: &testSlice, equal: true},
		{a: &testSlice, b: ([]int)(nil), equal: false},
		{a: testMap, b: &testMap, equal: true},
		{a: &testMap, b: testMap, equal: true},
		{a: &testMap, b: &testMap, equal: true},
		{a: &testMap, b: &testSlice, equal: false},
		{a: &testMap, b: (map[int]int)(nil), equal: false},
		{a: testVar, b: testVarPtr, equal: true},
		{a: testVar, b: &testVarPtr, equal: true},
		{a: (*int)(nil), b: (*int)(nil), equal: true},
		// Struct
		{a: testStruct{V: 1}, b: testStruct{V: 1}, equal: true},
		{a: testStruct{V: 1}, b: testStruct{V: 2}, equal: false},
		// Compare method
		{a: testComparable{v: 1}, b: testComparable{v: 2}, equal: true},
	}
	for n, tt := range tests {
		t.Run(fmt.Sprintf("case-%d", n+1), func(t *testing.T) {
			assert.Equal(t, tt.equal, compare(tt.a, tt.b))
		})
	}
}
