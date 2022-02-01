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
				Type:        "test",
				ID:          []byte{10, 10, 10},
				Index:       []byte{11, 11, 11},
				EventDate:   time.Unix(12, 0),
				MessageDate: time.Unix(13, 0),
				Data: map[string][]byte{
					"a": {14, 14, 14},
					"b": {15, 15, 15},
				},
				Signatures: map[string]EventSignature{
					"c": {Signer: []byte{16}, Signature: []byte{16}},
					"d": {Signer: []byte{17}, Signature: []byte{17}},
				},
			},
			wantErr: false,
		},
		{
			event: Event{
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
				assert.Equal(t, tt.event.Type, event.Type)
				assert.Equal(t, tt.event.ID, event.ID)
				assert.Equal(t, tt.event.Index, event.Index)
				assert.Equal(t, tt.event.EventDate, event.EventDate)
				assert.Equal(t, tt.event.MessageDate, event.MessageDate)
				assert.Equal(t, tt.event.Data, event.Data)
				assert.Equal(t, tt.event.Signatures, event.Signatures)
			}
		})
	}
}
