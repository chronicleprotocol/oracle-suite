package dump

import (
	"bytes"
	"encoding"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"reflect"
	"strconv"
)

const dumpMaxDepth = 64

// Dump converts an arbitrary value to a scalar value.
//
// The purpose of this function is to provide an alternative to
// fmt verbs such as %#v or %+v that returns a value in a more
// human-readable format.
//
//   - Simple types, like numbers and strings and booleans, are returned as-is.
//   - For types that implement fmt.Stringer, the result of String is returned.
//   - For types that implement encoding.TextMarshaler, the result of
//     MarshalText is returned.
//   - For types that implement json.Marshaler, the result of MarshalJSON is
//     returned. If the result is a JSON string, number, or boolean, it is
//     converted to a Go type.
//   - Byte slices and arrays are represented as hex strings.
//   - For types that implement  error, the result of Error is returned.
//   - In maps, slices, and arrays, each element is recursively converted
//     according to these rules and then represented as a JSON.
//
// If a returned value is a JSON, it is returned as a json.RawMessage.
func Dump(v any) any {
	return dump(v, dumpMaxDepth)
}

//nolint:gocyclo,funlen
func dump(v any, depth int) (ret any) {
	defer func() {
		if r := recover(); r != nil {
			ret = fmt.Sprintf("<panic: %v>", r)
		}
	}()
	if depth <= 0 {
		return "<max depth reached>"
	}
	if isSimpleType(v) {
		return v
	}
	if isNil(v) {
		return nil
	}
	switch t := v.(type) {
	case fmt.Stringer:
		return t.String()
	case json.RawMessage:
		return fromJSON(t)
	case encoding.TextMarshaler:
		b, err := t.MarshalText()
		if err != nil {
			return fmt.Sprintf("<error: %v>", err)
		}
		return string(b)
	case json.Marshaler:
		return fromJSON(toJSON(t))
	case error:
		return t.Error()
	default:
		rv := reflect.ValueOf(v)
		if !rv.IsValid() {
			return "<invalid value>"
		}
		rt := rv.Type()
		switch rv.Kind() {
		case reflect.Struct:
			m := map[string]any{}
			for n := 0; n < rv.NumField(); n++ {
				if rt.Field(n).IsExported() {
					m[rt.Field(n).Name] = dump(rv.Field(n).Interface(), depth-1)
				}
			}
			return toJSON(m)
		case reflect.Slice, reflect.Array:
			if rt.Elem().Kind() == reflect.Uint8 {
				if !rv.CanAddr() {
					cpy := reflect.MakeSlice(reflect.SliceOf(reflect.TypeOf(byte(0))), rv.Len(), rv.Len())
					reflect.Copy(cpy, rv)
					rv = cpy
				}
				return "0x" + hex.EncodeToString(rv.Bytes())
			}
			var m []any
			for i := 0; i < rv.Len(); i++ {
				m = append(m, dump(rv.Index(i).Interface(), depth-1))
			}
			return toJSON(m)
		case reflect.Map:
			m := map[string]any{}
			for _, k := range rv.MapKeys() {
				m[fmt.Sprint(dump(k, depth-1))] = dump(rv.MapIndex(k).Interface(), depth-1)
			}
			return toJSON(m)
		case reflect.Ptr, reflect.Interface:
			return dump(rv.Elem().Interface(), depth-1)
		default:
			return fmt.Sprintf("%v", v)
		}
	}
}

func toJSON(v any) any {
	b, err := json.Marshal(v)
	if err != nil {
		return fmt.Sprintf("<error: %v>", err)
	}
	return json.RawMessage(b)
}

func fromJSON(v any) any {
	if v == nil {
		return nil
	}
	if b, ok := v.(json.RawMessage); ok {
		if bytes.HasPrefix(b, []byte{'"'}) {
			if s, err := strconv.Unquote(string(b)); err == nil {
				return s
			}
			return v
		}
		if i, err := strconv.ParseInt(string(b), 10, 64); err == nil {
			return i
		}
		if f, err := strconv.ParseFloat(string(b), 64); err == nil {
			return f
		}
		if b, err := strconv.ParseBool(string(b)); err == nil {
			return b
		}
	}
	return v
}

func isSimpleType(v any) bool {
	switch v.(type) {
	case nil:
		return true
	case int, int8, int16, int32, int64, uint, uint8, uint16, uint32, uint64:
		return true
	case float32, float64:
		return true
	case bool:
		return true
	case string:
		return true
	default:
		return false
	}
}

func isNil(v any) bool {
	if v == nil {
		return true
	}
	switch reflect.TypeOf(v).Kind() {
	case reflect.Chan, reflect.Func, reflect.Interface, reflect.Map, reflect.Ptr, reflect.Slice:
		return reflect.ValueOf(v).IsNil()
	}
	return false
}
