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

	"github.com/chronicleprotocol/oracle-suite/pkg/ethereumv2/types"
)

func Test_defaultResolver_resolve(t *testing.T) {
	tests := []struct {
		resps        []any
		minResponses int
		want         any
		wantErr      bool
	}{
		{
			resps:        []any{newAny(`"a"`)},
			minResponses: 1,
			want:         newAny(`"a"`),
		},
		{
			resps:        []any{newAny(`"a"`), newAny(`"a"`), newAny(`"b"`)},
			minResponses: 1,
			want:         newAny(`"a"`),
		},
		{
			resps:        []any{newAny(`"a"`), newAny(`"a"`), newAny(`"b"`)},
			minResponses: 2,
			want:         newAny(`"a"`),
		},
		{
			resps:        []any{newAny(`"b"`), newAny(`"a"`), newAny(`"a"`)},
			minResponses: 2,
			want:         newAny(`"a"`),
		},
		{
			resps:        []any{newAny(`"a"`), errors.New("err")},
			minResponses: 1,
			want:         newAny(`"a"`),
		},
		{
			resps:        []any{newAny(`"a"`), newAny(`"a"`), errors.New("err"), errors.New("err")},
			minResponses: 2,
			want:         newAny(`"a"`),
		},
		{
			resps:        []any{newAny(`"a"`), newAny(`"a"`), errors.New("err")},
			minResponses: 3,
			wantErr:      true,
		},
		{
			resps:        []any{newAny(`"a"`), newAny(`"b"`)},
			minResponses: 1,
			wantErr:      true,
		},
		{
			resps:        []any{newAny(`"a"`), newAny(`"a"`), newAny(`"b"`), newAny(`"b"`)},
			minResponses: 2,
			wantErr:      true,
		},
		{
			resps:        []any{newAny(`"a"`), newAny(`"a"`), errors.New("err")},
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
		resps        []any
		minResponses int
		want         any
		wantErr      bool
	}{
		{
			resps:        []any{hexToNumberPtr(`0x1`)},
			minResponses: 1,
			want:         hexToNumberPtr(`0x1`),
		},
		{
			resps:        []any{hexToNumberPtr(`0x1`), hexToNumberPtr(`0x2`)},
			minResponses: 1,
			want:         hexToNumberPtr(`0x1`),
		},
		{
			resps:        []any{hexToNumberPtr(`0x1`), hexToNumberPtr(`0x2`)},
			minResponses: 2,
			want:         hexToNumberPtr(`0x1`),
		},
		{
			resps:        []any{hexToNumberPtr(`0x2`), hexToNumberPtr(`0x1`)},
			minResponses: 2,
			want:         hexToNumberPtr(`0x1`),
		},
		{
			resps:        []any{hexToNumberPtr(`0x1`), hexToNumberPtr(`0x2`), hexToNumberPtr(`0x3`)},
			minResponses: 3,
			want:         hexToNumberPtr(`0x2`),
		},
		{
			resps:        []any{hexToNumberPtr(`0x1`), hexToNumberPtr(`0x2`), hexToNumberPtr(`0x3`), hexToNumberPtr(`0x4`)},
			minResponses: 4,
			want:         hexToNumberPtr(`0x2`),
		},
		{
			resps:        []any{hexToNumberPtr(`0x2`), hexToNumberPtr(`0x2`), hexToNumberPtr(`0x4`), hexToNumberPtr(`0x4`)},
			minResponses: 4,
			want:         hexToNumberPtr(`0x3`),
		},
		{
			resps:        []any{hexToNumberPtr(`0x1`), hexToNumberPtr(`0x2`), errors.New("err")},
			minResponses: 2,
			want:         hexToNumberPtr(`0x1`),
		},
		{
			resps:        []any{hexToNumberPtr(`0x1`), errors.New("err"), errors.New("err")},
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
		resps           []any
		minResponses    int
		maxBlocksBehind int
		want            any
		wantErr         bool
	}{
		{
			resps:           []any{hexToNumberPtr(`0x1`)},
			minResponses:    1,
			maxBlocksBehind: 1,
			want:            hexToNumberPtr(`0x1`),
		},
		{
			resps:           []any{hexToNumberPtr(`0x1`), hexToNumberPtr(`0x2`)},
			minResponses:    1,
			maxBlocksBehind: 1,
			want:            hexToNumberPtr(`0x1`),
		},
		{
			resps:           []any{hexToNumberPtr(`0x1`), hexToNumberPtr(`0x2`)},
			minResponses:    1,
			maxBlocksBehind: 0,
			want:            hexToNumberPtr(`0x2`),
		},
		{
			resps:           []any{hexToNumberPtr(`0x1`), hexToNumberPtr(`0x2`), hexToNumberPtr(`0x3`)},
			minResponses:    3,
			maxBlocksBehind: 1,
			want:            hexToNumberPtr(`0x2`),
		},
		{
			resps:           []any{hexToNumberPtr(`0x1`), hexToNumberPtr(`0x2`), errors.New("err")},
			minResponses:    2,
			maxBlocksBehind: 1,
			want:            hexToNumberPtr(`0x1`),
		},
		{
			resps:           []any{hexToNumberPtr(`0x1`), errors.New("err"), errors.New("err")},
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

func hexToNumberPtr(hex string) *types.Number {
	n := types.HexToNumber(hex)
	return &n
}
