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

	"github.com/defiweb/go-eth/abi"
	"github.com/defiweb/go-eth/crypto"
	"github.com/defiweb/go-eth/types"

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
		ID:          crypto.Keccak256(l.TransactionHash.Bytes(), new(big.Int).SetUint64(*l.TransactionIndex).Bytes()).Bytes(),
		Index:       l.TransactionHash.Bytes(),
		EventDate:   time.Unix(guid.Timestamp, 0),
		MessageDate: time.Now(),
		Data:        data,
		Signatures:  map[string]messages.EventSignature{},
	}, nil
}

// teleportGUID as defined in:
// https://github.com/makerdao/dss-teleport/blob/master/src/TeleportGUID.sol
type teleportGUID struct {
	SourceDomain types.Hash `abi:"sourceDomain"`
	TargetDomain types.Hash `abi:"targetDomain"`
	Receiver     types.Hash `abi:"receiver"`
	Operator     types.Hash `abi:"operator"`
	Amount       *big.Int   `abi:"amount"`
	Nonce        *big.Int   `abi:"nonce"`
	Timestamp    int64      `abi:"timestamp"`
}

// hash is used to generate an oracle signature for the TeleportGUID struct.
// It must be compatible with the following contract:
// https://github.com/makerdao/dss-teleport/blob/master/src/TeleportGUID.sol
func (g *teleportGUID) hash() (types.Hash, error) {
	b, err := packTeleportGUID(g)
	if err != nil {
		return types.Hash{}, fmt.Errorf("unable to generate a hash for TeleportGUID: %w", err)
	}
	return crypto.Keccak256(b), nil
}

// packTeleportGUID converts teleportGUID to ABI encoded data.
func packTeleportGUID(guid *teleportGUID) ([]byte, error) {
	b, err := abi.EncodeValue(abiTeleportGUID, guid)
	if err != nil {
		return nil, fmt.Errorf("unable to encode TeleportGUID: %w", err)
	}
	return b, nil
}

// unpackTeleportGUID converts ABI encoded data to teleportGUID.
func unpackTeleportGUID(data []byte) (*teleportGUID, error) {
	x := abiTeleportGUID
	var guid teleportGUID
	if err := abi.DecodeValue(x, data, &guid); err != nil {
		return nil, fmt.Errorf("unable to decode TeleportGUID: %w", err)
	}
	return &guid, nil
}

var abiTeleportGUID = abi.MustParseType(
	`(
		bytes32 sourceDomain, 
		bytes32 targetDomain, 
		bytes32 receiver, 
		bytes32 operator, 
		uint128 amount, 
		uint80 nonce, 
		uint48 timestamp
	)`,
)
