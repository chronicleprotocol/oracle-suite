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

package maputil

import (
	"sort"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestKeys(t *testing.T) {
	t.Run("case-1", func(t *testing.T) {
		assert.ElementsMatch(t, []string{"a", "b"}, Keys(map[string]string{"a": "a", "b": "b"}))
	})
	t.Run("case-2", func(t *testing.T) {
		assert.ElementsMatch(t, []int{1, 2}, Keys(map[int]int{1: 1, 2: 2}))
	})
}

func TestSortKeys(t *testing.T) {
	t.Run("case-1", func(t *testing.T) {
		m := map[string]string{"b": "b", "a": "a"}
		assert.Equal(t, []string{"a", "b"}, SortKeys(m, sort.Strings))
	})
	t.Run("case-2", func(t *testing.T) {
		m := map[int]int{2: 2, 1: 1}
		assert.Equal(t, []int{1, 2}, SortKeys(m, sort.Ints))
	})
}

func TestCopy(t *testing.T) {
	t.Run("case-1", func(t *testing.T) {
		m := map[string]string{"a": "a", "b": "b"}
		assert.Equal(t, m, Copy(m))
		assert.NotSame(t, m, Copy(m))
	})
	t.Run("case-2", func(t *testing.T) {
		m := map[int]int{1: 1, 2: 2}
		assert.Equal(t, m, Copy(m))
		assert.NotSame(t, m, Copy(m))
	})
}

func TestMerge(t *testing.T) {
	t.Run("case-1", func(t *testing.T) {
		m1 := map[string]string{"a": "a"}
		m2 := map[string]string{"b": "b"}
		assert.Equal(t, map[string]string{"a": "a", "b": "b"}, Merge(m1, m2))
	})
	t.Run("case-2", func(t *testing.T) {
		m1 := map[int]int{1: 1}
		m2 := map[int]int{2: 2}
		assert.Equal(t, map[int]int{1: 1, 2: 2}, Merge(m1, m2))
	})
}
