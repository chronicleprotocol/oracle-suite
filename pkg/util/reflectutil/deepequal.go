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

package reflectutil

import (
	"reflect"
)

// FilterFunc filter out to check if two values don't need to compare
// Return false not to compare, true to compare
type FilterFunc func(any, any) bool

// DeepEqual reports whether x and y are "deeply equal".
// Do not test if filter function returns true for comparing two values.
func DeepEqual(x, y any, filter FilterFunc) bool {
	if x == nil || y == nil {
		return x == y
	}
	v1 := reflect.ValueOf(x)
	v2 := reflect.ValueOf(y)
	if v1.Type() != v2.Type() {
		return false
	}
	return deepEqual(v1, v2, filter)
}

// deepEqual tests for deep equality using reflected types recursively.
// Do not test if filter function returns true for comparing two values.
func deepEqual(v1, v2 reflect.Value, filter FilterFunc) bool { //nolint:gocyclo,funlen
	if !v1.IsValid() || !v2.IsValid() {
		return v1.IsValid() == v2.IsValid()
	}
	if v1.Type() != v2.Type() {
		return false
	}
	if filter != nil && !filter(v1, v2) {
		return true
	}

	switch v1.Kind() {
	case reflect.Struct:
		if v1.NumField() != v2.NumField() {
			return false
		}
		for i := 0; i < v1.NumField(); i++ {
			fieldStruct1 := v1.Type().Field(i)
			fieldStruct2 := v1.Type().Field(i)
			if filter != nil && !filter(fieldStruct1, fieldStruct2) {
				continue
			}
			if !deepEqual(v1.Field(i), v2.Field(i), filter) {
				return false
			}
		}
	case reflect.Pointer:
		if v1.UnsafePointer() == v2.UnsafePointer() {
			return true
		}
		return deepEqual(v1.Elem(), v2.Elem(), filter)
	case reflect.Slice:
		if v1.IsNil() != v2.IsNil() {
			return false
		}
		if v1.Len() != v2.Len() {
			return false
		}
		if v1.UnsafePointer() == v2.UnsafePointer() {
			return true
		}
		for i := 0; i < v1.Len(); i++ {
			if !deepEqual(v1.Index(i), v2.Index(i), filter) {
				return false
			}
		}
	case reflect.Map:
		if v1.IsNil() != v2.IsNil() {
			return false
		}
		if v1.Len() != v2.Len() {
			return false
		}
		if v1.UnsafePointer() == v2.UnsafePointer() {
			return true
		}
		if len(v1.MapKeys()) != len(v2.MapKeys()) {
			return false
		}
		for _, k := range v1.MapKeys() {
			val1 := v1.MapIndex(k)
			val2 := v2.MapIndex(k)
			if !deepEqual(val1, val2, filter) {
				return false
			}
		}
	case reflect.Func:
		if v1.IsNil() && v2.IsNil() {
			return true
		}
		return false
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
		return v1.Int() == v2.Int()
	case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64, reflect.Uintptr:
		return v1.Uint() == v2.Uint()
	case reflect.String:
		return v1.String() == v2.String()
	case reflect.Bool:
		return v1.Bool() == v2.Bool()
	case reflect.Float32, reflect.Float64:
		return v1.Float() == v2.Float()
	case reflect.Complex64, reflect.Complex128:
		return v1.Complex() == v2.Complex()
	default:
		return v1.Interface() == v2.Interface()
	}
	return true
}
