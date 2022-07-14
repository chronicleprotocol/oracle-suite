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
	"bytes"
	"context"
	"errors"
	"time"

	"github.com/chronicleprotocol/oracle-suite/pkg/log"
	"github.com/chronicleprotocol/oracle-suite/pkg/starknet"
	"github.com/chronicleprotocol/oracle-suite/pkg/transport/messages"
	"github.com/chronicleprotocol/oracle-suite/pkg/util/retry"
)

const TeleportEventType = "teleport_starknet"
const LoggerTag = "STARKNET_TELEPORT_LISTENER"
const retryAttempts = 10              // The maximum number of attempts to call Sequencer in case of an error.
const retryInterval = 6 * time.Second // The delay between retry attempts.

// Sequencer is a Starknet sequencer.
type Sequencer interface {
	GetPendingBlock(ctx context.Context) (*starknet.Block, error)
	GetLatestBlock(ctx context.Context) (*starknet.Block, error)
	GetBlockByNumber(ctx context.Context, blockNumber uint64) (*starknet.Block, error)
}

// TeleportListenerConfig contains a configuration options for NewTeleportListener.
type TeleportListenerConfig struct {
	// Sequencer is an instance of Ethereum RPC sequencer.
	Sequencer Sequencer
	// Addresses is a list of contracts from which events will be fetched.
	Addresses []*starknet.Felt
	// Interval specifies how often listener should check for new events.
	Interval time.Duration
	// BlocksDelta is a list of distances between the latest accepted block on
	// the blockchain and blocks from which events are to be fetched. If empty,
	// then only events from pending block will be fetched. The purpose of this
	// field is to ensure that older events are resent from time to time.
	// This is to allow other clients on the Oracle network to restore its state
	// and ensure that no events are missed in the event of an Oracle failure.
	BlocksDelta []int
	// BlocksLimit specifies how from many blocks events can be fetched at once.
	BlocksLimit int
	// Logger is an instance of a logger. Logger is used mostly to report
	// recoverable errors.
	Logger log.Logger
}

// TeleportListener listens for TeleportGUID events on Starknet from pending
// blocks and, if BlockDelta is set, also from accepted blocks.
//
// https://github.com/makerdao/dss-teleport
type TeleportListener struct {
	eventCh chan *messages.Event

	// lastBlock is a number of last block from which events were fetched.
	// it is used in the nextBlockRange function.
	lastBlock uint64

	// Configuration parameters copied from TeleportListenerConfig:
	sequencer   Sequencer
	addresses   []*starknet.Felt
	interval    time.Duration
	blocksLimit uint64
	blocksDelta []uint64
	log         log.Logger
}

// NewTeleportListener creates a new instance of TeleportListener.
func NewTeleportListener(cfg TeleportListenerConfig) *TeleportListener {
	return &TeleportListener{
		eventCh:     make(chan *messages.Event),
		sequencer:   cfg.Sequencer,
		addresses:   cfg.Addresses,
		interval:    cfg.Interval,
		blocksLimit: uint64(cfg.BlocksLimit),
		blocksDelta: intsToUint64s(cfg.BlocksDelta),
		log:         cfg.Logger.WithField("tag", LoggerTag),
	}
}

// Events implements the publisher.Listener interface.
func (tl *TeleportListener) Events() chan *messages.Event {
	return tl.eventCh
}

// Start implements the publisher.Listener interface.
func (tl *TeleportListener) Start(ctx context.Context) error {
	go tl.fetchEventsRoutine(ctx)
	return nil
}

// fetchEventsRoutine periodically fetches TeleportGUID events from the
// blockchain.
func (tl *TeleportListener) fetchEventsRoutine(ctx context.Context) {
	t := time.NewTicker(tl.interval)
	defer t.Stop()
	for {
		select {
		case <-ctx.Done():
			close(tl.eventCh)
			return
		case <-t.C:
			if len(tl.blocksDelta) > 0 {
				// As explained in TeleportListenerConfig, blockDelta cannot be
				// empty to fetch events from accepted blocks.
				tl.processAcceptedBlocks(ctx)
			}
			tl.processPendingBlock(ctx)
		}
	}
}

