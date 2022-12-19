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

package types

import (
	"fmt"
	"testing"

	"github.com/stretchr/testify/assert"
)

func Test_BytesType_Unmarshal(t *testing.T) {
	tests := []struct {
		arg     string
		want    Bytes
		wantErr bool
	}{
		{arg: `"0xDEADBEEF"`, want: (Bytes)([]byte{0xDE, 0xAD, 0xBE, 0xEF})},
		{arg: `"DEADBEEF"`, want: (Bytes)([]byte{0xDE, 0xAD, 0xBE, 0xEF})},
		{arg: `"0x"`, want: (Bytes)([]byte{})},
		{arg: `""`, want: (Bytes)([]byte{})},
		{arg: `"0x0"`, want: (Bytes)([]byte{0x0})},
		{arg: `"foo"`, wantErr: true},
		{arg: `"0xZZ"`, wantErr: true},
	}
	for n, tt := range tests {
		t.Run(fmt.Sprintf("case-%d", n+1), func(t *testing.T) {
			v := &Bytes{}
			err := v.UnmarshalJSON([]byte(tt.arg))
			if tt.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
				assert.Equal(t, tt.want, *v)
			}
		})
	}
}

func Test_BytesType_Marshal(t *testing.T) {
	tests := []struct {
		arg  Bytes
		want string
	}{
		{arg: (Bytes)([]byte{0xDE, 0xAD, 0xBE, 0xEF}), want: `"0xdeadbeef"`},
		{arg: (Bytes)([]byte{}), want: `"0x"`},
	}
	for n, tt := range tests {
		t.Run(fmt.Sprintf("case-%d", n+1), func(t *testing.T) {
			j, err := tt.arg.MarshalJSON()
			assert.NoError(t, err)
			assert.Equal(t, tt.want, string(j))
		})
	}
}
