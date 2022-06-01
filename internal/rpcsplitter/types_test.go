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
	"fmt"
	"math"
	"math/big"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func Test_jsonType(t *testing.T) {
	tests := []struct {
		arg     string
		wantNil bool
	}{
		{arg: `{}`, wantNil: false},
		{arg: `[]`, wantNil: false},
		{arg: `1`, wantNil: false},
		{arg: `true`, wantNil: false},
		{arg: `null`, wantNil: false},
		{arg: `{"a": "b"}`, wantNil: false},
		{arg: `["a", "b"]`, wantNil: false},
		{arg: `foo`, wantNil: true},
	}
	for n, tt := range tests {
		t.Run(fmt.Sprintf("case-%d", n+1), func(t *testing.T) {
			v := newJSON(tt.arg)
			if tt.wantNil {
				assert.Nil(t, v)
			} else {
				// Test marshalling:
				j, err := v.MarshalJSON()
				require.NoError(t, err)
				// Test unmarshalling:
				v2 := &jsonType{}
				err = v2.UnmarshalJSON(j)
				require.NoError(t, err)
				// Check if marshalled object is the same as unmarshalled:
				assert.True(t, compare(v, v2))
				assert.JSONEq(t, tt.arg, string(j))
			}
		})
	}
}

func Test_blockNumberType_Unmarshal(t *testing.T) {
	tests := []struct {
		arg        string
		want       *blockIDType
		wantErr    bool
		isTag      bool
		isEarliest bool
		isLatest   bool
		isPending  bool
	}{
		{arg: `"0x0"`, want: (*blockIDType)(big.NewInt(0))},
		{arg: `"0xF"`, want: (*blockIDType)(big.NewInt(15))},
		{arg: `"0"`, want: (*blockIDType)(big.NewInt(0))},
		{arg: `"F"`, want: (*blockIDType)(big.NewInt(15))},
		{arg: `"earliest"`, want: (*blockIDType)(big.NewInt(earliestBlockNumber)), isTag: true, isEarliest: true},
		{arg: `"latest"`, want: (*blockIDType)(big.NewInt(latestBlockNumber)), isTag: true, isLatest: true},
		{arg: `"pending"`, want: (*blockIDType)(big.NewInt(pendingBlockNumber)), isTag: true, isPending: true},
		{arg: `"foo"`, wantErr: true},
		{arg: `"0xZ"`, wantErr: true},
	}
	for n, tt := range tests {
		t.Run(fmt.Sprintf("case-%d", n+1), func(t *testing.T) {
			v := &blockIDType{}
			err := v.UnmarshalJSON([]byte(tt.arg))
			if tt.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
				assert.Equal(t, tt.want, v)
				assert.Equal(t, tt.isTag, v.IsTag())
				assert.Equal(t, tt.isEarliest, v.IsEarliest())
				assert.Equal(t, tt.isLatest, v.IsLatest())
				assert.Equal(t, tt.isPending, v.IsPending())
			}
		})
	}
}

func Test_blockNumberType_Marshal(t *testing.T) {
	tests := []struct {
		arg  *blockIDType
		want string
	}{
		{arg: (*blockIDType)(big.NewInt(0)), want: `"0x0"`},
		{arg: (*blockIDType)(big.NewInt(15)), want: `"0xf"`},
		{arg: (*blockIDType)(big.NewInt(earliestBlockNumber)), want: `"earliest"`},
		{arg: (*blockIDType)(big.NewInt(latestBlockNumber)), want: `"latest"`},
		{arg: (*blockIDType)(big.NewInt(pendingBlockNumber)), want: `"pending"`},
	}
	for n, tt := range tests {
		t.Run(fmt.Sprintf("case-%d", n+1), func(t *testing.T) {
			j, err := tt.arg.MarshalJSON()
			assert.NoError(t, err)
			assert.Equal(t, tt.want, string(j))
		})
	}
}

