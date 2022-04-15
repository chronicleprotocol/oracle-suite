package formatter

import (
	"encoding/hex"
	"encoding/json"
	"fmt"
	"reflect"
	"strconv"

	"github.com/sirupsen/logrus"
)

// FieldSerializerFormatter will serialize the log field values to basic types.
// Other types will be serialized to JSON.
type FieldSerializerFormatter struct {
	Formatter         logrus.Formatter
	UseJSONRawMessage bool // If true, then json.RawMessage type will be used for fields serialized to JSON.
}

func (f *FieldSerializerFormatter) Format(e *logrus.Entry) ([]byte, error) {
	data := logrus.Fields{}
	for k, v := range e.Data {
		data[k] = format(v)
		if v, ok := data[k].(json.RawMessage); ok && !f.UseJSONRawMessage {
			data[k] = string(v)
		}
	}
	e.Data = data
	return f.Formatter.Format(e)
}

func format(s interface{}) interface{} {
	switch ts := s.(type) {
	case fmt.Stringer:
		return ts.String()
	case json.Marshaler:
		return toJSON(s)
	case []byte:
		return hex.EncodeToString(ts)
	case error:
		return ts.Error()
	default:
		v := reflect.ValueOf(s)
		t := v.Type()
		switch v.Kind() {
		case reflect.Struct:
			m := map[string]interface{}{}
			for n := 0; n < v.NumField(); n++ {
				m[t.Field(n).Name] = format(v.Field(n).Interface())
			}
			return toJSON(m)
		case reflect.Slice, reflect.Array:
			var m []interface{}
			for i := 0; i < v.Len(); i++ {
				m = append(m, format(v.Index(i).Interface()))
			}
			return toJSON(m)
		case reflect.Map:
			m := map[string]interface{}{}
			for _, k := range v.MapKeys() {
				m[fmt.Sprint(format(k))] = format(v.MapIndex(k).Interface())
			}
			return toJSON(m)
		case reflect.Ptr, reflect.Interface:
			return format(v.Elem().Interface())
		case reflect.Bool, reflect.Float32, reflect.Float64,
			reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64,
			reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
			return s
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
