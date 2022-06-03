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
	"sync"
	"time"

	"github.com/chronicleprotocol/oracle-suite/internal/starknet"
	"github.com/chronicleprotocol/oracle-suite/internal/util/retry"
	"github.com/chronicleprotocol/oracle-suite/pkg/log"
)

const retryAttempts = 3               // The maximum number of attempts to call EthClient in case of an error.
const retryInterval = 5 * time.Second // The delay between retry attempts.

type Sequencer interface {
	GetPendingBlock(ctx context.Context) (*starknet.Block, error)
	GetLatestBlock(ctx context.Context) (*starknet.Block, error)
	GetBlockByNumber(ctx context.Context, blockNumber uint64) (*starknet.Block, error)
}

// eventListeners listens for events on the blockchain.
type eventListener interface {
	start(ctx context.Context)
	events() chan *event
}

// event represents a starkware event from a transaction recipient.
type event struct {
	txnHash     *starknet.Felt
	fromAddress *starknet.Felt
	keys        []*starknet.Felt
	data        []*starknet.Felt
	time        time.Time
}

// acceptedBlockListener periodically fetches events from accepted blocks.
// It fetches events from blocks that are as far behind the most recent
// block as specified in blocksBehind.
type acceptedBlockListener struct {
	mu sync.Mutex

	// Configuration:
	sequencer    Sequencer
	addresses    []*starknet.Felt // The addresses of contract from which event should be handled.
	interval     time.Duration    // Time interval between pulling events from Sequencer.
	maxBlocks    uint64           // Maximum number of blocks from which logs can be fetched.
	blocksBehind []uint64         // Number of blocks behind the latest one.
	eventsCh     chan *event      // Channel to which events are sent.
	log          log.Logger       // Logger.

	// State:
	lastBlockNumber uint64 // Last block from which events were pulled.
}

// start implements the eventListener interface.
func (l *acceptedBlockListener) start(ctx context.Context) {
	if len(l.blocksBehind) == 0 {
		// If blocksBehind is empty, then there is nothing to do.
		return
	}
	go l.listenerRoutine(ctx)
}

// events implements the eventListener interface.
func (l *acceptedBlockListener) events() chan *event {
	return l.eventsCh
}

// nextBlockRange returns the next block range from which events should be
// fetched. It does not consider the blockBehind parameter.
func (l *acceptedBlockListener) nextBlockRange(ctx context.Context) (uint64, uint64, error) {
	block, err := getLatestBlock(ctx, l.sequencer)
	if err != nil {
		return 0, 0, err
	}

	from := l.lastBlockNumber + 1
	to := block.BlockNumber

	// Cap the number of blocks to fetch.
	if to-from > l.maxBlocks {
		from = to - l.maxBlocks + 1
	}

	return from, to, nil
}

// acceptedBlockEvents fetches events from the blockchain.
func (l *acceptedBlockListener) acceptedBlockEvents(ctx context.Context) (evts []*event, err error) {
	from, to, err := l.nextBlockRange(ctx)
	if err != nil {
		return nil, err
	}

	// There is no new blocks to fetch.
	if from == l.lastBlockNumber {
		return nil, nil
	}

	for blockNumber := from; blockNumber <= to; blockNumber++ {
		for _, blocksBehind := range l.blocksBehind {
			blockNumber := blockNumber - blocksBehind

			// Fetch a block.
			l.log.WithField("blockNumber", blockNumber).Info("Fetching Starknet block")
			block, err := getBlockByNumber(ctx, l.sequencer, blockNumber)
			if err != nil {
				l.log.WithError(err).Error("Unable to fetch Starknet block")
				continue
			}

			// Handle events from the block.
			for _, tx := range block.TransactionReceipts {
				for _, evt := range tx.Events {
					if isEventFromAddress(evt, l.addresses) {
						evts = append(evts, mapEvent(block, tx, evt))
					}
				}
			}
		}
	}

	l.lastBlockNumber = to

	return evts, nil
}

