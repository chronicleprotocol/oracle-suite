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

package hcl

import (
	"encoding"
	"fmt"
	"math/big"
	"reflect"

	"github.com/defiweb/go-anymapper"
	"github.com/hashicorp/hcl/v2"
	"github.com/zclconf/go-cty/cty"
)

// Marshaler marshals cty.Value.
type Marshaler interface {
	MarshalHCL() (cty.Value, error)
}

// Unmarshaler unmarshals a value from cty.Value.
type Unmarshaler interface {
	UnmarshalHCL(cty.Value) error
}

var mapper *anymapper.Mapper

var (
	bodyTy        = reflect.TypeOf((*hcl.Body)(nil)).Elem()
	bodyContentTy = reflect.TypeOf((*hcl.BodyContent)(nil)).Elem()
	bodySchemaTy  = reflect.TypeOf((*hcl.BodySchema)(nil)).Elem()
	rangeTy       = reflect.TypeOf((*hcl.Range)(nil)).Elem()
	ctyValTy      = reflect.TypeOf((*cty.Value)(nil)).Elem()
	bigIntTy      = reflect.TypeOf((*big.Int)(nil)).Elem()
	bigFloatTy    = reflect.TypeOf((*big.Float)(nil)).Elem()
	stringTy      = reflect.TypeOf("")
	anyTy         = reflect.TypeOf((*any)(nil)).Elem()
)

// derefType dereferences the given type until it is not a pointer or an
// interface.
func derefType(t reflect.Type) reflect.Type {
	for t.Kind() == reflect.Ptr || t.Kind() == reflect.Interface {
		t = t.Elem()
	}
	return t
}

// derefValue dereferences the given value until it is not a pointer or an
// interface. If the value is a nil pointer, it is initialized.
func derefValue(v reflect.Value) reflect.Value {
	for v.Kind() == reflect.Ptr || v.Kind() == reflect.Interface {
		if v.Kind() == reflect.Ptr && v.IsNil() && v.CanSet() {
			v.Set(reflect.New(v.Type().Elem()))
		}
		v = v.Elem()
	}
	return v
}

func ctyMapper(m *anymapper.Mapper, src, dst reflect.Type) anymapper.MapFunc {
	if src == ctyValTy {
		return fromCtyMapper(m, src, dst)
	}
	if dst == ctyValTy {
		return toCtyMapper(m, src, dst)
	}
	return nil
}

