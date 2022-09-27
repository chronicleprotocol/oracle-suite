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

package teleportstarknet

import (
	"bytes"
	"context"
	"errors"
	"time"

	"github.com/chronicleprotocol/oracle-suite/pkg/log"
	"github.com/chronicleprotocol/oracle-suite/pkg/log/null"
	"github.com/chronicleprotocol/oracle-suite/pkg/starknet"
	"github.com/chronicleprotocol/oracle-suite/pkg/transport/messages"
	"github.com/chronicleprotocol/oracle-suite/pkg/util/retry"
)

const TeleportEventType = "teleport_starknet"
const LoggerTag = "STARKNET_TELEPORT"
const retryInterval = 6 * time.Second // The delay between retry attempts.

// Sequencer is a Starknet sequencer.
type Sequencer interface {
	GetPendingBlock(ctx context.Context) (*starknet.Block, error)
	GetLatestBlock(ctx context.Context) (*starknet.Block, error)
	GetBlockByNumber(ctx context.Context, blockNumber uint64) (*starknet.Block, error)
}

// Config contains a configuration options for New.
type Config struct {
	// Sequencer is an instance of Ethereum RPC sequencer.
	Sequencer Sequencer
	// Addresses is a list of contracts from which events will be fetched.
	Addresses []*starknet.Felt
	// Interval specifies how often provider should check for new logs.
	Interval time.Duration
	// PrefetchPeriod specifies how far back in time provider should prefetch
	// logs. It is used only during the initial start of the provider.
	PrefetchPeriod time.Duration
	// Logger is an instance of a logger. Logger is used mostly to report
	// recoverable errors.
	Logger log.Logger
}

// EventProvider listens for TeleportGUID events on Starknet.
//
// https://github.com/makerdao/dss-teleport
type EventProvider struct {
	eventCh chan *messages.Event

	// Configuration parameters copied from Config:
	sequencer      Sequencer
	addresses      []*starknet.Felt
	interval       time.Duration
	prefetchPeriod time.Duration
	blockConfirms  uint64
	log            log.Logger

	// Used in tests only:
	disablePrefetchBlocksRoutine bool
	disablePendingBlockRoutine   bool
	disableAcceptedBlocksRoutine bool
}

// New creates a new instance of EventProvider.
func New(cfg Config) (*EventProvider, error) {
	if len(cfg.Addresses) == 0 {
		return nil, errors.New("no addresses provided")
	}
	if cfg.Interval == 0 {
		return nil, errors.New("interval is not set")
	}
	if cfg.Logger == nil {
		cfg.Logger = null.New()
	}
	return &EventProvider{
		eventCh:        make(chan *messages.Event),
		sequencer:      cfg.Sequencer,
		addresses:      cfg.Addresses,
		interval:       cfg.Interval,
		prefetchPeriod: cfg.PrefetchPeriod,
		log:            cfg.Logger.WithField("tag", LoggerTag),
	}, nil
}

// Events implements the publisher.EventPublisher interface.
func (ep *EventProvider) Events() chan *messages.Event {
	return ep.eventCh
}

// Start implements the publisher.EventPublisher interface.
func (ep *EventProvider) Start(ctx context.Context) error {
	if !ep.disablePrefetchBlocksRoutine {
		go ep.prefetchBlocksRoutine(ctx)
	}
	if !ep.disablePendingBlockRoutine {
		go ep.handlePendingBlockRoutine(ctx)
	}
	if !ep.disableAcceptedBlocksRoutine {
		go ep.handleAcceptedBlocksRoutine(ctx)
	}
	return nil
}

// prefetchBlocksRoutine fetches blocks from the past but not older than
// defined in the prefetch period. It is used to fetch logs that were emitted
// before the provider was started.
func (ep *EventProvider) prefetchBlocksRoutine(ctx context.Context) {
	if ep.prefetchPeriod == 0 {
		return
	}
	latestBlock, ok := ep.getLatestBlock(ctx)
	if !ok {
		return // Context wax canceled.
	}
	for bn := latestBlock.BlockNumber; bn > 0 && ctx.Err() == nil; bn-- {
		block, ok := ep.getBlockByNumber(ctx, bn)
		if !ok {
			return // Context wax canceled.
		}
		if time.Since(time.Unix(block.Timestamp, 0)) > ep.prefetchPeriod {
			return
		}
		ep.processBlock(block)
	}
}

