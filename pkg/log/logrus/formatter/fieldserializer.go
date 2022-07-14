package formatter

import (
	"encoding/json"

	"github.com/sirupsen/logrus"

	"github.com/chronicleprotocol/oracle-suite/pkg/util/dump"
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
		data[k] = dump.Dump(v)
		if v, ok := data[k].(json.RawMessage); ok && !f.UseJSONRawMessage {
			data[k] = string(v)
		}
	}
	e.Data = data
	return f.Formatter.Format(e)
}
