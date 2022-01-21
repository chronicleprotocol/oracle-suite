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

package ethereum

import (
	"context"
	"fmt"
	"math/big"
	"time"

	"github.com/ethereum/go-ethereum/accounts/abi"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/ethereum/go-ethereum/crypto"

	"github.com/chronicleprotocol/oracle-suite/pkg/ethereum"
	"github.com/chronicleprotocol/oracle-suite/pkg/log"
	"github.com/chronicleprotocol/oracle-suite/pkg/transport/messages"
)

const WormholeEventType = "wormhole"

// wormholeTopic0 is Keccak256("WormholeInitialized((bytes32,bytes32,address,address,uint128,uint80,uint48))")
var wormholeTopic0 = ethereum.HexToHash("0x0162851814eed5360ac17d1b3d942e6619fa2d803de71e7159ed9bebf724072a")

// WormholeListener listens to particular logs on Ethereum compatible blockchain and
// converts them into event messages.
type WormholeListener struct {
	msgCh    chan *messages.Event // List of channels to which messages will be sent.
	listener logListener
	log      log.Logger
}

// WormholeListenerConfig contains a configuration options for NewWormholeListener.
type WormholeListenerConfig struct {
	// Ethereum client.
	Client EthClient
	// Addresses is a list of contracts from which logs will be fetched.
	Addresses []ethereum.Address
	// Interval specifies how often listener should check for new logs.
	Interval time.Duration
	// BlocksBehind specifies the distance between the newest block on the
	// blockchain and the newest block from which logs are to be taken. This
	// parameter can be used to ensure sufficient acknowledgements.
	BlocksBehind int
	// MaxBlocks specifies how from many blocks logs can be fetched at once.
	MaxBlocks int
	// Log is an instance of a logger. Logger is used mostly to report
	// recoverable errors.
	Log log.Logger
}

// NewWormholeListener returns a new instance of the WormholeListener struct.
func NewWormholeListener(cfg WormholeListenerConfig) *WormholeListener {
	return &WormholeListener{
		msgCh: make(chan *messages.Event, 128),
		listener: newEthClientLogListener(
			cfg.Client,
			cfg.Addresses,
			[]common.Hash{wormholeTopic0},
			cfg.Interval,
			uint64(cfg.BlocksBehind),
			uint64(cfg.MaxBlocks),
			cfg.Log,
		),
		log: cfg.Log,
	}
}

// Events implements the publisher.Listener interface.
func (l *WormholeListener) Events() chan *messages.Event {
	return l.msgCh
}

// Start implements the publisher.Listener interface.
func (l *WormholeListener) Start(ctx context.Context) error {
	l.listener.Start(ctx)
	go l.listenerRoutine(ctx)
	return nil
}

func (l *WormholeListener) listenerRoutine(ctx context.Context) {
	ch := l.listener.Logs()
	for {
		select {
		case <-ctx.Done():
			return
		case log := <-ch:
			msg, err := logToMessage(log)
			if err != nil {
				l.log.WithError(err).Error("Unable to convert log to message")
				continue
			}
			l.msgCh <- msg
		}
	}
}

// logToMessage creates a transport message of "event" type from
// given Ethereum log.
func logToMessage(log types.Log) (*messages.Event, error) {
	guid, err := unpackWormholeGUID(log.Data)
	if err != nil {
		return nil, err
	}
	hash, err := guid.hash()
	if err != nil {
		return nil, err
	}
	data := map[string][]byte{
		"hash":  hash.Bytes(), // Hash to be used to calculate a signature.
		"event": log.Data,     // Event data.
	}
	return &messages.Event{
		Type: WormholeEventType,
		// ID is additionally hashed to ensure that it is not similar to
		// any other field, so it will not be misused. This field is intended
		// to be used only be the event store.
		ID:         crypto.Keccak256Hash(append(log.TxHash.Bytes(), big.NewInt(int64(log.Index)).Bytes()...)).Bytes(),
		Index:      log.TxHash.Bytes(),
		Date:       time.Now(),
		Data:       data,
		Signatures: map[string][]byte{},
	}, nil
}

// wormholeGUID as defined in:
// https://github.com/makerdao/dss-wormhole/blob/master/src/WormholeGUID.sol
type wormholeGUID struct {
	sourceDomain common.Hash
	targetDomain common.Hash
	receiver     common.Address
	operator     common.Address
	amount       *big.Int
	nonce        *big.Int
	timestamp    int64
}

// hash is used to generate an oracle signature for the WormholeGUID struct.
// It must be compatible with the following contract:
// https://github.com/makerdao/dss-wormhole/blob/master/src/WormholeOracleAuth.sol
func (g *wormholeGUID) hash() (common.Hash, error) {
	b, err := packWormholeGUID(g)
	if err != nil {
		return common.Hash{}, fmt.Errorf("unable to generate a hash for WormholeGUID: %w", err)
	}
	return crypto.Keccak256Hash(b), nil
}

// packWormholeGUID converts wormholeGUID to ABI encoded data.
func packWormholeGUID(g *wormholeGUID) ([]byte, error) {
	b, err := abiWormholeGUID.Pack(
		g.sourceDomain,
		g.targetDomain,
		g.receiver,
		g.operator,
		g.amount,
		g.nonce,
		big.NewInt(g.timestamp),
	)
	if err != nil {
		return nil, fmt.Errorf("unable to pack WormholeGUID: %w", err)
	}
	return b, nil
}

// unpackWormholeGUID converts ABI encoded data to wormholeGUID.
func unpackWormholeGUID(data []byte) (*wormholeGUID, error) {
	u, err := abiWormholeGUID.Unpack(data)
	if err != nil {
		return nil, fmt.Errorf("unable to unpack WormholeGUID: %w", err)
	}
	return &wormholeGUID{
		sourceDomain: bytes32ToHash(u[0].([32]uint8)),
		targetDomain: bytes32ToHash(u[1].([32]uint8)),
		receiver:     u[2].(common.Address),
		operator:     u[3].(common.Address),
		amount:       u[4].(*big.Int),
		nonce:        u[5].(*big.Int),
		timestamp:    u[6].(*big.Int).Int64(),
	}, nil
}

func bytes32ToHash(b [32]uint8) common.Hash {
	return common.BytesToHash(b[:])
}

var abiWormholeGUID abi.Arguments

func init() {
	bytes32, _ := abi.NewType("bytes32", "", nil)
	address, _ := abi.NewType("address", "", nil)
	uint128, _ := abi.NewType("uint128", "", nil)
	uint80, _ := abi.NewType("uint128", "", nil)
	uint48, _ := abi.NewType("uint48", "", nil)
	abiWormholeGUID = abi.Arguments{
		{Type: bytes32}, // sourceDomain
		{Type: bytes32}, // targetDomain
		{Type: address}, // receiver
		{Type: address}, // operator
		{Type: uint128}, // amount
		{Type: uint80},  // nonce
		{Type: uint48},  // timestamp
	}
}
