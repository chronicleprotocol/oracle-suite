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
	"reflect"
)

// compare reports if two values are deeply equal. This function is similar to
// reflect.DeepEqual but it ignores pointers (comparing a value with the same
// value passed as a pointer will return true).
//
// If a structure contains unexported fields, compare will always return false.
//
// This function DOES NOT work with recursive data structures!
//nolint:funlen,gocyclo
func compare(a, b any) bool {
	if a == nil && b == nil {
		return true
	}
	var cmp func(a, b reflect.Value) bool
	cmp = func(a, b reflect.Value) bool {
		if !a.IsValid() || !b.IsValid() {
			return a.IsValid() == b.IsValid()
		}
		if a.Kind() == reflect.Ptr && b.Kind() == reflect.Ptr && a.Pointer() == b.Pointer() {
			return true
		}
		if a.Kind() == reflect.Ptr || a.Kind() == reflect.Interface {
			return cmp(a.Elem(), b)
		}
		if b.Kind() == reflect.Ptr || b.Kind() == reflect.Interface {
			return cmp(a, b.Elem())
		}
		if a.Type() != b.Type() {
			return false
		}
		switch a.Kind() {
		case reflect.Array:
			for i := 0; i < a.Len(); i++ {
				if !cmp(a.Index(i), b.Index(i)) {
					return false
				}
			}
			return true
		case reflect.Slice:
			if a.Pointer() == b.Pointer() {
				return true
			}
			if a.Len() != b.Len() {
				return false
			}
			for i := 0; i < a.Len(); i++ {
				if !cmp(a.Index(i), b.Index(i)) {
					return false
				}
			}
			return true
		case reflect.Map:
			if a.Pointer() == b.Pointer() {
				return true
			}
			if a.Len() != b.Len() {
				return false
			}
			for _, k := range a.MapKeys() {
				av := a.MapIndex(k)
				bv := b.MapIndex(k)
				if !cmp(av, bv) {
					return false
				}
			}
			return true
		case reflect.Struct:
			for i := 0; i < a.NumField(); i++ {
				if !cmp(a.Field(i), b.Field(i)) {
					return false
				}
			}
			return true
		case reflect.String:
			return a.String() == b.String()
		case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
			return a.Int() == b.Int()
		case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64, reflect.Uintptr:
			return a.Uint() == b.Uint()
		case reflect.Float32, reflect.Float64:
			return a.Float() == b.Float()
		case reflect.Complex64, reflect.Complex128:
			return a.Complex() == b.Complex()
		case reflect.Bool:
			return a.Bool() == b.Bool()
		default:
			if a.CanInterface() && a.Type().Comparable() {
				return a.Interface() == b.Interface()
			}
			return false
		}
	}
	return cmp(reflect.ValueOf(a), reflect.ValueOf(b))
}
