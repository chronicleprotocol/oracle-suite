//  Copyright (C) 2021-2023 Chronicle Labs, Inc.
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
	"reflect"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestMessageMap_Keys(t *testing.T) {
	tests := []struct {
		name string
		mm   MessageMap
		want []string
	}{
		{
			name: "empty",
			mm:   MessageMap{},
			want: []string{},
		},
		{
			name: "all",
			mm:   AllMessagesMap,
			want: []string{
				"data_point/v1",
				"greet/v1",
				"musig_commitment/v1.1",
				"musig_initialize/v1.1",
				"musig_partial_signature/v1.1",
				"musig_signature/v1.1",
				"musig_terminate/v1.1",
				"price/v0",
				"price/v1",
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := tt.mm.Keys(); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("Keys() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestMessageMap_SelectByTopic(t *testing.T) {
	tests := []struct {
		name    string
		mm      MessageMap
		topics  []string
		want    MessageMap
		wantErr bool
		errMsg  string
	}{
		{
			name:   "empty",
			mm:     MessageMap{},
			topics: []string{},
			want:   MessageMap{},
		},
		{
			name: "all",
			mm:   AllMessagesMap,
			topics: []string{
				"data_point/v1",
				"greet/v1",
				"musig_commitment/v1.1",
				"musig_initialize/v1.1",
				"musig_partial_signature/v1.1",
				"musig_signature/v1.1",
				"musig_terminate/v1.1",
				"price/v0",
				"price/v1",
			},
			want: AllMessagesMap,
		},
		{
			name: "some",
			mm:   AllMessagesMap,
			topics: []string{
				"data_point/v1",
			},
			want: MessageMap{
				"data_point/v1": (*DataPoint)(nil),
			},
		},
		{
			name: "none",
			mm:   AllMessagesMap,
			topics: []string{
				"foo",
			},
			wantErr: true,
			errMsg:  "key not present: foo",
		},
		{
			name: "one of two",
			mm:   AllMessagesMap,
			topics: []string{
				"foo",
				"data_point/v1",
			},
			wantErr: true,
			errMsg:  "key not present: foo",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := tt.mm.SelectByTopic(tt.topics...)
			if (err != nil) != tt.wantErr {
				t.Errorf("SelectByTopic() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if tt.wantErr {
				assert.EqualError(t, err, tt.errMsg)
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("SelectByTopic() got = %v, want %v", got, tt.want)
			}
		})
	}
}
