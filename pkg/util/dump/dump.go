package dump

import (
	"encoding/hex"
	"encoding/json"
	"fmt"
	"reflect"
	"strconv"
)

// Dump converts an arbitrary value to a simple scalar value.
//
// The purpose of this function is to provide an alternative to
// fmt verbs such as %#v or %+v that returns a value in a more
// human-readable format.
//
// String, int and floats are returned as is.
// Byte slices are converted to hex format.
// Complex data structures are represented as JSON.
//
// It does not support recursive data.
//
//nolint:gocyclo
func Dump(v any) any {
	if isNil(v) {
		return nil
	}
	switch tv := v.(type) {
	case nil:
		return nil
	case float32, float64, int, int8, int16, int32, int64, uint, uint8, uint16, uint32, uint64, bool, string:
		return v
	case []byte:
		return "0x" + hex.EncodeToString(tv)
	case error:
		return tv.Error()
	case fmt.Stringer:
		return tv.String()
	case json.Marshaler:
		return toJSON(v)
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
				m[rt.Field(n).Name] = Dump(rv.Field(n).Interface())
			}
			return toJSON(m)
		case reflect.Slice, reflect.Array:
			var m []any
			for i := 0; i < rv.Len(); i++ {
				m = append(m, Dump(rv.Index(i).Interface()))
			}
			return toJSON(m)
		case reflect.Map:
			m := map[string]any{}
			for _, k := range rv.MapKeys() {
				m[fmt.Sprint(Dump(k))] = Dump(rv.MapIndex(k).Interface())
			}
			return toJSON(m)
		case reflect.Ptr, reflect.Interface:
			return Dump(rv.Elem().Interface())
		default:
			return fmt.Sprint(rv)
		}
	}
}

func toJSON(v any) json.RawMessage {
	b, err := json.Marshal(v)
	if err != nil {
		return json.RawMessage(strconv.Quote(err.Error()))
	}
	return b
}

func isNil(i any) bool {
	if i == nil {
		return true
	}
	switch reflect.TypeOf(i).Kind() {
	case reflect.Chan, reflect.Func, reflect.Interface, reflect.Map, reflect.Ptr, reflect.Slice:
		return reflect.ValueOf(i).IsNil()
	}
	return false
}
