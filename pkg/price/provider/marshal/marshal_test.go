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
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestNewMarshaller(t *testing.T) {
	expectedMap := map[FormatType]interface{}{
		Plain:  (*plain)(nil),
		JSON:   (*json)(nil),
		NDJSON: (*json)(nil),
		Trace:  (*trace)(nil),
	}
	formatMap := map[FormatType]string{
		Plain:  "plain",
		JSON:   "json",
		NDJSON: "ndjson",
		Trace:  "trace",
	}
	for ct, st := range formatMap {
		t.Run(st, func(t *testing.T) {
			m, err := NewMarshal(ct)

			assert.NoError(t, err)
			assert.Implements(t, (*Marshaller)(nil), m)
			assert.IsType(t, m, &Marshal{})
			assert.IsType(t, m.marshaller, expectedMap[ct])
		})
	}
}
