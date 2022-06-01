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
	"context"
	"fmt"
	"time"

	"github.com/ethereum/go-ethereum/accounts/abi"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/crypto"

	"github.com/chronicleprotocol/oracle-suite/internal/starknet"
	"github.com/chronicleprotocol/oracle-suite/pkg/log"
	"github.com/chronicleprotocol/oracle-suite/pkg/transport/messages"
)

const WormholeStarknetEventType = "wormhole_starknet"

// WormholeListener listens to particular logs on Ethereum compatible blockchain and
// converts them into event messages.
type WormholeListener struct {
	msgCh    chan *messages.Event // List of channels to which messages will be sent.
	listener *eventListener
	log      log.Logger
}

// WormholeListenerConfig contains a configuration options for NewWormholeListener.
type WormholeListenerConfig struct {
	// Client is an instance of Ethereum RPC sequencer.
	Client Sequencer
	// Addresses is a list of contracts from which events will be fetched.
	Addresses []*starknet.Felt
	// Interval specifies how often listener should check for new events.
	Interval time.Duration
	// BlocksBehind specifies the distance between the newest block on the
	// blockchain and the newest block from which logs are to be taken. This
	// parameter can be used to ensure sufficient block confirmations.
	BlocksBehind int
	// MaxBlocks specifies how from many blocks logs can be fetched at once.
	MaxBlocks int
	// Logger is an instance of a logger. Logger is used mostly to report
	// recoverable errors.
	Logger log.Logger
}

func NewWormholeListener(cfg WormholeListenerConfig) *WormholeListener {
	return &WormholeListener{
		msgCh: make(chan *messages.Event, 1),
		listener: newEventListener(
			cfg.Client,
			cfg.Addresses,
			cfg.Interval,
			uint64(cfg.BlocksBehind),
			uint64(cfg.MaxBlocks),
			cfg.Logger,
		),
		log: cfg.Logger,
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
	ch := l.listener.Events()
	for {
		select {
		case <-ctx.Done():
			return
		case evt := <-ch:
			msg, err := eventToMessage(evt)
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
func eventToMessage(evt *event) (*messages.Event, error) {
	guid, err := packWormholeGUID(evt)
	if err != nil {
		return nil, err
	}
	hash := crypto.Keccak256Hash(guid)
	data := map[string][]byte{
		"hash":  hash.Bytes(), // Hash to be used to calculate a signature.
		"event": guid,         // NodeEvent data.
	}
	return &messages.Event{
		Type:        WormholeStarknetEventType,
		ID:          eventUniqueID(evt),
		Index:       evt.txnHash.Bytes(),
		EventDate:   evt.time,
		MessageDate: time.Now(),
		Data:        data,
		Signatures:  map[string]messages.EventSignature{},
	}, nil
}

func eventUniqueID(evt *event) []byte {
	var b []byte
	b = append(b, evt.txnHash.Bytes()...)
	b = append(b, evt.fromAddress.Bytes()...)
	for _, k := range evt.keys {
		b = append(b, k.Bytes()...)
	}
	for _, d := range evt.data {
		b = append(b, d.Bytes()...)
	}
	return crypto.Keccak256Hash(b).Bytes()
}

// packWormholeGUID converts wormholeGUID to ABI encoded data.
func packWormholeGUID(evt *event) ([]byte, error) {
	if len(evt.data) < 7 {
		return nil, fmt.Errorf("invalid number of data items: %d", len(evt.data))
	}
	b, err := abiWormholeGUID.Pack(
		toL1String(evt.data[0]),
		toL1String(evt.data[1]),
		toBytes32(evt.data[2]),
		toBytes32(evt.data[3]),
		evt.data[4].Int,
		evt.data[5].Int,
		evt.data[6].Int,
	)
	if err != nil {
		return nil, fmt.Errorf("unable to pack WormholeGUID: %w", err)
	}
	return b, nil
}

func toL1String(f *starknet.Felt) common.Hash {
	var s common.Hash
	copy(s[:], f.Bytes())
	return s
}

func toBytes32(f *starknet.Felt) common.Hash {
	var s common.Hash
	b := f.Bytes()
	if len(b) > 32 {
		return s
	}
	copy(s[32-len(b):], b)
	return s
}

var abiWormholeGUID abi.Arguments

func init() {
	bytes32, _ := abi.NewType("bytes32", "", nil)
	uint128, _ := abi.NewType("uint128", "", nil)
	uint80, _ := abi.NewType("uint128", "", nil)
	uint48, _ := abi.NewType("uint48", "", nil)
	abiWormholeGUID = abi.Arguments{
		{Type: bytes32}, // sourceDomain
		{Type: bytes32}, // targetDomain
		{Type: bytes32}, // receiver
		{Type: bytes32}, // operator
		{Type: uint128}, // amount
		{Type: uint80},  // nonce
		{Type: uint48},  // timestamp
	}
}