// ctyMapper is a mapping function that maps cty.Value to other types.
//
//nolint:funlen,gocyclo
func fromCtyMapper(_ *anymapper.Mapper, src, dst reflect.Type) anymapper.MapFunc {
	if src != ctyValTy {
		return nil
	}

	// cty.Value -> any
	// To be able to reuse the existing mapping functions defined below, we
	// create an auxiliary variable based on the cty.Value type, and we use
	// that variable as the destination.
	if dst == anyTy {
		return func(m *anymapper.Mapper, _ *anymapper.Context, src, dst reflect.Value) error {
			typ := src.Interface().(cty.Value).Type()
			switch {
			case typ == cty.String:
				var aux string
				if err := m.MapRefl(src, reflect.ValueOf(&aux)); err != nil {
					return err
				}
				dst.Set(reflect.ValueOf(aux))
			case typ == cty.Number:
				var aux float64
				if err := m.MapRefl(src, reflect.ValueOf(&aux)); err != nil {
					return err
				}
				dst.Set(reflect.ValueOf(aux))
			case typ == cty.Bool:
				var aux bool
				if err := m.MapRefl(src, reflect.ValueOf(&aux)); err != nil {
					return err
				}
				dst.Set(reflect.ValueOf(aux))
			case typ.IsListType() || typ.IsSetType() || typ.IsTupleType():
				var aux []any
				if err := m.MapRefl(src, reflect.ValueOf(&aux)); err != nil {
					return err
				}
				dst.Set(reflect.ValueOf(aux))
			case typ.IsMapType() || typ.IsObjectType():
				var aux map[string]any
				if err := m.MapRefl(src, reflect.ValueOf(&aux)); err != nil {
					return err
				}
				dst.Set(reflect.ValueOf(aux))
			case typ == cty.DynamicPseudoType:
				dst.Set(reflect.Zero(dst.Type()))
			default:
				dst.Set(src)
			}
			return nil
		}
	}

	// cty.Value -> cty.Value
	if dst == ctyValTy {
		return func(m *anymapper.Mapper, _ *anymapper.Context, src, dst reflect.Value) error {
			dst.Set(src)
			return nil
		}
	}

	// cty.Value -> big.Int
	if dst == bigIntTy {
		return func(m *anymapper.Mapper, _ *anymapper.Context, src, dst reflect.Value) error {
			val := src.Interface().(cty.Value)
			if val.Type() != cty.Number {
				return fmt.Errorf("cannot decode %s into big.Int", val.Type().FriendlyName())
			}
			if !val.AsBigFloat().IsInt() {
				return fmt.Errorf("cannot decode a float number into big.Int")
			}
			bi, acc := val.AsBigFloat().Int(nil)
			if acc != big.Exact {
				return fmt.Errorf("cannot decode a float number into big.Int")
			}
			dst.Set(reflect.ValueOf(bi).Elem())
			return nil
		}
	}

	// cty.Value -> big.Float
	if dst == bigFloatTy {
		return func(m *anymapper.Mapper, _ *anymapper.Context, src, dst reflect.Value) error {
			val := src.Interface().(cty.Value)
			if val.Type() != cty.Number {
				return fmt.Errorf("cannot decode %s into big.Float", val.Type().FriendlyName())
			}
			dst.Set(reflect.ValueOf(val.AsBigFloat()).Elem())
			return nil
		}
	}

	// cty.Value -> Unmarshaler
	// cty.Value -> TextUnmarshaler
	// cty.Value -> string
	// cty.Value -> bool
	// cty.Value -> int*
	// cty.Value -> uint*
	// cty.Value -> float*
	// cty.Value -> slice
	// cty.Value -> map
	return func(m *anymapper.Mapper, _ *anymapper.Context, src, dst reflect.Value) error {
		ctyVal := src.Interface().(cty.Value)

		// Try to use unmarshaler interfaces.
		if dst.CanAddr() {
			if u, ok := dst.Addr().Interface().(Unmarshaler); ok {
				return u.UnmarshalHCL(ctyVal)
			}
			if u, ok := dst.Addr().Interface().(encoding.TextUnmarshaler); ok && ctyVal.Type() == cty.String {
				return u.UnmarshalText([]byte(ctyVal.AsString()))
			}
		}

		// Try to map the cty.Value to the basic types.
		switch dst.Kind() {
		case reflect.String:
			if ctyVal.Type() != cty.String {
				return fmt.Errorf(
					"cannot decode %s type into a string",
					ctyVal.Type().FriendlyName(),
				)
			}
			dst.SetString(ctyVal.AsString())
		case reflect.Bool:
			if ctyVal.Type() != cty.Bool {
				return fmt.Errorf(
					"cannot decode %s type into a bool",
					ctyVal.Type().FriendlyName(),
				)
			}
			dst.SetBool(ctyVal.True())
		case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
			if ctyVal.Type() != cty.Number {
				return fmt.Errorf(
					"cannot decode %s type into a %s type",
					ctyVal.Type().FriendlyName(), dst.Kind(),
				)
			}
			if !ctyVal.AsBigFloat().IsInt() {
				return fmt.Errorf(
					"cannot decode %s type into a %s type: not an integer",
					ctyVal.Type().FriendlyName(),
					dst.Kind(),
				)
			}
			i64, acc := ctyVal.AsBigFloat().Int64()
			if acc != big.Exact {
				return fmt.Errorf(
					"cannot decode %s type into a %s type: too large",
					ctyVal.Type().FriendlyName(),
					dst.Kind(),
				)
			}
			return m.MapReflContext(m.Context.WithStrictTypes(false), reflect.ValueOf(i64), dst)
		case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64:
			if ctyVal.Type() != cty.Number {
				return fmt.Errorf(
					"cannot decode %s type into a %s type",
					ctyVal.Type().FriendlyName(),
					dst.Kind(),
				)
			}
			if !ctyVal.AsBigFloat().IsInt() {
				return fmt.Errorf(
					"cannot decode %s type into a %s type: not an integer",
					ctyVal.Type().FriendlyName(),
					dst.Kind(),
				)
			}
			u64, acc := ctyVal.AsBigFloat().Uint64()
			if acc != big.Exact {
				return fmt.Errorf(
					"cannot decode %s type into a %s type: too large",
					ctyVal.Type().FriendlyName(),
					dst.Kind(),
				)
			}
			return m.MapReflContext(m.Context.WithStrictTypes(false), reflect.ValueOf(u64), dst)
		case reflect.Float32, reflect.Float64:
			if ctyVal.Type() != cty.Number {
				return fmt.Errorf(
					"cannot decode %s type into a %s type",
					ctyVal.Type().FriendlyName(),
					dst.Kind(),
				)
			}
			return m.MapReflContext(m.Context.WithStrictTypes(false), reflect.ValueOf(ctyVal.AsBigFloat()), dst)
		case reflect.Slice:
			if !ctyVal.Type().IsListType() && !ctyVal.Type().IsSetType() && !ctyVal.Type().IsTupleType() {
				return fmt.Errorf(
					"cannot decode %s type into a slice",
					ctyVal.Type().FriendlyName(),
				)
			}
			dstSlice := reflect.MakeSlice(dst.Type(), 0, ctyVal.LengthInt())
			for it := ctyVal.ElementIterator(); it.Next(); {
				_, v := it.Element()
				elem := reflect.New(dst.Type().Elem())
				if err := m.MapRefl(reflect.ValueOf(v), elem); err != nil {
					return err
				}
				dstSlice = reflect.Append(dstSlice, elem.Elem())
			}
			dst.Set(dstSlice)
		case reflect.Map:
			if !ctyVal.Type().IsMapType() && !ctyVal.Type().IsObjectType() {
				return fmt.Errorf(
					"cannot decode %s type into a map",
					ctyVal.Type().FriendlyName(),
				)
			}
			dstMap := reflect.MakeMap(dst.Type())
			for it := ctyVal.ElementIterator(); it.Next(); {
				k, v := it.Element()
				key := reflect.New(dst.Type().Key())
				if err := m.MapRefl(reflect.ValueOf(k), key); err != nil {
					return err
				}
				val := reflect.New(dst.Type().Elem())
				if err := m.MapRefl(reflect.ValueOf(v), val); err != nil {
					return err
				}
				dstMap.SetMapIndex(key.Elem(), val.Elem())
			}
			dst.Set(dstMap)
		default:
			return fmt.Errorf("unsupported type %s", dst.Type())
		}
		return nil
	}
}

