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

// comparable works with the compare method. If this interface is implemented,
// then the compare function will be used to compare values. The method
// will always be invoked on a pointer receiver, and the v argument will
// always be given as a pointer.
type comparable interface {
	// Compare returns true if the v arg is the same as a receiver.
	Compare(v interface{}) bool
}

// compare reports if two values are deeply equal. This function is similar to
// reflect.DeepEqual but it ignores pointers (comparing a value with the same
// value passed as a pointer will return true) and it will use a custom method
// for comparison if the value implements the comparable interface.
//
// If a structure contains unexported fields, compare will always return false.
// To properly handle those structures, the comparable interface needs to be
// implemented.
//
// This function DOES NOT work with recursive data structures!
//nolint:funlen,gocyclo
func compare(a, b interface{}) bool {
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
		if a.CanInterface() {
			if c, ok := ptr(a).Interface().(comparable); ok {
				return c.Compare(ptr(b).Interface())
			}
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
		default:
			if a.CanInterface() && a.Type().Comparable() {
				return a.Interface() == b.Interface()
			}
			return false
		}
	}
	return cmp(reflect.ValueOf(a), reflect.ValueOf(b))
}

// ptr returns a pointer to a given value.
func ptr(v reflect.Value) reflect.Value {
	pt := reflect.PtrTo(v.Type())
	pv := reflect.New(pt.Elem())
	pv.Elem().Set(v)
	return pv
}
