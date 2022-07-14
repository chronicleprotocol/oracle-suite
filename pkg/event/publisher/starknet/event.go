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
	"time"

	"github.com/ethereum/go-ethereum/accounts/abi"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/crypto"

	"github.com/chronicleprotocol/oracle-suite/pkg/starknet"
	"github.com/chronicleprotocol/oracle-suite/pkg/transport/messages"
)

// eventToMessage converts Starkware event to a transport message.
func eventToMessage(b *starknet.Block, tx *starknet.TransactionReceipt, e *starknet.Event) (*messages.Event, error) {
	guid, err := packTeleportGUID(e)
	if err != nil {
		return nil, err
	}
	hash := crypto.Keccak256Hash(guid)
	data := map[string][]byte{
		"hash":  hash.Bytes(), // Hash to be used to calculate a signature.
		"event": guid,         // Event data.
	}
	return &messages.Event{
		Type:        TeleportEventType,
		ID:          eventUniqueID(tx, e),
		Index:       tx.TransactionHash.Bytes(),
		EventDate:   time.Unix(b.Timestamp, 0),
		MessageDate: time.Now(),
		Data:        data,
		Signatures:  map[string]messages.EventSignature{},
	}, nil
}

// eventUniqueID returns a unique ID for the given event.
func eventUniqueID(tx *starknet.TransactionReceipt, e *starknet.Event) []byte {
	var b []byte
	b = append(b, tx.TransactionHash.Bytes()...)
	b = append(b, e.FromAddress.Bytes()...)
	for _, k := range e.Keys {
		b = append(b, k.Bytes()...)
	}
	for _, d := range e.Data {
		b = append(b, d.Bytes()...)
	}
	return crypto.Keccak256Hash(b).Bytes()
}

// packTeleportGUID converts teleportGUID to ABI encoded data.
func packTeleportGUID(e *starknet.Event) ([]byte, error) {
	if len(e.Data) < 7 {
		return nil, fmt.Errorf("invalid number of data items: %d", len(e.Data))
	}
	b, err := abiTeleportGUID.Pack(
		toL1String(e.Data[0]),
		toL1String(e.Data[1]),
		toBytes32(e.Data[2]),
		toBytes32(e.Data[3]),
		e.Data[4].Int,
		e.Data[5].Int,
		e.Data[6].Int,
	)
	if err != nil {
		return nil, fmt.Errorf("unable to pack TeleportGUID: %w", err)
	}
	return b, nil
}

// toL1String converts a felt value to Ethereum hash.
func toL1String(f *starknet.Felt) common.Hash {
	var s common.Hash
	copy(s[:], f.Bytes())
	return s
}

// toBytes32 converts a felt value to Ethereum hash.
func toBytes32(f *starknet.Felt) common.Hash {
	var s common.Hash
	b := f.Bytes()
	if len(b) > 32 {
		return s
	}
	copy(s[32-len(b):], b)
	return s
}

var abiTeleportGUID abi.Arguments

func init() {
	bytes32, _ := abi.NewType("bytes32", "", nil)
	uint128, _ := abi.NewType("uint128", "", nil)
	uint80, _ := abi.NewType("uint128", "", nil)
	uint48, _ := abi.NewType("uint48", "", nil)
	abiTeleportGUID = abi.Arguments{
		{Type: bytes32}, // sourceDomain
		{Type: bytes32}, // targetDomain
		{Type: bytes32}, // receiver
		{Type: bytes32}, // operator
		{Type: uint128}, // amount
		{Type: uint80},  // nonce
		{Type: uint48},  // timestamp
	}
}