func Test_numberType_Unmarshal(t *testing.T) {
	tests := []struct {
		arg     string
		want    *numberType
		wantErr bool
	}{
		{arg: `"0x0"`, want: (*numberType)(big.NewInt(0))},
		{arg: `"0xF"`, want: (*numberType)(big.NewInt(15))},
		{arg: `"0"`, want: (*numberType)(big.NewInt(0))},
		{arg: `"F"`, want: (*numberType)(big.NewInt(15))},
		{arg: `"foo"`, wantErr: true},
		{arg: `"0xZ"`, wantErr: true},
	}
	for n, tt := range tests {
		t.Run(fmt.Sprintf("case-%d", n+1), func(t *testing.T) {
			v := &numberType{}
			err := v.UnmarshalJSON([]byte(tt.arg))
			if tt.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
				assert.Equal(t, tt.want, v)
			}
		})
	}
}

func Test_numberType_Marshal(t *testing.T) {
	tests := []struct {
		arg  *numberType
		want string
	}{
		{arg: (*numberType)(big.NewInt(0)), want: `"0x0"`},
		{arg: (*numberType)(big.NewInt(15)), want: `"0xf"`},
	}
	for n, tt := range tests {
		t.Run(fmt.Sprintf("case-%d", n+1), func(t *testing.T) {
			j, err := tt.arg.MarshalJSON()
			assert.NoError(t, err)
			assert.Equal(t, tt.want, string(j))
		})
	}
}

