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

package marshal

import (
	"bytes"
	"fmt"
	"io"
)

// FormatType describes output format type.
type FormatType int

const (
	Plain FormatType = iota
	JSON
	NDJSON
	Trace
)

// Marshaller is the interface which must be implemented by different
// marshallers used to format output for the CLI.
type Marshaller interface {
	Write(writer io.Writer, item interface{}) error
	Flush() error
}

// Marshal implements the Marshaller interface. It wraps other marshaller based
// on argument passed to the NewMarshal method.
type Marshal struct {
	marshaller Marshaller
}

// NewMarshal returns new Marshal instance.
func NewMarshal(format FormatType) (*Marshal, error) {
	switch format {
	case Plain:
		return &Marshal{marshaller: newPlain()}, nil
	case JSON:
		return &Marshal{marshaller: newJSON(false)}, nil
	case NDJSON:
		return &Marshal{marshaller: newJSON(true)}, nil
	case Trace:
		return &Marshal{marshaller: newTrace()}, nil
	}

	return nil, fmt.Errorf("unsupported format")
}

// Write implements the Marshaller interface.
func (m *Marshal) Write(writer io.Writer, item interface{}) error {
	return m.marshaller.Write(writer, item)
}

// Flush implements the Marshaller interface.
func (m *Marshal) Flush() error {
	return m.marshaller.Flush()
}

// Marshall marshals list of items.
func Marshall(format FormatType, items ...interface{}) ([]byte, error) {
	m, err := NewMarshal(format)
	if err != nil {
		return nil, err
	}

	buf := &bytes.Buffer{}
	for _, item := range items {
		err = m.Write(buf, item)
		if err != nil {
			return nil, err
		}
	}

	err = m.Flush()
	if err != nil {
		return nil, err
	}

	return buf.Bytes(), nil
}