// toCtyMapper is a mapping function that maps any types to cty.Value.
func toCtyMapper(_ *anymapper.Mapper, src, dst reflect.Type) anymapper.MapFunc { //nolint:gocyclo
	if dst != ctyValTy {
		return nil
	}

	// cty.Value -> cty.Value
	if src == ctyValTy {
		return func(m *anymapper.Mapper, _ *anymapper.Context, src, dst reflect.Value) error {
			dst.Set(src)
			return nil
		}
	}

	// big.Int -> cty.Value
	if src == bigIntTy {
		return func(_ *anymapper.Mapper, _ *anymapper.Context, src, dst reflect.Value) error {
			val, ok := src.Interface().(big.Int)
			if !ok {
				return fmt.Errorf("cannot encode from big.Int %s", src.Type().Name())
			}
			dst.Set(reflect.ValueOf(cty.NumberUIntVal(val.Uint64())))
			return nil
		}
	}

	// big.Float -> cty.Value
	if src == bigFloatTy {
		return func(_ *anymapper.Mapper, _ *anymapper.Context, src, dst reflect.Value) error {
			val, ok := src.Interface().(big.Float)
			if !ok {
				return fmt.Errorf("cannot encode from big.Float %s", src.Type().Name())
			}
			valFloat, _ := val.Float64()
			dst.Set(reflect.ValueOf(cty.NumberFloatVal(valFloat)))
			return nil
		}
	}

	// string -> cty.Value
	// bool -> cty.Value
	// int -> cty.Value
	// uint -> cty.Value
	// float -> cty.Value
	// slice -> cty.Value
	// map -> cty.Value
	return func(m *anymapper.Mapper, _ *anymapper.Context, src, dst reflect.Value) error {
		// Try to use marshaler interfaces.
		if src.CanAddr() { // i.e. *config.URL
			switch t := src.Addr().Interface().(type) {
			case Marshaler:
				ctyVal, err := t.MarshalHCL()
				if err != nil {
					return err
				}
				dst.Set(reflect.ValueOf(ctyVal))
				return nil
			case encoding.TextMarshaler:
				text, err := t.MarshalText()
				if err != nil {
					return err
				}
				ctyVal := cty.StringVal(string(text))
				dst.Set(reflect.ValueOf(ctyVal))
				return nil
			}
		} else { // i.e. origin.ContractAddresses
			switch t := src.Interface().(type) {
			case Marshaler:
				ctyVal, err := t.MarshalHCL()
				if err != nil {
					return err
				}
				dst.Set(reflect.ValueOf(ctyVal))
				return nil
			case encoding.TextMarshaler:
				text, err := t.MarshalText()
				if err != nil {
					return err
				}
				ctyVal := cty.StringVal(string(text))
				dst.Set(reflect.ValueOf(ctyVal))
				return nil
			}
		}

		switch src.Kind() {
		case reflect.String:
			ctyVal := cty.StringVal(src.String())
			dst.Set(reflect.ValueOf(ctyVal))
		case reflect.Bool:
			ctyVal := cty.BoolVal(src.Bool())
			dst.Set(reflect.ValueOf(ctyVal))
		case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
			ctyVal := cty.NumberIntVal(src.Int())
			dst.Set(reflect.ValueOf(ctyVal))
		case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64:
			ctyVal := cty.NumberUIntVal(src.Uint())
			dst.Set(reflect.ValueOf(ctyVal))
		case reflect.Float32, reflect.Float64:
			ctyVal := cty.NumberFloatVal(src.Float())
			dst.Set(reflect.ValueOf(ctyVal))
		case reflect.Slice:
			dstSlice := make([]cty.Value, src.Len())
			for i := 0; i < src.Len(); i++ {
				elem := reflect.New(dst.Type())
				if err := m.MapRefl(src.Index(i), elem); err != nil {
					return err
				}
				dstSlice[i] = *(elem.Interface().(*cty.Value))
			}
			if src.Len() > 0 {
				dst.Set(reflect.ValueOf(cty.ListVal(dstSlice)))
			} else {
				dst.Set(reflect.ValueOf(cty.ListValEmpty(cty.NilType)))
			}
		case reflect.Map:
			dstMap := make(map[string]cty.Value)
			for it := src.MapRange(); it.Next(); {
				keyRv := reflect.New(stringTy)
				if err := m.MapRefl(it.Key(), keyRv); err != nil {
					return err
				}
				key := derefValue(keyRv).Interface().(string)
				if key == "" {
					continue
				}
				valRv := reflect.New(dst.Type())
				if err := m.MapRefl(it.Value(), valRv); err != nil {
					return err
				}
				val := derefValue(valRv).Interface().(cty.Value)
				dstMap[key] = val
			}
			dst.Set(reflect.ValueOf(cty.MapVal(dstMap)))
		default:
			return fmt.Errorf("unsupported type %s to ctyValue", src.Type())
		}
		return nil
	}
}

func init() {
	mapper = anymapper.New()
	mapper.Context.StrictTypes = true
	mapper.Mappers[ctyValTy] = ctyMapper
}
