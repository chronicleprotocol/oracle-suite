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

package messages

import (
	"fmt"
	"strings"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestEvent_Marshalling(t *testing.T) {
	tests := []struct {
		event   Event
		wantErr bool
	}{
		{
			event: Event{
				Date:  time.Unix(9, 0),
				Type:  "test",
				ID:    []byte{10, 10, 10},
				Index: []byte{11, 11, 11},
				Data: map[string][]byte{
					"a": {12, 12, 12},
					"b": {13, 13, 13},
				},
				Signatures: map[string]EventSignature{
					"c": {Signer: []byte{14}, Signature: []byte{15}},
					"d": {Signer: []byte{15}, Signature: []byte{16}},
				},
			},
			wantErr: false,
		},
		{
			event: Event{
				Date: time.Unix(9, 0),
				Type: strings.Repeat("a", 1*1024*1024),
			},
			wantErr: true,
		},
	}
	for n, tt := range tests {
		t.Run(fmt.Sprintf("case-%d", n+1), func(t *testing.T) {
			msg, err := tt.event.MarshallBinary()
			if tt.wantErr {
				assert.Error(t, err)
			} else {
				event := &Event{}
				err := event.UnmarshallBinary(msg)

				require.NoError(t, err)
				assert.Equal(t, tt.event.Date, event.Date)
				assert.Equal(t, tt.event.Type, event.Type)
				assert.Equal(t, tt.event.ID, event.ID)
				assert.Equal(t, tt.event.Index, event.Index)
				assert.Equal(t, tt.event.Data, event.Data)
				assert.Equal(t, tt.event.Signatures, event.Signatures)
			}
		})
	}
}
