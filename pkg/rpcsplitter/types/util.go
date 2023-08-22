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

package types

import (
	"bytes"
	"encoding/hex"
	"fmt"
	"math/big"
)

// bytesMarshalJSON encodes the given bytes as a JSON string where each byte is
// represented by a two-digit hex number. The hex string is always even-length
// and prefixed with "0x".
func bytesMarshalJSON(input []byte) []byte {
	return naiveQuote(bytesMarshalText(input))
}

// bytesMarshalText encodes the given bytes as a string where each byte is
// represented by a two-digit hex number. The hex string is always even-length
// and prefixed with "0x".
func bytesMarshalText(input []byte) []byte {
	return bytesToHex(input)
}

// bytesUnmarshalJSON decodes the given JSON string where each byte is
// represented by a two-digit hex number. The hex string may be prefixed with
// "0x". If the hex string is odd-length, it is padded with a leading zero.
func bytesUnmarshalJSON(input []byte, output *[]byte) error {
	if bytes.Equal(input, []byte("null")) {
		return nil
	}
	return bytesUnmarshalText(naiveUnquote(input), output)
}

// bytesUnmarshalText decodes the given string where each byte is represented by
// a two-digit hex number. The hex string may be prefixed with "0x". If the hex
// string is odd-length, it is padded with a leading zero.
func bytesUnmarshalText(input []byte, output *[]byte) error {
	var err error
	*output, err = hexToBytes(input)
	return err
}

// fixedBytesUnmarshalJSON works like bytesUnmarshalJSON, but it is designed to
// be used with fixed-size byte arrays. The given byte array must be large
// enough to hold the decoded data.
func fixedBytesUnmarshalJSON(input, output []byte) error {
	if bytes.Equal(input, []byte("null")) {
		return nil
	}
	return fixedBytesUnmarshalText(naiveUnquote(input), output)
}

// fixedBytesUnmarshalText works like bytesUnmarshalText, but it is designed to
// be used with fixed-size byte arrays. The given byte array must be large
// enough to hold the decoded data.
func fixedBytesUnmarshalText(input, output []byte) error {
	data, err := hexToBytes(input)
	if err != nil {
		return err
	}
	if len(data) > len(output) {
		return fmt.Errorf("hex string has length %d, want %d", len(data), len(output))
	}
	copy(output[len(output)-len(data):], data)
	return nil
}

// numberMarshalJSON encodes the given big integer as JSON string where number
// is resented in hexadecimal format. The hex string is prefixed with "0x".
// Negative numbers are prefixed with "-0x".
func numberMarshalJSON(input *big.Int) []byte {
	return naiveQuote(numberMarshalText(input))
}

// numberMarshalText encodes the given big integer as string where number is
// resented in hexadecimal format. The hex string is prefixed with "0x".
// Negative numbers are prefixed with "-0x".
func numberMarshalText(input *big.Int) []byte {
	return bigIntToHex(input)
}

// numberUnmarshalJSON decodes the given JSON string where number is resented in
// hexadecimal format. The hex string may be prefixed with "0x". Negative numbers
// must start with minus sign.
func numberUnmarshalJSON(input []byte, output *big.Int) error {
	return numberUnmarshalText(naiveUnquote(input), output)
}

// numberUnmarshalText decodes the given string where number is resented in
// hexadecimal format. The hex string may be prefixed with "0x". Negative numbers
// must start with minus sign.
func numberUnmarshalText(input []byte, output *big.Int) error {
	data, err := hexToBigInt(input)
	if err != nil {
		return err
	}
	output.Set(data)
	return nil
}

// bigIntToHex returns the hex representation of the given big integer.
// The hex string is prefixed with "0x". Negative numbers are prefixed with
// "-0x".
func bigIntToHex(x *big.Int) []byte {
	sign := x.Sign()
	switch {
	case sign == 0:
		return []byte("0x0")
	case sign > 0:
		return []byte("0x" + x.Text(16))
	default:
		return []byte("-0x" + x.Text(16)[1:])
	}
}

// hexToBigInt returns the big integer representation of the given hex string.
// The hex string may be prefixed with "0x".
func hexToBigInt(h []byte) (*big.Int, error) {
	if bytes.Equal(h, []byte("0x0")) {
		return big.NewInt(0), nil
	}
	isNeg := len(h) > 1 && h[0] == '-'
	if isNeg {
		h = h[1:]
	}
	if has0xPrefix(h) {
		h = h[2:]
	}
	if len(h) == 0 {
		return nil, fmt.Errorf("empty hex string")
	}
	x, ok := new(big.Int).SetString(string(h), 16)
	if !ok {
		return nil, fmt.Errorf("invalid hex string")
	}
	if isNeg {
		x.Neg(x)
	}
	return x, nil
}

// bytesToHex returns the hex representation of the given bytes. The hex string
// is always even-length and prefixed with "0x".
func bytesToHex(b []byte) []byte {
	r := make([]byte, len(b)*2+2)
	copy(r, `0x`)
	hex.Encode(r[2:], b)
	return r
}

// hexToBytes returns the bytes representation of the given hex string.
// Unlike hex.DecodeString, it does not require an even number of digits.
// The hex string may be prefixed with "0x".
func hexToBytes(h []byte) ([]byte, error) {
	if bytes.Equal(h, []byte("0x0")) {
		return []byte{0}, nil
	}
	if has0xPrefix(h) {
		h = h[2:]
	}
	if len(h)%2 != 0 {
		h = append([]byte{'0'}, h...)
	}
	r := make([]byte, len(h)/2)
	_, err := hex.Decode(r, h)
	return r, err
}

// has0xPrefix returns true if the given byte slice starts with "0x".
func has0xPrefix(h []byte) bool {
	return len(h) >= 2 && h[0] == '0' && (h[1] == 'x' || h[1] == 'X')
}

// naiveQuote returns a double-quoted string. It does not perform any escaping.
func naiveQuote(i []byte) []byte {
	b := make([]byte, len(i)+2)
	b[0] = '"'
	b[len(b)-1] = '"'
	copy(b[1:], i)
	return b
}

// naiveUnquote returns the string inside the quotes. It does not perform any
// unescaping.
func naiveUnquote(i []byte) []byte {
	if len(i) >= 2 && i[0] == '"' && i[len(i)-1] == '"' {
		return i[1 : len(i)-1]
	}
	return i
}