func (l *acceptedBlockListener) listenerRoutine(ctx context.Context) {
	t := time.NewTicker(l.interval)
	defer t.Stop()

	fetch := func() {
		l.mu.Lock()
		defer l.mu.Unlock()

		// Fetch events and send them to the channel.
		evts, err := l.acceptedBlockEvents(ctx)
		if err != nil {
			l.log.WithError(err).Error("Unable to fetch events")
			return
		}
		for _, evt := range evts {
			l.eventsCh <- evt
		}
	}

	for {
		fetch()
		select {
		case <-ctx.Done():
			return
		case <-t.C:
			fetch()
		}
	}
}

// pendingBlockListener periodically fetches events from the pending block.
type pendingBlockListener struct {
	mu sync.Mutex

	// Configuration:
	sequencer Sequencer
	addresses []*starknet.Felt // The addresses of contract from which event should be handled.
	interval  time.Duration    // Time interval between pulling events from Sequencer.
	eventsCh  chan *event      // Channel to which events are sent.
	log       log.Logger       // Logger.
}

// start implements the eventListener interface.
func (l *pendingBlockListener) start(ctx context.Context) {
	go l.listenerRoutine(ctx)
}

// events implements the eventListener interface.
func (l *pendingBlockListener) events() chan *event {
	return l.eventsCh
}

// pendingBlockEvents fetches events from the blockchain.
func (l *pendingBlockListener) pendingBlockEvents(ctx context.Context) (evts []*event, err error) {
	// Fetch a block.
	block, err := getPendingBlock(ctx, l.sequencer)
	if err != nil {
		l.log.WithError(err).Error("Unable to fetch Starknet block")
		return nil, err
	}

	// Handle events from the block.
	for _, tx := range block.TransactionReceipts {
		for _, evt := range tx.Events {
			if isEventFromAddress(evt, l.addresses) {
				evts = append(evts, mapEvent(block, tx, evt))
			}
		}
	}

	return evts, nil
}

func (l *pendingBlockListener) listenerRoutine(ctx context.Context) {
	t := time.NewTicker(l.interval)
	defer t.Stop()

	fetch := func() {
		l.mu.Lock()
		defer l.mu.Unlock()

		// Fetch events and send them to the channel.
		l.log.Info("Fetching pending Starknet block")
		evts, err := l.pendingBlockEvents(ctx)
		if err != nil {
			l.log.WithError(err).Error("Unable to fetch events")
			return
		}
		for _, evt := range evts {
			l.eventsCh <- evt
		}
	}

	for {
		fetch()
		select {
		case <-ctx.Done():
			return
		case <-t.C:
			fetch()
		}
	}
}

func isEventFromAddress(evt *starknet.Event, addrs []*starknet.Felt) bool {
	for _, addr := range addrs {
		if bytes.Equal(evt.FromAddress.Bytes(), addr.Bytes()) {
			return true
		}
	}
	return false
}

func mapEvent(block *starknet.Block, tx *starknet.TransactionReceipt, evt *starknet.Event) *event {
	return &event{
		txnHash:     tx.TransactionHash,
		fromAddress: evt.FromAddress,
		time:        time.Unix(block.Timestamp, 0),
		keys:        evt.Keys,
		data:        evt.Data,
	}
}

func getBlockByNumber(ctx context.Context, seq Sequencer, num uint64) (block *starknet.Block, err error) {
	err = retry.Retry(
		ctx,
		func() error {
			var err error
			block, err = seq.GetBlockByNumber(ctx, num)
			return err
		},
		retryAttempts,
		retryInterval,
	)
	return block, err
}

func getLatestBlock(ctx context.Context, seq Sequencer) (block *starknet.Block, err error) {
	err = retry.Retry(
		ctx,
		func() error {
			var err error
			block, err = seq.GetLatestBlock(ctx)
			return err
		},
		retryAttempts,
		retryInterval,
	)
	return block, err
}

func getPendingBlock(ctx context.Context, seq Sequencer) (block *starknet.Block, err error) {
	err = retry.Retry(
		ctx,
		func() error {
			var err error
			block, err = seq.GetPendingBlock(ctx)
			return err
		},
		retryAttempts,
		retryInterval,
	)
	return block, err
}
