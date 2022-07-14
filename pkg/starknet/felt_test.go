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

package starknet

import (
	"fmt"
	"testing"

	"github.com/stretchr/testify/assert"
)

func Test_HexToFelt(t *testing.T) {
	tests := []struct {
		arg  string
		want []byte
	}{
		{arg: `0x0`, want: []byte{}},
		{arg: `0x1`, want: []byte{1}},
		{arg: `0xDEADBEEF`, want: []byte{0xDE, 0xAD, 0xBE, 0xEF}},
	}
	for n, tt := range tests {
		t.Run(fmt.Sprintf("case-%d", n+1), func(t *testing.T) {
			assert.Equal(t, tt.want, HexToFelt(tt.arg).Bytes())
		})
	}
}

func Test_Felt_Unmarshal(t *testing.T) {
	tests := []struct {
		arg     string
		want    []byte
		wantErr bool
	}{
		{arg: `""`, want: []byte{}},
		{arg: `"0x0"`, want: []byte{}},
		{arg: `"0x1"`, want: []byte{1}},
		{arg: `"0xDEADBEEF"`, want: []byte{0xDE, 0xAD, 0xBE, 0xEF}},
		{arg: `"0"`, want: []byte{}},
		{arg: `"1"`, want: []byte{1}},
		{arg: `"123456"`, want: []byte{0x1, 0xe2, 0x40}},
		{arg: `0x0`, wantErr: true},
		{arg: `DEADBEEF`, wantErr: true},
	}
	for n, tt := range tests {
		t.Run(fmt.Sprintf("case-%d", n+1), func(t *testing.T) {
			v := &Felt{}
			err := v.UnmarshalJSON([]byte(tt.arg))
			if tt.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
				assert.Equal(t, tt.want, v.Bytes())
			}
		})
	}
}

func Test_Felt_Marshal(t *testing.T) {
	tests := []struct {
		arg  *Felt
		want string
	}{
		{arg: HexToFelt("0x0"), want: `"0x0"`},
		{arg: HexToFelt("0xdeadbeef"), want: `"0xdeadbeef"`},
		{arg: HexToFelt("deadbeef"), want: `"0xdeadbeef"`},
	}
	for n, tt := range tests {
		t.Run(fmt.Sprintf("case-%d", n+1), func(t *testing.T) {
			j, err := tt.arg.MarshalJSON()
			assert.NoError(t, err)
			assert.Equal(t, tt.want, string(j))
		})
	}
}
