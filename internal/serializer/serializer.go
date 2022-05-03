package serializer

import (
	"encoding/hex"
	"encoding/json"
	"fmt"
	"reflect"
	"strconv"
)

// Serialize converts an arbitrary value to a simple scalar value. Complex data
// structure data is represented as JSON. Binary data is converted to hex format.
//
// It does not support recursive data.
func Serialize(s interface{}) interface{} {
	switch ts := s.(type) {
	case float32, float64, int, int8, int16, int32, int64, uint, uint8, uint16, uint32, uint64, bool, string:
		return s
	case []byte:
		return "0x" + hex.EncodeToString(ts)
	case error:
		return ts.Error()
	case fmt.Stringer:
		return ts.String()
	case json.Marshaler:
		return toJSON(s)
	default:
		v := reflect.ValueOf(s)
		t := v.Type()
		switch v.Kind() {
		case reflect.Struct:
			m := map[string]interface{}{}
			for n := 0; n < v.NumField(); n++ {
				m[t.Field(n).Name] = Serialize(v.Field(n).Interface())
			}
			return toJSON(m)
		case reflect.Slice, reflect.Array:
			var m []interface{}
			for i := 0; i < v.Len(); i++ {
				m = append(m, Serialize(v.Index(i).Interface()))
			}
			return toJSON(m)
		case reflect.Map:
			m := map[string]interface{}{}
			for _, k := range v.MapKeys() {
				m[fmt.Sprint(Serialize(k))] = Serialize(v.MapIndex(k).Interface())
			}
			return toJSON(m)
		case reflect.Ptr, reflect.Interface:
			return Serialize(v.Elem().Interface())
		default:
			return fmt.Sprint(s)
		}
	}
}

func toJSON(s interface{}) json.RawMessage {
	j, err := json.Marshal(s)
	if err != nil {
		return json.RawMessage(strconv.Quote(err.Error()))
	}
	return j
}
