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

const TeleportEventType = "teleport_evm"
const LoggerTag = "TELEPORT_LISTENER"

// teleportTopic0 is Keccak256("TeleportGUID((bytes32,bytes32,bytes32,bytes32,uint128,uint80,uint48))")
var teleportTopic0 = ethereum.HexToHash("0x9f692a9304834fdefeb4f9cd17d1493600af19c70af547480cccf4a8a4a7752c")

// TeleportListener listens to particular logs on Ethereum compatible blockchain and
// converts them into event messages.
type TeleportListener struct {
	msgCh    chan *messages.Event // List of channels to which messages will be sent.
	listener *logListener
	log      log.Logger
}

// TeleportListenerConfig contains a configuration options for NewTeleportListener.
type TeleportListenerConfig struct {
	// Client is an instance of Ethereum RPC client.
	Client Client
	// Addresses is a list of contracts from which logs will be fetched.
	Addresses []ethereum.Address
	// Interval specifies how often listener should check for new logs.
	Interval time.Duration
	// BlocksDelta specifies the distance between the newest block on the
	// blockchain and the newest block from which logs are to be taken.
	BlocksDelta []int
	// BlocksLimit specifies how from many blocks logs can be fetched at once.
	BlocksLimit int
	// Logger is an instance of a logger. Logger is used mostly to report
	// recoverable errors.
	Logger log.Logger
}

// NewTeleportListener returns a new instance of the TeleportListener struct.
func NewTeleportListener(cfg TeleportListenerConfig) *TeleportListener {
	logger := cfg.Logger.WithField("tag", LoggerTag)
	return &TeleportListener{
		msgCh: make(chan *messages.Event, 1),
		listener: &logListener{
			client:      cfg.Client,
			addresses:   cfg.Addresses,
			topics:      [][]common.Hash{{teleportTopic0}},
			interval:    cfg.Interval,
			blocksDelta: intsToUint64s(cfg.BlocksDelta),
			blocksLimit: uint64(cfg.BlocksLimit),
			logCh:       make(chan types.Log, 1),
			logger:      logger,
		},
		log: logger,
	}
}

// Events implements the publisher.Listener interface.
func (l *TeleportListener) Events() chan *messages.Event {
	return l.msgCh
}

// Start implements the publisher.Listener interface.
func (l *TeleportListener) Start(ctx context.Context) error {
	l.listener.start(ctx)
	go l.listenerRoutine(ctx)
	return nil
}

func (l *TeleportListener) listenerRoutine(ctx context.Context) {
	for {
		select {
		case <-ctx.Done():
			return
		case log := <-l.listener.logs():
			msg, err := logToMessage(log)
			if err != nil {
				l.log.WithError(err).Error("Unable to convert logger to message")
				continue
			}
			l.msgCh <- msg
		}
	}
}

// logToMessage creates a transport message of "event" type from
// given Ethereum logger.
func logToMessage(log types.Log) (*messages.Event, error) {
	guid, err := unpackTeleportGUID(log.Data)
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
		Type: TeleportEventType,
		// ID is additionally hashed to ensure that it is not similar to
		// any other field, so it will not be misused. This field is intended
		// to be used only be the event store.
		ID:          crypto.Keccak256Hash(append(log.TxHash.Bytes(), big.NewInt(int64(log.Index)).Bytes()...)).Bytes(),
		Index:       log.TxHash.Bytes(),
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

// intsToUint64s converts int slice to uint64 slice.
func intsToUint64s(i []int) []uint64 {
	u := make([]uint64, len(i))
	for n, v := range i {
		u[n] = uint64(v)
	}
	return u
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
