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

	"github.com/chronicleprotocol/oracle-suite/pkg/log"
)

const retryAttempts = 3               // The maximum number of attempts to call EthClient in case of an error.
const retryInterval = 5 * time.Second // The delay between retry attempts.

type EthClient interface {
	BlockNumber(ctx context.Context) (uint64, error)
	FilterLogs(ctx context.Context, q geth.FilterQuery) ([]types.Log, error)
}

// logListener listens for Ethereum logs and returns them using a channel.
type logListener interface {
	Start(ctx context.Context)
	Logs() chan types.Log
}

// ethClientLogListener implements logListener interface.
type ethClientLogListener struct {
	mu sync.Mutex

	client          EthClient
	addresses       []common.Address
	topics          [][]common.Hash
	interval        time.Duration // Time interval between pulling logs from Ethereum client.
	lastBlockNumber uint64        // Last block from which logs were pulled.
	blocksBehind    uint64        // Number of blocks behind the latest one.
	maxBlocks       uint64        // Maximum number of blocks from which logs can be fetched.
	outCh           chan types.Log
	log             log.Logger
}

// newEthClientLogListener returns a logListener.
func newEthClientLogListener(
	client EthClient,
	addresses []common.Address,
	topics []common.Hash,
	interval time.Duration,
	blocksBehind uint64,
	maxBlocks uint64,
	logger log.Logger,
) logListener {

	l := &ethClientLogListener{
		client:       client,
		addresses:    addresses,
		interval:     interval,
		blocksBehind: blocksBehind,
		maxBlocks:    maxBlocks,
		outCh:        make(chan types.Log, 1),
		log:          logger,
	}
	for _, t := range topics {
		l.topics = append(l.topics, []common.Hash{t})
	}
	return l
}

// Start implements the logListener interface.
func (l *ethClientLogListener) Start(ctx context.Context) {
	go l.listenerRoutine(ctx)
}

// Logs implements the logListener interface.
func (l *ethClientLogListener) Logs() chan types.Log {
	return l.outCh
}

func (l *ethClientLogListener) listenerRoutine(ctx context.Context) {
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
				logs, err := l.nextLogs(ctx)
				if err != nil {
					l.log.WithError(err).Error("Unable to fetch logs")
					return
				}
				for _, log := range logs {
					l.outCh <- log
				}
			}()
		}
	}
}

// nextLogs returns a logs from a range returned by the nextBlockNumberRange
// method and updates lastBlockNumber variable.
func (l *ethClientLogListener) nextLogs(ctx context.Context) ([]types.Log, error) {
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

	// Fetch logs.
	var all []types.Log
	for _, addr := range l.addresses {
		var logs []types.Log
		if err = retry(func() error {
			if l.log.Level() >= log.Debug {
				l.log.
					WithField("from", from).
					WithField("to", to).
					WithField("address", addr.String()).
					WithField("topics", l.topics).
					Debug("Fetching Ethereum logs")
			}
			logs, err = l.client.FilterLogs(ctx, geth.FilterQuery{
				FromBlock: new(big.Int).SetUint64(from),
				ToBlock:   new(big.Int).SetUint64(to),
				Addresses: []common.Address{addr},
				Topics:    l.topics,
			})
			return err
		}); err != nil {
			return nil, err
		}
		all = append(all, logs...)
	}

	// The last block number should be updated at the last possible moment so
	// in case of an error, this method will try to fetch once again logs from
	// the same range.
	l.lastBlockNumber = to

	return all, nil
}

// nextBlockNumberRange returns the next block range from which logs should
// be fetched.
func (l *ethClientLogListener) nextBlockNumberRange(ctx context.Context) (uint64, uint64, error) {
	var err error
	var curr uint64
	err = retry(func() error {
		curr, err = l.client.BlockNumber(ctx)
		return err
	})
	if err != nil {
		return 0, 0, err
	}
	from := l.lastBlockNumber + 1
	to := curr - l.blocksBehind
	if from > to {
		from = to
	}
	if to-from > l.maxBlocks {
		from = to - l.maxBlocks
	}
	return from, to, nil
}

// retry runs the f function until it returns nil. Maximum number of retries
// and delay between them are defined in the retryAttempts and retryInterval
// constants.
func retry(f func() error) (err error) {
	for i := 0; i < retryAttempts; i++ {
		if i > 0 {
			time.Sleep(retryInterval)
		}
		err = f()
		if err == nil {
			return nil
		}
	}
	return err
}
