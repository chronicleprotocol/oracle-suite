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
	"math/big"
	"testing"

	"github.com/defiweb/go-eth/types"
	"github.com/stretchr/testify/assert"
)

func TestGreet_MarshallBinary(t *testing.T) {
	tests := []struct {
		name    string
		greet   Greet
		wantErr bool
	}{
		{
			name: "valid Greet",
			greet: Greet{
				Signature:  types.MustSignatureFromHex("0x00112233445566778899aabbccddeeff00112233445566778899aabbccddeeff00112233445566778899aabbccddeeff00112233445566778899aabbccddeeff00"),
				PublicKeyX: big.NewInt(1234567890),
				PublicKeyY: big.NewInt(1234567890),
			},
			wantErr: false,
		},
		{
			name: "nil values",
			greet: Greet{
				Signature:  types.MustSignatureFromHex("0x00112233445566778899aabbccddeeff00112233445566778899aabbccddeeff00112233445566778899aabbccddeeff00112233445566778899aabbccddeeff00"),
				PublicKeyX: nil,
				PublicKeyY: nil,
			},
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := tt.greet.MarshallBinary()
			if tt.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func TestGreet_UnmarshallBinary(t *testing.T) {
	tests := []struct {
		name    string
		bytes   []byte
		wantErr bool
	}{
		{
			name: "valid bytes",
			bytes: func() []byte {
				greet := Greet{
					Signature:  types.MustSignatureFromHex("0x00112233445566778899aabbccddeeff00112233445566778899aabbccddeeff00112233445566778899aabbccddeeff00112233445566778899aabbccddeeff00"),
					PublicKeyX: big.NewInt(1234567890),
					PublicKeyY: big.NewInt(1234567890),
				}
				bytes, _ := greet.MarshallBinary()
				return bytes
			}(),
			wantErr: false,
		},
		{
			name:    "invalid bytes",
			bytes:   []byte("invalid bytes"),
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var greet Greet
			err := greet.UnmarshallBinary(tt.bytes)
			if tt.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func FuzzGreet_UnmarshallBinary(f *testing.F) {
	f.Fuzz(func(t *testing.T, data []byte) {
		_ = (&Greet{}).UnmarshallBinary(data)
	})
}
