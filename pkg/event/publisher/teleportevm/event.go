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

package teleportevm

import (
	"fmt"
	"math/big"
	"time"

	"github.com/ethereum/go-ethereum/accounts/abi"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/ethereum/go-ethereum/crypto"

	"github.com/chronicleprotocol/oracle-suite/pkg/transport/messages"
)

// logToMessage converts a TeleportGUID event to a transport message.
func logToMessage(l types.Log) (*messages.Event, error) {
	guid, err := unpackTeleportGUID(l.Data)
	if err != nil {
		return nil, err
	}
	hash, err := guid.hash()
	if err != nil {
		return nil, err
	}
	data := map[string][]byte{
		"hash":  hash.Bytes(), // Hash to be used to calculate a signature.
		"event": l.Data,       // Event data.
	}
	return &messages.Event{
		Type: TeleportEventType,
		// ID is additionally hashed to ensure that it is not similar to
		// any other field, so it will not be misused. This field is intended
		// to be used only be the event store.
		ID:          crypto.Keccak256Hash(append(l.TxHash.Bytes(), big.NewInt(int64(l.Index)).Bytes()...)).Bytes(),
		Index:       l.TxHash.Bytes(),
		EventDate:   time.Unix(guid.timestamp, 0),
		MessageDate: time.Now(),
		Data:        data,
		Signatures:  map[string]messages.EventSignature{},
	}, nil
}

// teleportGUID as defined in:
// https://github.com/makerdao/dss-teleport/blob/master/src/TeleportGUID.sol
type teleportGUID struct {
	sourceDomain common.Hash
	targetDomain common.Hash
	receiver     common.Hash
	operator     common.Hash
	amount       *big.Int
	nonce        *big.Int
	timestamp    int64
}

// hash is used to generate an oracle signature for the TeleportGUID struct.
// It must be compatible with the following contract:
// https://github.com/makerdao/dss-teleport/blob/master/src/TeleportGUID.sol
func (g *teleportGUID) hash() (common.Hash, error) {
	b, err := packTeleportGUID(g)
	if err != nil {
		return common.Hash{}, fmt.Errorf("unable to generate a hash for TeleportGUID: %w", err)
	}
	return crypto.Keccak256Hash(b), nil
}

// packTeleportGUID converts teleportGUID to ABI encoded data.
func packTeleportGUID(g *teleportGUID) ([]byte, error) {
	b, err := abiTeleportGUID.Pack(
		g.sourceDomain,
		g.targetDomain,
		g.receiver,
		g.operator,
		g.amount,
		g.nonce,
		big.NewInt(g.timestamp),
	)
	if err != nil {
		return nil, fmt.Errorf("unable to pack TeleportGUID: %w", err)
	}
	return b, nil
}

// unpackTeleportGUID converts ABI encoded data to teleportGUID.
func unpackTeleportGUID(data []byte) (*teleportGUID, error) {
	u, err := abiTeleportGUID.Unpack(data)
	if err != nil {
		return nil, fmt.Errorf("unable to unpack TeleportGUID: %w", err)
	}
	return &teleportGUID{
		sourceDomain: bytes32ToHash(u[0].([32]uint8)),
		targetDomain: bytes32ToHash(u[1].([32]uint8)),
		receiver:     bytes32ToHash(u[2].([32]uint8)),
		operator:     bytes32ToHash(u[3].([32]uint8)),
		amount:       u[4].(*big.Int),
		nonce:        u[5].(*big.Int),
		timestamp:    u[6].(*big.Int).Int64(),
	}, nil
}

func bytes32ToHash(b [32]uint8) common.Hash {
	return common.BytesToHash(b[:])
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