// processAcceptedBlocks fetches TeleportGUID events from accepted blocks and
// converts them into event messages. Converted messages are sent to the
// eventCh channel.
func (tl *TeleportListener) processAcceptedBlocks(ctx context.Context) {
	from, to, err := tl.nextBlockRange(ctx)
	if err != nil {
		tl.log.
			WithError(err).
			Error("Unable to get latest block")
		return
	}
	if from == tl.lastBlock {
		return // There is no new blocks to fetch.
	}
	for _, delta := range tl.blocksDelta {
		if delta > from {
			delta = from // To prevent overflow.
		}
		for num := from - delta; num <= to-delta; num++ {
			if ctx.Err() != nil {
				return
			}
			tl.log.
				WithField("blockNumber", num).
				Info("Fetching block")
			block, err := tl.getBlockByNumber(ctx, num)
			if errors.Is(err, context.Canceled) {
				continue
			}
			if err != nil {
				tl.log.
					WithError(err).
					Error("Unable to fetch block")
				continue
			}
			tl.processBlock(block)
		}
	}
	tl.lastBlock = to
}

// processPendingBlock fetches TeleportGUID events from pending block and
// converts them into event messages. Converted messages are sent to the
// eventCh channel.
func (tl *TeleportListener) processPendingBlock(ctx context.Context) {
	block, err := tl.getPendingBlock(ctx)
	if errors.Is(err, context.Canceled) {
		return
	}
	if err != nil {
		tl.log.
			WithError(err).
			Error("Unable to fetch pending block")
		return
	}
	tl.processBlock(block)
}

// processBlock finds TeleportGUID events in the given block and converts them
// into event messages. Converted messages are sent to the eventCh channel.
func (tl *TeleportListener) processBlock(block *starknet.Block) {
	for _, tx := range block.TransactionReceipts {
		for _, evt := range tx.Events {
			if !tl.isTeleportEvent(evt) {
				continue
			}
			msg, err := eventToMessage(block, tx, evt)
			if err != nil {
				tl.log.
					WithError(err).
					Error("Unable to convert event to message")
				continue
			}
			tl.eventCh <- msg
		}
	}
}

// nextBlockRange returns the range of blocks from which logs should be
// fetched.
func (tl *TeleportListener) nextBlockRange(ctx context.Context) (uint64, uint64, error) {
	// Get the latest block number.
	block, err := tl.getLatestBlock(ctx)
	if err != nil {
		return 0, 0, err
	}
	to := block.BlockNumber
	// Set "from" to the next block and check if "from" is greater than "to",
	// if so, then there are no new blocks to fetch.
	from := tl.lastBlock + 1
	if from > to {
		return to, to, nil
	}
	// Limit the number of blocks to fetch.
	if to-from > tl.blocksLimit {
		from = to - tl.blocksLimit + 1
	}
	return from, to, nil
}

// isTeleportEvent checks if the given event was emitted by the Teleport
// gateway.
func (tl *TeleportListener) isTeleportEvent(evt *starknet.Event) bool {
	for _, addr := range tl.addresses {
		if bytes.Equal(evt.FromAddress.Bytes(), addr.Bytes()) {
			return true
		}
	}
	return false
}

func (tl *TeleportListener) getBlockByNumber(ctx context.Context, num uint64) (block *starknet.Block, err error) {
	err = retry.Retry(
		ctx,
		func() error {
			var err error
			block, err = tl.sequencer.GetBlockByNumber(ctx, num)
			return err
		},
		retryAttempts,
		retryInterval,
	)
	return block, err
}

func (tl *TeleportListener) getLatestBlock(ctx context.Context) (block *starknet.Block, err error) {
	err = retry.Retry(
		ctx,
		func() error {
			var err error
			block, err = tl.sequencer.GetLatestBlock(ctx)
			return err
		},
		retryAttempts,
		retryInterval,
	)
	return block, err
}

func (tl *TeleportListener) getPendingBlock(ctx context.Context) (block *starknet.Block, err error) {
	err = retry.Retry(
		ctx,
		func() error {
			var err error
			block, err = tl.sequencer.GetPendingBlock(ctx)
			return err
		},
		retryAttempts,
		retryInterval,
	)
	return block, err
}

// intsToUint64s converts int slice to uint64 slice.
func intsToUint64s(i []int) []uint64 {
	u := make([]uint64, len(i))
	for n, v := range i {
		u[n] = uint64(v)
	}
	return u
}
