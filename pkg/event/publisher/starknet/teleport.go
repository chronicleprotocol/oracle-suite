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

const TeleportStarknetEventType = "teleport_starknet"

// TeleportListenerConfig contains a configuration options for NewTeleportListener.
type TeleportListenerConfig struct {
	// Sequencer is an instance of Ethereum RPC sequencer.
	Sequencer Sequencer
	// Addresses is a list of contracts from which events will be fetched.
	Addresses []*starknet.Felt
	// Interval specifies how often listener should check for new events.
	Interval time.Duration
	// BlocksDelta specifies the distance between the newest block on the
	// blockchain and the newest block from which events are to be taken.
	BlocksDelta []int
	// BlocksLimit specifies how from many blocks logs can be fetched at once.
	BlocksLimit int
	// Logger is an instance of a logger. Logger is used mostly to report
	// recoverable errors.
	Logger log.Logger
}

// TeleportListener listens to particular logs on Ethereum compatible blockchain and
// converts them into event messages.
type TeleportListener struct {
	listeners []eventListener
	messageCh chan *messages.Event
	eventsCh  chan *event
	log       log.Logger
}

// NewTeleportListener creates a new instance of TeleportListener.
func NewTeleportListener(cfg TeleportListenerConfig) *TeleportListener {
	eventsCh := make(chan *event)
	return &TeleportListener{
		listeners: []eventListener{
			&acceptedBlockListener{
				sequencer:   cfg.Sequencer,
				addresses:   cfg.Addresses,
				interval:    cfg.Interval,
				blocksDelta: intsToUint64s(cfg.BlocksDelta),
				blocksLimit: uint64(cfg.BlocksLimit),
				eventCh:     eventsCh,
				logger:      cfg.Logger,
			},
			&pendingBlockListener{
				sequencer: cfg.Sequencer,
				addresses: cfg.Addresses,
				interval:  cfg.Interval,
				eventsCh:  eventsCh,
				logger:    cfg.Logger,
			},
		},
		messageCh: make(chan *messages.Event, 1),
		eventsCh:  eventsCh,
		log:       cfg.Logger,
	}
}

// Events implements the publisher.Listener interface.
func (l *TeleportListener) Events() chan *messages.Event {
	return l.messageCh
}

// Start implements the publisher.Listener interface.
func (l *TeleportListener) Start(ctx context.Context) error {
	for _, listener := range l.listeners {
		listener.start(ctx)
	}
	go l.listenerRoutine(ctx)
	return nil
}

func (l *TeleportListener) listenerRoutine(ctx context.Context) {
	for {
		select {
		case <-ctx.Done():
			close(l.messageCh)
			close(l.eventsCh)
			return
		case evt := <-l.eventsCh:
			msg, err := eventToMessage(evt)
			if err != nil {
				l.log.WithError(err).Error("Unable to convert log to message")
				continue
			}
			l.messageCh <- msg
		}
	}
}

// eventToMessage converts Starkware event to a transport message.
func eventToMessage(evt *event) (*messages.Event, error) {
	guid, err := packTeleportGUID(evt)
	if err != nil {
		return nil, err
	}
	hash := crypto.Keccak256Hash(guid)
	data := map[string][]byte{
		"hash":  hash.Bytes(), // Hash to be used to calculate a signature.
		"event": guid,         // NodeEvent data.
	}
	return &messages.Event{
		Type:        TeleportStarknetEventType,
		ID:          eventUniqueID(evt),
		Index:       evt.txnHash.Bytes(),
		EventDate:   evt.time,
		MessageDate: time.Now(),
		Data:        data,
		Signatures:  map[string]messages.EventSignature{},
	}, nil
}

// eventUniqueID returns a unique ID for the given event.
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

// packTeleportGUID converts teleportGUID to ABI encoded data.
func packTeleportGUID(evt *event) ([]byte, error) {
	if len(evt.data) < 7 {
		return nil, fmt.Errorf("invalid number of data items: %d", len(evt.data))
	}
	b, err := abiTeleportGUID.Pack(
		toL1String(evt.data[0]),
		toL1String(evt.data[1]),
		toBytes32(evt.data[2]),
		toBytes32(evt.data[3]),
		evt.data[4].Int,
		evt.data[5].Int,
		evt.data[6].Int,
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
