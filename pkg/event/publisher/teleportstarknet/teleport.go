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
	"net/http"
	"sync"
	"time"

	"github.com/chronicleprotocol/oracle-suite/pkg/log"
	"github.com/chronicleprotocol/oracle-suite/pkg/log/null"
	"github.com/chronicleprotocol/oracle-suite/pkg/starknet"
	"github.com/chronicleprotocol/oracle-suite/pkg/transport/messages"
	"github.com/chronicleprotocol/oracle-suite/pkg/util/retry"
)

const TeleportEventType = "teleport_starknet"
const LoggerTag = "STARKNET_TELEPORT"

// retryInterval is the interval between retry attempts in case of an error
// while communicating with a Starknet node.
const retryInterval = 5 * time.Second

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
	// Interval specifies how often provider should check for new events.
	Interval time.Duration
	// PrefetchPeriod specifies how far back in time provider should prefetch
	// events. It is used only during the initial start of the provider.
	PrefetchPeriod time.Duration
	// Logger is an instance of a logger. Logger is used mostly to report
	// recoverable errors.
	Logger log.Logger
}

// EventProvider listens for TeleportGUID events on Starknet.
//
// https://github.com/makerdao/dss-teleport
// https://github.com/makerdao/starknet-dai-bridge
//
// It periodically fetches pending block, looks for TeleportGUID events,
// converts them into messages.Event and sends them to the channel provided
// by Events method.
//
// During the initial start of the provider it also fetches older blocks
// until it reaches the block that is older than the prefetch period. This is
// done to fetch events that were emitted before the provider was started.
//
// Finally, it also listens for newly accepted blocks. This is done to make
// sure that provider does not miss any events from the pending block. This
// can happen if the Starknet node becomes unavailable, so it cannot fetch
// the pending block. If at that time the pending block become accepted, the
// events that would have been added since the time the node became unavailable
// would be lost.
//
// In the event of an error in communication with a Starknet node, whether
// related to network errors or the node itself, the provider will try to
// repeat requests to the node indefinitely.
type EventProvider struct {
	mu      sync.Mutex
	eventCh chan *messages.Event

	// Configuration parameters copied from Config:
	sequencer      Sequencer
	addresses      []*starknet.Felt
	interval       time.Duration
	prefetchPeriod time.Duration
	log            log.Logger

	// Fields for tracking transactions from a pending block, used in the
	// processBlock method:
	pendingParent *starknet.Felt
	pendingTxs    []*starknet.Felt

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

// prefetchBlocksRoutine fetches older blocks until it reaches the block that
// is older than the prefetch period. This is done to fetch events that were
// emitted before the provider was started.
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
			return // End of the prefetch period reached.
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
		return // Context was canceled.
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
				continue // There is no new blocks.
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
	ep.mu.Lock()
	defer ep.mu.Unlock()

	ep.log.
		WithFields(log.Fields{
			"blockNumber": block.BlockNumber,
			"status":      block.Status,
		}).
		Info("Processing block")

	isPending := block.Status == "PENDING"

	// Clear list of processed pending transactions if there is a new pending
	// block. New pending block detected when the block has a different parent
	// than the current pending block.
	if isPending && (ep.pendingParent == nil || block.ParentBlockHash.Cmp(ep.pendingParent.Int) != 0) {
		ep.pendingParent = block.ParentBlockHash
		ep.pendingTxs = nil
	}

	for _, tx := range block.TransactionReceipts {
		// Check if transaction from pending block was already processed.
		// Because new transactions are constantly added to the pending block,
		// we need to keep track of processed transactions to avoid duplicates.
		skip := false
		if isPending {
			for _, txHash := range ep.pendingTxs {
				if txHash.Cmp(tx.TransactionHash.Int) == 0 {
					// Transaction was already processed so it should be
					// skipped.
					skip = true
					break
				}
			}
			if !skip {
				ep.pendingTxs = append(ep.pendingTxs, tx.TransactionHash)
			}
		}
		if skip {
			continue
		}

		// Handle TeleportGUID events.
		for _, evt := range tx.Events {
			if !ep.isTeleportEvent(evt) {
				continue
			}
			event, err := eventToMessage(block, tx, evt)
			if err != nil {
				ep.log.
					WithError(err).
					Error("Unable to convert event to message")
				continue
			}
			ep.eventCh <- event
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
			if err, ok := err.(starknet.HTTPError); ok && err.StatusCode == http.StatusTooManyRequests {
				ep.log.WithError(err).Debug("Unable to get block by number")
				return err
			}
			if err != nil {
				ep.log.WithError(err).Error("Unable to get block by number")
			}
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
			if err, ok := err.(starknet.HTTPError); ok && err.StatusCode == http.StatusTooManyRequests {
				ep.log.WithError(err).Debug("Unable to get latest block")
				return err
			}
			if err != nil {
				ep.log.WithError(err).Error("Unable to get latest block")
			}
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
			if err, ok := err.(starknet.HTTPError); ok && err.StatusCode == http.StatusTooManyRequests {
				ep.log.WithError(err).Debug("Unable to get pending block")
				return err
			}
			if err != nil {
				ep.log.WithError(err).Error("Unable to get pending block")
			}
			return err
		},
		retryInterval,
	)
	return block, ctx.Err() == nil
}
