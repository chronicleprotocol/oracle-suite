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

package sliceutil

import (
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestCopy(t *testing.T) {
	m := []string{"a", "b", "c"}
	assert.Equal(t, m, Copy(m))
	assert.NotSame(t, m, Copy(m))
}

func TestContains(t *testing.T) {
	m := []string{"a", "b", "c"}
	assert.True(t, Contains(m, "a"))
	assert.False(t, Contains(m, "d"))
}

func TestContainsAll(t *testing.T) {
	m := []string{"a", "b", "c"}
	assert.True(t, ContainsAll(m, []string{"a", "b"}))
	assert.False(t, ContainsAll(m, []string{"a", "d"}))
}

func TestMap(t *testing.T) {
	m := []string{"a", "b", "c"}
	assert.Equal(t, []string{"A", "B", "C"}, Map(m, strings.ToUpper))
	assert.NotSame(t, m, Map(m, strings.ToUpper))
}

func TestFilter(t *testing.T) {
	m := []string{"a", "b", "c"}
	assert.Equal(t, []string{"a", "b"}, Filter(m, func(s string) bool { return s != "c" }))
	assert.NotSame(t, m, Filter(m, func(s string) bool { return s != "c" }))
}

func TestIsUnique(t *testing.T) {
	assert.True(t, IsUnique([]string{"a", "b", "c"}))
	assert.False(t, IsUnique([]string{"a", "b", "a"}))
}

func TestIntersect(t *testing.T) {
	assert.Equal(t, []string{"a", "b"}, Intersect([]string{"a", "b", "c"}, []string{"a", "b"}))
	assert.Equal(t, []string{"a", "b"}, Intersect([]string{"a", "b"}, []string{"a", "b", "c"}))
	assert.Equal(t, []string{"a", "b"}, Intersect([]string{"a", "b", "c"}, []string{"a", "b", "c"}, []string{"a", "b"}))
	assert.Equal(t, []string{}, Intersect([]string{"a", "b", "c"}, []string{"d", "e", "f"}))
	assert.Equal(t, []string{}, Intersect([]string{"d", "e", "f"}, []string{"a", "b", "c"}))
	assert.Equal(t, []string{}, Intersect([]string{"a", "b", "c"}, []string{}))
	assert.Equal(t, []string{}, Intersect([]string{}, []string{"a", "b", "c"}))
	assert.Equal(t, []string{}, Intersect([]string{}, []string{}))
}

func TestIndexOf(t *testing.T) {
	assert.Equal(t, 1, IndexOf([]string{"a", "b", "c"}, "b"))
	assert.Equal(t, -1, IndexOf([]string{"a", "b", "c"}, "d"))
}

func TestAppendUnique(t *testing.T) {
	assert.Equal(t, []string{"a", "b", "c"}, AppendUnique([]string{"a", "b"}, "c"))
	assert.Equal(t, []string{"a", "b", "c"}, AppendUnique([]string{"a", "b", "c"}, "c"))
	assert.Equal(t, []string{"a", "b", "c", "d"}, AppendUnique([]string{"a", "b", "c"}, "d"))
}