// handlePendingBlockRoutine periodically fetches TeleportGUID events from
// the pending block.
func (ep *EventProvider) handlePendingBlockRoutine(ctx context.Context) {
	t := time.NewTicker(ep.interval)
	defer t.Stop()
	for {
		select {
		case <-ctx.Done():
			return
		case <-t.C:
			block, ok := ep.getPendingBlock(ctx)
			if !ok {
				return // Context wax canceled.
			}
			ep.processBlock(block)
		}
	}
}

// handleAcceptedBlocksRoutine periodically fetches TeleportGUID events from
// the accepted blocks.
func (ep *EventProvider) handleAcceptedBlocksRoutine(ctx context.Context) {
	latestBlock, ok := ep.getLatestBlock(ctx)
	if !ok {
		// Context was canceled.
		return
	}
	t := time.NewTicker(ep.interval)
	defer t.Stop()
	for {
		select {
		case <-ctx.Done():
			return
		case <-t.C:
			currentBlock, ok := ep.getLatestBlock(ctx)
			if !ok {
				return // Context was canceled.
			}
			if currentBlock.BlockNumber <= latestBlock.BlockNumber {
				continue
			}
			for bn := latestBlock.BlockNumber + 1; bn <= currentBlock.BlockNumber; bn++ {
				block, ok := ep.getBlockByNumber(ctx, bn)
				if !ok {
					return // Context was canceled.
				}
				ep.processBlock(block)
			}
			latestBlock = currentBlock
		}
	}
}

// processBlock finds TeleportGUID events in the given block and converts them
// into event messages. Converted messages are sent to the eventCh channel.
func (ep *EventProvider) processBlock(block *starknet.Block) {
	for _, tx := range block.TransactionReceipts {
		for _, evt := range tx.Events {
			if !ep.isTeleportEvent(evt) {
				continue
			}
			msg, err := eventToMessage(block, tx, evt)
			if err != nil {
				ep.log.
					WithError(err).
					Error("Unable to convert event to message")
				continue
			}
			ep.eventCh <- msg
		}
	}
}

// isTeleportEvent checks if the given event was emitted by the Teleport
// gateway.
func (ep *EventProvider) isTeleportEvent(evt *starknet.Event) bool {
	for _, addr := range ep.addresses {
		if bytes.Equal(evt.FromAddress.Bytes(), addr.Bytes()) {
			return true
		}
	}
	return false
}

// getBlockByNumber returns a block with the given number.
//
// The method will try to fetch blocks indefinitely in case of an error.
// The only way to stop this method from trying again is to cancel the
// context. In that case, the method will return false as a second return
// value.
func (ep *EventProvider) getBlockByNumber(ctx context.Context, num uint64) (block *starknet.Block, ok bool) {
	retry.TryForever(
		ctx,
		func() error {
			var err error
			block, err = ep.sequencer.GetBlockByNumber(ctx, num)
			return err
		},
		retryInterval,
	)
	return block, ctx.Err() == nil
}

// getLatestBlock returns the latest block.
//
// The method will try to fetch blocks indefinitely in case of an error.
// The only way to stop this method from trying again is to cancel the
// context. In that case, the method will return false as a second return
// value.
func (ep *EventProvider) getLatestBlock(ctx context.Context) (block *starknet.Block, ok bool) {
	retry.TryForever(
		ctx,
		func() error {
			var err error
			block, err = ep.sequencer.GetLatestBlock(ctx)
			return err
		},
		retryInterval,
	)
	return block, ctx.Err() == nil
}

// getPendingBlock returns the pending block.
//
// The method will try to fetch blocks indefinitely in case of an error.
// The only way to stop this method from trying again is to cancel the
// context. In that case, the method will return false as a second return
// value.
func (ep *EventProvider) getPendingBlock(ctx context.Context) (block *starknet.Block, ok bool) {
	retry.TryForever(
		ctx,
		func() error {
			var err error
			block, err = ep.sequencer.GetPendingBlock(ctx)
			return err
		},
		retryInterval,
	)
	return block, ctx.Err() == nil
}