func Test_bytesType_Unmarshal(t *testing.T) {
	tests := []struct {
		arg     string
		want    bytesType
		wantErr bool
	}{
		{arg: `"0xDEADBEEF"`, want: (bytesType)([]byte{0xDE, 0xAD, 0xBE, 0xEF})},
		{arg: `"DEADBEEF"`, want: (bytesType)([]byte{0xDE, 0xAD, 0xBE, 0xEF})},
		{arg: `"0x"`, want: (bytesType)([]byte{})},
		{arg: `""`, want: (bytesType)([]byte{})},
		{arg: `"0x0"`, wantErr: true},
		{arg: `"foo"`, wantErr: true},
		{arg: `"0xZZ"`, wantErr: true},
	}
	for n, tt := range tests {
		t.Run(fmt.Sprintf("case-%d", n+1), func(t *testing.T) {
			v := &bytesType{}
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

func Test_bytesType_Marshal(t *testing.T) {
	tests := []struct {
		arg  bytesType
		want string
	}{
		{arg: (bytesType)([]byte{0xDE, 0xAD, 0xBE, 0xEF}), want: `"0xdeadbeef"`},
		{arg: (bytesType)([]byte{}), want: `"0x"`},
	}
	for n, tt := range tests {
		t.Run(fmt.Sprintf("case-%d", n+1), func(t *testing.T) {
			j, err := tt.arg.MarshalJSON()
			assert.NoError(t, err)
			assert.Equal(t, tt.want, string(j))
		})
	}
}

func Test_addressType_Unmarshal(t *testing.T) {
	tests := []struct {
		arg     string
		want    addressType
		wantErr bool
	}{
		{
			arg:  `"0x00112233445566778899aabbccddeeff00112233"`,
			want: (addressType)([addressLength]byte{0x00, 0x11, 0x22, 0x33, 0x44, 0x55, 0x66, 0x77, 0x88, 0x99, 0xaa, 0xbb, 0xcc, 0xdd, 0xee, 0xff, 0x00, 0x11, 0x22, 0x33}),
		},
		{
			arg:  `"00112233445566778899aabbccddeeff00112233"`,
			want: (addressType)([addressLength]byte{0x00, 0x11, 0x22, 0x33, 0x44, 0x55, 0x66, 0x77, 0x88, 0x99, 0xaa, 0xbb, 0xcc, 0xdd, 0xee, 0xff, 0x00, 0x11, 0x22, 0x33}),
		},
		{
			arg:     `"00112233445566778899aabbccddeeff0011223344"`,
			wantErr: true,
		},
		{
			arg:     `"0x00112233445566778899aabbccddeeff0011223344"`,
			wantErr: true,
		},
		{
			arg:     `"0x00112233445566778899aabbccddeeff001122"`,
			wantErr: true,
		},
		{
			arg:     `"""`,
			wantErr: true,
		},
	}
	for n, tt := range tests {
		t.Run(fmt.Sprintf("case-%d", n+1), func(t *testing.T) {
			v := &addressType{}
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

func Test_addressType_Marshal(t *testing.T) {
	tests := []struct {
		arg  addressType
		want string
	}{
		{
			arg:  (addressType)([addressLength]byte{0x00, 0x11, 0x22, 0x33, 0x44, 0x55, 0x66, 0x77, 0x88, 0x99, 0xaa, 0xbb, 0xcc, 0xdd, 0xee, 0xff, 0x00, 0x11, 0x22, 0x33}),
			want: `"0x00112233445566778899aabbccddeeff00112233"`,
		},
		{
			arg:  addressType{},
			want: `"0x0000000000000000000000000000000000000000"`,
		},
	}
	for n, tt := range tests {
		t.Run(fmt.Sprintf("case-%d", n+1), func(t *testing.T) {
			j, err := tt.arg.MarshalJSON()
			assert.NoError(t, err)
			assert.Equal(t, tt.want, string(j))
		})
	}
}

func Test_addressesType_Unmarshal(t *testing.T) {
	tests := []struct {
		arg     string
		want    addressesType
		wantErr bool
	}{
		{
			arg:  `"0x00112233445566778899aabbccddeeff00112233"`,
			want: (addressesType)([]addressType{{0x00, 0x11, 0x22, 0x33, 0x44, 0x55, 0x66, 0x77, 0x88, 0x99, 0xaa, 0xbb, 0xcc, 0xdd, 0xee, 0xff, 0x00, 0x11, 0x22, 0x33}}),
		},
		{
			arg:  `"00112233445566778899aabbccddeeff00112233"`,
			want: (addressesType)([]addressType{{0x00, 0x11, 0x22, 0x33, 0x44, 0x55, 0x66, 0x77, 0x88, 0x99, 0xaa, 0xbb, 0xcc, 0xdd, 0xee, 0xff, 0x00, 0x11, 0x22, 0x33}}),
		},
		{
			arg: `["0x00112233445566778899aabbccddeeff00112233", "0x00112233445566778899aabbccddeeff00112233"]`,
			want: (addressesType)([]addressType{
				{0x00, 0x11, 0x22, 0x33, 0x44, 0x55, 0x66, 0x77, 0x88, 0x99, 0xaa, 0xbb, 0xcc, 0xdd, 0xee, 0xff, 0x00, 0x11, 0x22, 0x33},
				{0x00, 0x11, 0x22, 0x33, 0x44, 0x55, 0x66, 0x77, 0x88, 0x99, 0xaa, 0xbb, 0xcc, 0xdd, 0xee, 0xff, 0x00, 0x11, 0x22, 0x33},
			}),
		},
		{
			arg:     `"00112233445566778899aabbccddeeff0011223344"`,
			wantErr: true,
		},
		{
			arg:     `"0x00112233445566778899aabbccddeeff0011223344"`,
			wantErr: true,
		},
		{
			arg:     `"0x00112233445566778899aabbccddeeff001122"`,
			wantErr: true,
		},
		{
			arg:     `"""`,
			wantErr: true,
		},
	}
	for n, tt := range tests {
		t.Run(fmt.Sprintf("case-%d", n+1), func(t *testing.T) {
			v := &addressesType{}
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

func Test_addressesType_Marshal(t *testing.T) {
	tests := []struct {
		arg  addressesType
		want string
	}{
		{
			arg:  (addressesType)([]addressType{{0x00, 0x11, 0x22, 0x33, 0x44, 0x55, 0x66, 0x77, 0x88, 0x99, 0xaa, 0xbb, 0xcc, 0xdd, 0xee, 0xff, 0x00, 0x11, 0x22, 0x33}}),
			want: `["0x00112233445566778899aabbccddeeff00112233"]`,
		},
		{
			arg:  addressesType{},
			want: `[]`,
		},
	}
	for n, tt := range tests {
		t.Run(fmt.Sprintf("case-%d", n+1), func(t *testing.T) {
			j, err := tt.arg.MarshalJSON()
			assert.NoError(t, err)
			assert.Equal(t, tt.want, string(j))
		})
	}
}

func Test_hashType_Unmarshal(t *testing.T) {
	tests := []struct {
		arg     string
		want    hashType
		wantErr bool
	}{
		{
			arg:  `"0x00112233445566778899aabbccddeeff00112233445566778899aabbccddeeff"`,
			want: (hashType)([hashLength]byte{0x00, 0x11, 0x22, 0x33, 0x44, 0x55, 0x66, 0x77, 0x88, 0x99, 0xaa, 0xbb, 0xcc, 0xdd, 0xee, 0xff, 0x00, 0x11, 0x22, 0x33, 0x44, 0x55, 0x66, 0x77, 0x88, 0x99, 0xaa, 0xbb, 0xcc, 0xdd, 0xee, 0xff}),
		},
		{
			arg:  `"00112233445566778899aabbccddeeff00112233445566778899aabbccddeeff"`,
			want: (hashType)([hashLength]byte{0x00, 0x11, 0x22, 0x33, 0x44, 0x55, 0x66, 0x77, 0x88, 0x99, 0xaa, 0xbb, 0xcc, 0xdd, 0xee, 0xff, 0x00, 0x11, 0x22, 0x33, 0x44, 0x55, 0x66, 0x77, 0x88, 0x99, 0xaa, 0xbb, 0xcc, 0xdd, 0xee, 0xff}),
		},
		{
			arg:     `"00112233445566778899aabbccddeeff00112233445566778899aabbccddeeff00"`,
			wantErr: true,
		},
		{
			arg:     `"0x00112233445566778899aabbccddeeff00112233445566778899aabbccddeeff00"`,
			wantErr: true,
		},
		{
			arg:     `"0x00112233445566778899aabbccddeeff00112233445566778899aabbccddee"`,
			wantErr: true,
		},
		{
			arg:     `"""`,
			wantErr: true,
		},
	}
	for n, tt := range tests {
		t.Run(fmt.Sprintf("case-%d", n+1), func(t *testing.T) {
			v := &hashType{}
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

func Test_hashType_Marshal(t *testing.T) {
	tests := []struct {
		arg  hashType
		want string
	}{
		{
			arg:  (hashType)([hashLength]byte{0x00, 0x11, 0x22, 0x33, 0x44, 0x55, 0x66, 0x77, 0x88, 0x99, 0xaa, 0xbb, 0xcc, 0xdd, 0xee, 0xff, 0x00, 0x11, 0x22, 0x33, 0x44, 0x55, 0x66, 0x77, 0x88, 0x99, 0xaa, 0xbb, 0xcc, 0xdd, 0xee, 0xff}),
			want: `"0x00112233445566778899aabbccddeeff00112233445566778899aabbccddeeff"`,
		},
		{
			arg:  hashType{},
			want: `"0x0000000000000000000000000000000000000000000000000000000000000000"`,
		},
	}
	for n, tt := range tests {
		t.Run(fmt.Sprintf("case-%d", n+1), func(t *testing.T) {
			j, err := tt.arg.MarshalJSON()
			assert.NoError(t, err)
			assert.Equal(t, tt.want, string(j))
		})
	}
}

func Test_hashesType_Unmarshal(t *testing.T) {
	tests := []struct {
		arg     string
		want    hashesType
		wantErr bool
	}{
		{
			arg:  `"0x00112233445566778899aabbccddeeff00112233445566778899aabbccddeeff"`,
			want: (hashesType)([]hashType{{0x00, 0x11, 0x22, 0x33, 0x44, 0x55, 0x66, 0x77, 0x88, 0x99, 0xaa, 0xbb, 0xcc, 0xdd, 0xee, 0xff, 0x00, 0x11, 0x22, 0x33, 0x44, 0x55, 0x66, 0x77, 0x88, 0x99, 0xaa, 0xbb, 0xcc, 0xdd, 0xee, 0xff}}),
		},
		{
			arg:  `"00112233445566778899aabbccddeeff00112233445566778899aabbccddeeff"`,
			want: (hashesType)([]hashType{{0x00, 0x11, 0x22, 0x33, 0x44, 0x55, 0x66, 0x77, 0x88, 0x99, 0xaa, 0xbb, 0xcc, 0xdd, 0xee, 0xff, 0x00, 0x11, 0x22, 0x33, 0x44, 0x55, 0x66, 0x77, 0x88, 0x99, 0xaa, 0xbb, 0xcc, 0xdd, 0xee, 0xff}}),
		},
		{
			arg: `["0x00112233445566778899aabbccddeeff00112233445566778899aabbccddeeff", "0x00112233445566778899aabbccddeeff00112233445566778899aabbccddeeff"]`,
			want: (hashesType)([]hashType{
				{0x00, 0x11, 0x22, 0x33, 0x44, 0x55, 0x66, 0x77, 0x88, 0x99, 0xaa, 0xbb, 0xcc, 0xdd, 0xee, 0xff, 0x00, 0x11, 0x22, 0x33, 0x44, 0x55, 0x66, 0x77, 0x88, 0x99, 0xaa, 0xbb, 0xcc, 0xdd, 0xee, 0xff},
				{0x00, 0x11, 0x22, 0x33, 0x44, 0x55, 0x66, 0x77, 0x88, 0x99, 0xaa, 0xbb, 0xcc, 0xdd, 0xee, 0xff, 0x00, 0x11, 0x22, 0x33, 0x44, 0x55, 0x66, 0x77, 0x88, 0x99, 0xaa, 0xbb, 0xcc, 0xdd, 0xee, 0xff},
			}),
		},
		{
			arg:     `"00112233445566778899aabbccddeeff00112233445566778899aabbccddeeff00"`,
			wantErr: true,
		},
		{
			arg:     `"0x00112233445566778899aabbccddeeff00112233445566778899aabbccddeeff00"`,
			wantErr: true,
		},
		{
			arg:     `"0x00112233445566778899aabbccddeeff00112233445566778899aabbccddee"`,
			wantErr: true,
		},
		{
			arg:     `"""`,
			wantErr: true,
		},
	}
	for n, tt := range tests {
		t.Run(fmt.Sprintf("case-%d", n+1), func(t *testing.T) {
			v := &hashesType{}
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

func Test_hashesType_Marshal(t *testing.T) {
	tests := []struct {
		arg  hashesType
		want string
	}{
		{
			arg:  (hashesType)([]hashType{{0x00, 0x11, 0x22, 0x33, 0x44, 0x55, 0x66, 0x77, 0x88, 0x99, 0xaa, 0xbb, 0xcc, 0xdd, 0xee, 0xff, 0x00, 0x11, 0x22, 0x33, 0x44, 0x55, 0x66, 0x77, 0x88, 0x99, 0xaa, 0xbb, 0xcc, 0xdd, 0xee, 0xff}}),
			want: `["0x00112233445566778899aabbccddeeff00112233445566778899aabbccddeeff"]`,
		},
		{
			arg:  hashesType{},
			want: `[]`,
		},
	}
	for n, tt := range tests {
		t.Run(fmt.Sprintf("case-%d", n+1), func(t *testing.T) {
			j, err := tt.arg.MarshalJSON()
			assert.NoError(t, err)
			assert.Equal(t, tt.want, string(j))
		})
	}
}

func Test_bigIntToHex(t *testing.T) {
	tests := []struct {
		arg  uint64
		want []byte
	}{
		{arg: 0, want: []byte("0x0")},
		{arg: 1, want: []byte("0x1")},
		{arg: math.MaxUint64, want: []byte("0xffffffffffffffff")},
	}
	for n, tt := range tests {
		t.Run(fmt.Sprintf("case-%d", n+1), func(t *testing.T) {
			assert.Equal(t, tt.want, bigIntToHex(new(big.Int).SetUint64(tt.arg)))
		})
	}
}

func Test_hexToBigInt(t *testing.T) {
	tests := []struct {
		arg     []byte
		want    uint64
		wantErr bool
	}{
		{arg: []byte("0x0"), want: 0, wantErr: false},
		{arg: []byte("0x1"), want: 1, wantErr: false},
		{arg: []byte("0xffffffffffffffff"), want: math.MaxUint64, wantErr: false},
		{arg: []byte("0"), want: 0, wantErr: false},
		{arg: []byte("1"), want: 1, wantErr: false},
		{arg: []byte("ffffffffffffffff"), want: math.MaxUint64, wantErr: false},
		{arg: []byte("foo"), want: 0, wantErr: true},
	}
	for n, tt := range tests {
		t.Run(fmt.Sprintf("case-%d", n+1), func(t *testing.T) {
			got, err := hexToBigInt(tt.arg)
			if tt.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
				assert.Equal(t, tt.want, got.Uint64())
			}
		})
	}
}
