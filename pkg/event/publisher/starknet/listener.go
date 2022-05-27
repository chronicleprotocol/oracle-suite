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
	"math"
	"sync"
	"time"

	"github.com/chronicleprotocol/oracle-suite/internal/starknet"
	"github.com/chronicleprotocol/oracle-suite/pkg/log"
)

type Client interface {
	BlockNumber(ctx context.Context) (uint64, error)
	GetBlockByNumber(ctx context.Context, blockNumber uint64, scope starknet.Scope) (*starknet.Block, error)
}

type event struct {
	txnHash     *starknet.Felt
	fromAddress *starknet.Felt
	time        time.Time
	keys        []*starknet.Felt
	data        []*starknet.Felt
}

type eventListener struct {
	mu sync.Mutex

	client          Client
	addresses       []*starknet.Felt
	interval        time.Duration // Time interval between pulling logs from Ethereum client.
	lastBlockNumber uint64        // Last block from which logs were pulled.
	blocksBehind    uint64        // Number of blocks behind the latest one.
	maxBlocks       uint64        // Maximum number of blocks from which logs can be fetched.
	outCh           chan *event   // Channel to which events are sent.
	log             log.Logger    // Logger.
}

func newEventListener(client Client, addresses []*starknet.Felt, interval time.Duration, blocksBehind, maxBlocks uint64, logger log.Logger) *eventListener {
	return &eventListener{
		client:       client,
		addresses:    addresses,
		interval:     interval,
		blocksBehind: blocksBehind,
		maxBlocks:    maxBlocks,
		outCh:        make(chan *event),
		log:          logger,
	}
}

// Start implements the logListener interface.
func (l *eventListener) Start(ctx context.Context) {
	go l.listenerRoutine(ctx)
}

func (l *eventListener) Events() chan *event {
	return l.outCh
}

// nextBlockNumberRange returns the next block range from which logs should
// be fetched.
func (l *eventListener) nextBlockNumberRange(ctx context.Context) (uint64, uint64, error) {
	curr, err := l.client.BlockNumber(ctx)
	if err != nil {
		return 0, 0, err
	}
	from := l.lastBlockNumber + 1
	to := uint64(math.Max(0, float64(curr-l.blocksBehind)))
	if from > to {
		from = to
	}
	if to-from > l.maxBlocks {
		from = to - l.maxBlocks + 1
	}
	return from, to, nil
}

// nextTransactions returns a logs from a range returned by the nextBlockNumberRange
// method and updates lastBlockNumber variable.
func (l *eventListener) nextTransactions(ctx context.Context) ([]*event, error) {
	// Find the next block number range.
	from, to, err := l.nextBlockNumberRange(ctx)
	if err != nil {
		return nil, err
	}

	// If the "from" var is equal to the last block number, it means that there
	// were no new block since the last invoke of this method, so there are no
	// new logs.
	if from == l.lastBlockNumber {
		return nil, nil
	}

	// Fetch transactions from a block.
	var all []*event
	for n := from; n <= to; n++ {
		if l.log.Level() >= log.Debug {
			l.log.
				WithField("blockNumber", n).
				Debug("Fetching Starknet block")
		}
		block, err := l.client.GetBlockByNumber(ctx, n, starknet.ScopeFullTXNAndReceipts)
		if err != nil {
			l.log.WithError(err).Error("Unable to fetch Starknet block")
			continue
		}
		for _, tx := range block.Transactions {
			for _, evt := range tx.Events {
				include := false
				for _, addr := range l.addresses {
					if bytes.Equal(evt.FromAddress.Bytes(), addr.Bytes()) {
						include = true
						break
					}
				}
				if include {
					all = append(all, &event{
						txnHash:     tx.TxnHash,
						fromAddress: evt.FromAddress,
						time:        time.Unix(block.AcceptedTime, 0),
						keys:        evt.Keys,
						data:        evt.Data,
					})
				}
			}
		}
	}

	l.lastBlockNumber = to

	return all, nil
}

func (l *eventListener) listenerRoutine(ctx context.Context) {
	t := time.NewTicker(l.interval)
	defer t.Stop()
	for {
		select {
		case <-ctx.Done():
			l.mu.Lock()
			close(l.outCh)
			l.mu.Unlock()
			return
		case <-t.C:
			func() {
				l.mu.Lock()
				defer l.mu.Unlock()
				txns, err := l.nextTransactions(ctx)
				if err != nil {
					l.log.WithError(err).Error("Unable to fetch events")
					return
				}
				for _, tx := range txns {
					l.outCh <- tx
				}
			}()
		}
	}
}
