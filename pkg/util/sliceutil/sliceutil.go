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

package sliceutil

// Copy returns a copy of the slice.
func Copy[T any](s []T) []T {
	newSlice := make([]T, len(s))
	copy(newSlice, s)
	return newSlice
}

// Contains returns true if s slice contains e element.
func Contains[T comparable](s []T, e T) bool {
	for _, x := range s {
		if x == e {
			return true
		}
	}
	return false
}
