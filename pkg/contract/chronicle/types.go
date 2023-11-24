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

package chronicle

import (
	"fmt"

	goethABI "github.com/defiweb/go-eth/abi"
	"github.com/defiweb/go-eth/types"
)

// FeedIDs contains list of feed IDs.
// A feed ID is the first byte of the feed address.
type FeedIDs [256]bool

// FeedIDsFromAddresses returns a FeedIDs from a list of feed addresses.
func FeedIDsFromAddresses(addresses []types.Address) FeedIDs {
	var fids FeedIDs
	fids.Add(addresses...)
	return fids
}

// FeedIDsFromIDs returns a FeedIDs from a list of feed IDs.
func FeedIDsFromIDs(ids []byte) FeedIDs {
	var fids FeedIDs
	fids.SetFeedIDs(ids)
	return fids
}

// FeedIDs returns a feed IDs as a byte slice.
// Each byte in the slice represents a feed ID.
func (b FeedIDs) FeedIDs() []byte {
	var ids []byte
	for i, v := range b {
		if v {
			ids = append(ids, byte(i))
		}
	}
	return ids
}

// SetFeedIDs sets the feed IDs from a byte slice.
func (b *FeedIDs) SetFeedIDs(ids []byte) {
	for _, id := range ids {
		(*b)[id] = true
	}
}

// Has returns true if the given address is included in the feed IDs.
func (b FeedIDs) Has(address ...types.Address) bool {
	for _, a := range address {
		if !b[a[0]] {
			return false
		}
	}
	return true
}

// Add adds the given address to the feed IDs.
func (b *FeedIDs) Add(address ...types.Address) {
	for _, a := range address {
		b[a[0]] = true
	}
}

// uint256FeedBloomType represents the feedBloom type in the ABI.
// It implements the abi.Type interface.
type uint256FeedBloomType struct{}

// IsDynamic implements the abi.Type interface.
func (b uint256FeedBloomType) IsDynamic() bool {
	return false
}

// CanonicalType implements the abi.Type interface.
func (b uint256FeedBloomType) CanonicalType() string {
	return "uint256"
}

// String implements the abi.Type interface.
func (b uint256FeedBloomType) String() string {
	return "uint256_feedBloom"
}

// Value implements the abi.Type interface.
func (b uint256FeedBloomType) Value() goethABI.Value {
	return new(uint256FeedBloomValue)
}

// uint256FeedBloomValue is the value of the feedBloom type in the ABI.
// It implements the abi.Value interface.
type uint256FeedBloomValue FeedIDs

// IsDynamic implements the abi.Value interface.
func (b uint256FeedBloomValue) IsDynamic() bool {
	return false
}

// EncodeABI implements the abi.Value interface.
func (b uint256FeedBloomValue) EncodeABI() (goethABI.Words, error) {
	w := goethABI.Word{}
	for i := 0; i < 32; i++ {
		for j := 0; j < 8; j++ {
			if b[i*8+j] {
				w[i] |= 1 << uint(j)
			}
		}
	}
	return goethABI.Words{w}, nil
}

// DecodeABI implements the abi.Value interface.
func (b *uint256FeedBloomValue) DecodeABI(words goethABI.Words) (int, error) {
	if len(words) == 0 {
		return 0, fmt.Errorf("abi: cannot decode BytesFlags from empty data")
	}
	for i, v := range words[0] {
		for j := 0; j < 8; j++ {
			b[i*8+j] = v&(1<<uint(j)) != 0
		}
	}
	return 1, nil
}

// MapFrom implements the abi.MapFrom interface.
func (b *uint256FeedBloomValue) MapFrom(_ goethABI.Mapper, src any) error {
	if s, ok := src.(FeedIDs); ok {
		*b = uint256FeedBloomValue(s)
		return nil
	}
	return fmt.Errorf("abi: cannot map %T to %T", src, b)
}

// MapTo implements the abi.MapFrom interface.
func (b *uint256FeedBloomValue) MapTo(_ goethABI.Mapper, dst any) error {
	if s, ok := dst.(*FeedIDs); ok {
		*s = FeedIDs(*b)
		return nil
	}
	return fmt.Errorf("abi: cannot map %T to %T", b, dst)
}

// bytes32StringType represents the string32 type in the ABI.
// The string32 type is a null-terminated string represented as a bytes32.
// It implements the abi.Type interface.
type bytes32StringType struct{}

// IsDynamic implements the abi.Type interface.
func (b bytes32StringType) IsDynamic() bool {
	return false
}

// CanonicalType implements the abi.Type interface.
func (b bytes32StringType) CanonicalType() string {
	return "bytes32"
}

// String implements the abi.Type interface.
func (b bytes32StringType) String() string {
	return "bytes32_string"
}

// Value implements the abi.Type interface.
func (b bytes32StringType) Value() goethABI.Value {
	return new(bytes32StringValue)
}

// bytes32StringValue is the value of the string32 type in the ABI.
// It implements the abi.Value interface.
type bytes32StringValue string

// IsDynamic implements the abi.Value interface.
func (b bytes32StringValue) IsDynamic() bool {
	return false
}

// EncodeABI implements the abi.Value interface.
func (b bytes32StringValue) EncodeABI() (goethABI.Words, error) {
	if len(b) == 0 {
		return goethABI.Words{{}}, nil
	}
	w := goethABI.BytesToWords(stringToBytes32(string(b)))
	if len(w) != 1 {
		return nil, fmt.Errorf("abi: cannot encode %s, must be 32 bytes or less", b)
	}
	return w, nil
}

// DecodeABI implements the abi.Value interface.
func (b *bytes32StringValue) DecodeABI(words goethABI.Words) (int, error) {
	if len(words) == 0 {
		return 0, fmt.Errorf("abi: cannot decode empty data")
	}
	*b = bytes32StringValue(bytes32ToString(words[0][:]))
	return 1, nil
}

// MapFrom implements the abi.MapFrom interface.
func (b *bytes32StringValue) MapFrom(_ goethABI.Mapper, src any) error {
	if s, ok := src.(string); ok {
		*b = bytes32StringValue(s)
		return nil
	}
	return fmt.Errorf("abi: cannot map %T to %T", src, b)
}

// MapTo implements the abi.MapFrom interface.
func (b *bytes32StringValue) MapTo(_ goethABI.Mapper, dst any) error {
	if s, ok := dst.(*string); ok {
		*s = string(*b)
		return nil
	}
	return fmt.Errorf("abi: cannot map %T to %T", b, dst)
}
