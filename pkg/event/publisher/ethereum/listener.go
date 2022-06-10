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
	"math/big"
	"sync"
	"time"

	geth "github.com/ethereum/go-ethereum"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/types"

	"github.com/chronicleprotocol/oracle-suite/internal/util/retry"
	"github.com/chronicleprotocol/oracle-suite/pkg/log"
)

const retryAttempts = 3               // The maximum number of attempts to call Client in case of an error.
const retryInterval = 5 * time.Second // The delay between retry attempts.

type Client interface {
	BlockNumber(ctx context.Context) (uint64, error)
	FilterLogs(ctx context.Context, q geth.FilterQuery) ([]types.Log, error)
}

// logListener periodically fetches logs from Ethereum compatible blockchains.
type logListener struct {
	mu sync.Mutex

	// Listener parameters:
	client      Client
	addresses   []common.Address
	topics      [][]common.Hash
	interval    time.Duration
	blocksDelta []uint64
	blocksLimit uint64
	logCh       chan types.Log
	logger      log.Logger

	// State:
	lastBlock uint64
}

func (l *logListener) start(ctx context.Context) {
	go l.listenerRoutine(ctx)
}

func (l *logListener) logs() chan types.Log {
	return l.logCh
}

func (l *logListener) listenerRoutine(ctx context.Context) {
	t := time.NewTicker(l.interval)
	defer t.Stop()
	for {
		select {
		case <-ctx.Done():
			l.mu.Lock()
			close(l.logCh)
			l.mu.Unlock()
			return
		case <-t.C:
			l.fetchLogs(ctx)
		}
	}
}

func (l *logListener) fetchLogs(ctx context.Context) {
	from, to, err := l.nextRange(ctx)
	if err != nil {
		l.logger.WithError(err).Error("Unable to get latest block number")
		return
	}
	// There is no new blocks to fetch.
	if from == l.lastBlock {
		return
	}
	for _, delta := range l.blocksDelta {
		for _, addr := range l.addresses {
			from := from - delta
			to := to - delta
			l.logger.
				WithFields(log.Fields{
					"from":    from,
					"to":      to,
					"address": addr.String(),
					"topics":  l.topics,
				}).
				Info("Fetching logs")
			logs, err := l.filterLogs(
				ctx,
				geth.FilterQuery{
					FromBlock: new(big.Int).SetUint64(from),
					ToBlock:   new(big.Int).SetUint64(to),
					Addresses: []common.Address{addr},
					Topics:    l.topics,
				},
			)
			if err != nil {
				l.logger.WithError(err).Error("Unable to fetch logs")
				continue
			}
			for _, log := range logs {
				l.logCh <- log
			}
		}
	}
	l.lastBlock = to
	return
}

func (l *logListener) nextRange(ctx context.Context) (uint64, uint64, error) {
	to, err := l.getBlockNumber(ctx)
	if err != nil {
		return 0, 0, err
	}
	from := l.lastBlock + 1
	// No new blocks since the last check.
	if from > to {
		from = to
	}
	// Cap the number of blocks to fetch.
	if to-from > l.blocksLimit {
		from = to - l.blocksLimit + 1
	}
	return from, to, nil
}

func (l *logListener) getBlockNumber(ctx context.Context) (uint64, error) {
	var err error
	var res uint64
	err = retry.Retry(
		ctx,
		func() error {
			res, err = l.client.BlockNumber(ctx)
			return err
		},
		retryAttempts,
		retryInterval,
	)
	if err != nil {
		return 0, err
	}
	return res, nil
}

func (l *logListener) filterLogs(ctx context.Context, query geth.FilterQuery) ([]types.Log, error) {
	var err error
	var res []types.Log
	err = retry.Retry(
		ctx,
		func() error {
			res, err = l.client.FilterLogs(ctx, query)
			return err
		},
		retryAttempts,
		retryInterval,
	)
	if err != nil {
		return nil, err
	}
	return res, nil
}
