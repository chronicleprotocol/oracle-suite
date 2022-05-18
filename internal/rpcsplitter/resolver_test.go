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

package rpcsplitter

import (
	"errors"
	"fmt"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func Test_defaultResolver_resolve(t *testing.T) {
	tests := []struct {
		resps        []interface{}
		minResponses int
		want         interface{}
		wantErr      bool
	}{
		{
			resps:        []interface{}{newJSON(`"a"`)},
			minResponses: 1,
			want:         newJSON(`"a"`),
		},
		{
			resps:        []interface{}{newJSON(`"a"`), newJSON(`"a"`), newJSON(`"b"`)},
			minResponses: 1,
			want:         newJSON(`"a"`),
		},
		{
			resps:        []interface{}{newJSON(`"a"`), newJSON(`"a"`), newJSON(`"b"`)},
			minResponses: 2,
			want:         newJSON(`"a"`),
		},
		{
			resps:        []interface{}{newJSON(`"b"`), newJSON(`"a"`), newJSON(`"a"`)},
			minResponses: 2,
			want:         newJSON(`"a"`),
		},
		{
			resps:        []interface{}{newJSON(`"a"`), errors.New("err")},
			minResponses: 1,
			want:         newJSON(`"a"`),
		},
		{
			resps:        []interface{}{newJSON(`"a"`), newJSON(`"a"`), errors.New("err"), errors.New("err")},
			minResponses: 2,
			want:         newJSON(`"a"`),
		},
		{
			resps:        []interface{}{newJSON(`"a"`), newJSON(`"a"`), errors.New("err")},
			minResponses: 3,
			wantErr:      true,
		},
		{
			resps:        []interface{}{newJSON(`"a"`), newJSON(`"b"`)},
			minResponses: 1,
			wantErr:      true,
		},
		{
			resps:        []interface{}{newJSON(`"a"`), newJSON(`"a"`), newJSON(`"b"`), newJSON(`"b"`)},
			minResponses: 2,
			wantErr:      true,
		},
		{
			resps:        []interface{}{newJSON(`"a"`), newJSON(`"a"`), errors.New("err")},
			minResponses: 3,
			wantErr:      true,
		},
	}
	for n, tt := range tests {
		t.Run(fmt.Sprintf("case-%d", n), func(t *testing.T) {
			r := defaultResolver{minResponses: tt.minResponses}
			v, err := r.resolve(tt.resps)
			if tt.wantErr {
				require.Error(t, err)
				return
			}
			assert.Equal(t, tt.want, v)
		})
	}
}

func Test_gasValueResolver_resolve(t *testing.T) {
	tests := []struct {
		resps        []interface{}
		minResponses int
		want         interface{}
		wantErr      bool
	}{
		{
			resps:        []interface{}{newNumber(`0x1`)},
			minResponses: 1,
			want:         newNumber(`0x1`),
		},
		{
			resps:        []interface{}{newNumber(`0x1`), newNumber(`0x2`)},
			minResponses: 1,
			want:         newNumber(`0x1`),
		},
		{
			resps:        []interface{}{newNumber(`0x1`), newNumber(`0x2`)},
			minResponses: 2,
			want:         newNumber(`0x1`),
		},
		{
			resps:        []interface{}{newNumber(`0x2`), newNumber(`0x1`)},
			minResponses: 2,
			want:         newNumber(`0x1`),
		},
		{
			resps:        []interface{}{newNumber(`0x1`), newNumber(`0x2`), newNumber(`0x3`)},
			minResponses: 3,
			want:         newNumber(`0x2`),
		},
		{
			resps:        []interface{}{newNumber(`0x1`), newNumber(`0x2`), newNumber(`0x3`), newNumber(`0x4`)},
			minResponses: 4,
			want:         newNumber(`0x2`),
		},
		{
			resps:        []interface{}{newNumber(`0x2`), newNumber(`0x2`), newNumber(`0x4`), newNumber(`0x4`)},
			minResponses: 4,
			want:         newNumber(`0x3`),
		},
		{
			resps:        []interface{}{newNumber(`0x1`), newNumber(`0x2`), errors.New("err")},
			minResponses: 2,
			want:         newNumber(`0x1`),
		},
		{
			resps:        []interface{}{newNumber(`0x1`), errors.New("err"), errors.New("err")},
			minResponses: 2,
			wantErr:      true,
		},
	}
	for n, tt := range tests {
		t.Run(fmt.Sprintf("case-%d", n), func(t *testing.T) {
			r := gasValueResolver{minResponses: tt.minResponses}
			v, err := r.resolve(tt.resps)
			if tt.wantErr {
				require.Error(t, err)
				return
			}
			assert.Equal(t, tt.want, v)
		})
	}
}

func Test_blockNumberResolver_resolve(t *testing.T) {
	tests := []struct {
		resps           []interface{}
		minResponses    int
		maxBlocksBehind int
		want            interface{}
		wantErr         bool
	}{
		{
			resps:           []interface{}{newNumber(`0x1`)},
			minResponses:    1,
			maxBlocksBehind: 1,
			want:            newNumber(`0x1`),
		},
		{
			resps:           []interface{}{newNumber(`0x1`), newNumber(`0x2`)},
			minResponses:    1,
			maxBlocksBehind: 1,
			want:            newNumber(`0x1`),
		},
		{
			resps:           []interface{}{newNumber(`0x1`), newNumber(`0x2`)},
			minResponses:    1,
			maxBlocksBehind: 0,
			want:            newNumber(`0x2`),
		},
		{
			resps:           []interface{}{newNumber(`0x1`), newNumber(`0x2`), newNumber(`0x3`)},
			minResponses:    3,
			maxBlocksBehind: 1,
			want:            newNumber(`0x2`),
		},
		{
			resps:           []interface{}{newNumber(`0x1`), newNumber(`0x2`), errors.New("err")},
			minResponses:    2,
			maxBlocksBehind: 1,
			want:            newNumber(`0x1`),
		},
		{
			resps:           []interface{}{newNumber(`0x1`), errors.New("err"), errors.New("err")},
			minResponses:    2,
			maxBlocksBehind: 1,
			wantErr:         true,
		},
	}
	for n, tt := range tests {
		t.Run(fmt.Sprintf("case-%d", n), func(t *testing.T) {
			r := blockNumberResolver{minResponses: tt.minResponses, maxBlocksBehind: tt.maxBlocksBehind}
			v, err := r.resolve(tt.resps)
			if tt.wantErr {
				require.Error(t, err)
				return
			}
			assert.Equal(t, tt.want, v)
		})
	}
}
