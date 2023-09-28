package dump

import (
	"encoding"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"reflect"
)

const dumpMaxDepth = 64

// Dump converts an arbitrary value to a scalar value.
//
// The purpose of this function is to provide an alternative to
// fmt verbs such as %#v or %+v that returns a value in a more
// human-readable format.
//
//   - Simple types, like numbers and strings and booleans, are returned as-is.
//   - For types that implement json.Marshaler, the result of MarshalJSON is
//     returned.
//   - For types that implement encoding.TextMarshaler, the result of
//     MarshalText is returned.
//   - For types that implement fmt.Stringer, the result of String is returned.
//   - Byte slices and arrays are represented as hex strings.
//   - For types that implement  error, the result of Error is returned.
//   - In maps, slices, and arrays, each element is recursively normalized
//     according to these rules and then represented as a JSON.
func Dump(v any) any {
	v = dump(v, dumpMaxDepth)
	if isSimpleType(v) {
		return v
	}
	b, err := json.Marshal(v)
	if err != nil {
		return fmt.Sprintf("<error: %v>", err)
	}
	return json.RawMessage(b)
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
	if isNil(v) {
		return nil
	}
	if isSimpleType(v) {
		return v
	}
	switch t := v.(type) {
	case json.RawMessage:
		return t
	case json.Marshaler:
		b, err := t.MarshalJSON()
		if err != nil {
			return fmt.Sprintf("<error: %v>", err)
		}
		return json.RawMessage(b)
	case encoding.TextMarshaler:
		b, err := t.MarshalText()
		if err != nil {
			return fmt.Sprintf("<error: %v>", err)
		}
		return string(b)
	case fmt.Stringer:
		return t.String()
	case error:
		return t.Error()
	case []byte:
		return "0x" + hex.EncodeToString(t)
	default:
		rv := reflect.ValueOf(v)
		if v == nil || rv.IsZero() {
			return nil
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
			return m
		case reflect.Slice, reflect.Array:
			var m []any
			for i := 0; i < rv.Len(); i++ {
				m = append(m, dump(rv.Index(i).Interface(), depth-1))
			}
			return m
		case reflect.Map:
			m := map[string]any{}
			for _, k := range rv.MapKeys() {
				m[fmt.Sprint(dump(k, depth-1))] = dump(rv.MapIndex(k).Interface(), depth-1)
			}
			return m
		case reflect.Ptr, reflect.Interface:
			return dump(rv.Elem().Interface(), depth-1)
		default:
			return fmt.Sprintf("%v", v)
		}
	}
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
