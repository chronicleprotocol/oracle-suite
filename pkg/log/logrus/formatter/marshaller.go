package formatter

import (
	"encoding/json"
	"fmt"
	"reflect"

	"github.com/sirupsen/logrus"
)

// MarshallerFormatter formatter can marshal field values to string or JSON.
// Unlike default logrus implementation, it can handle nested types that
// support fmt.Stringer or json.Marshaler interfaces.
type MarshallerFormatter struct {
	Formatter logrus.Formatter
}

func (f *MarshallerFormatter) Format(e *logrus.Entry) ([]byte, error) {
	data := logrus.Fields{}
	for k, v := range e.Data {
		data[k] = format(v)
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
			m := map[interface{}]interface{}{}
			for _, k := range v.MapKeys() {
				m[k] = format(v.MapIndex(k).Interface())
			}
			return toJSON(m)
		case reflect.Ptr, reflect.Interface:
			return format(v.Elem().Interface())
		default:
			return fmt.Sprint(s)
		}
	}
}

func toJSON(s interface{}) string {
	j, err := json.Marshal(s)
	if err != nil {
		return err.Error()
	}
	return string(j)
}
