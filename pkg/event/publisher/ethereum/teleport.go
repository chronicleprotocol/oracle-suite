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
	"time"

	geth "github.com/ethereum/go-ethereum"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/types"

	"github.com/chronicleprotocol/oracle-suite/internal/util/retry"
	"github.com/chronicleprotocol/oracle-suite/pkg/ethereum"
	"github.com/chronicleprotocol/oracle-suite/pkg/log"
	"github.com/chronicleprotocol/oracle-suite/pkg/transport/messages"
)

const TeleportEventType = "teleport_evm"
const LoggerTag = "TELEPORT_LISTENER"
const retryAttempts = 3               // The maximum number of attempts to call Client in case of an error.
const retryInterval = 5 * time.Second // The delay between retry attempts.

// teleportTopic0 is Keccak256("TeleportGUID((bytes32,bytes32,bytes32,bytes32,uint128,uint80,uint48))")
var teleportTopic0 = ethereum.HexToHash("0x9f692a9304834fdefeb4f9cd17d1493600af19c70af547480cccf4a8a4a7752c")

// Client is a Ethereum compatible client.
type Client interface {
	BlockNumber(ctx context.Context) (uint64, error)
	FilterLogs(ctx context.Context, q geth.FilterQuery) ([]types.Log, error)
}

// TeleportListenerConfig contains a configuration options for NewTeleportListener.
type TeleportListenerConfig struct {
	// Client is an instance of Ethereum RPC client.
	Client Client
	// Addresses is a list of contracts from which logs will be fetched.
	Addresses []ethereum.Address
	// Interval specifies how often listener should check for new logs.
	Interval time.Duration
	// BlocksDelta is a list of distances between the latest block on the
	// blockchain and blocks from which logs are to be taken.
	BlocksDelta []int
	// BlocksLimit specifies how from many blocks logs can be fetched at once.
	BlocksLimit int
	// Logger is a current logger interface used by the TeleportListener.
	// The Logger is used to monitor asynchronous processes.
	Logger log.Logger
}

// TeleportListener listens to TeleportGUID events on Ethereum compatible
// blockchains.
type TeleportListener struct {
	eventCh chan *messages.Event

	// lastBlock is a number of last block from which events were fetched.
	// it is used in the nextBlockRange function.
	lastBlock uint64

	// Configuration parameters copied from TeleportListenerConfig:
	client      Client
	interval    time.Duration
	addresses   []common.Address
	blocksDelta []uint64
	blocksLimit uint64
	logger      log.Logger
}

// NewTeleportListener returns a new instance of the TeleportListener struct.
func NewTeleportListener(cfg TeleportListenerConfig) *TeleportListener {
	return &TeleportListener{
		client:      cfg.Client,
		interval:    cfg.Interval,
		addresses:   cfg.Addresses,
		blocksDelta: intsToUint64s(cfg.BlocksDelta),
		blocksLimit: uint64(cfg.BlocksLimit),
		logger:      cfg.Logger.WithField("tag", LoggerTag),
		eventCh:     make(chan *messages.Event),
	}
}

// Events implements the publisher.Listener interface.
func (l *TeleportListener) Events() chan *messages.Event {
	return l.eventCh
}

// Start implements the publisher.Listener interface.
func (l *TeleportListener) Start(ctx context.Context) error {
	go l.fetchLogsRoutine(ctx)
	return nil
}

// fetchLogsRoutine periodically fetches logs from the blockchain.
func (l *TeleportListener) fetchLogsRoutine(ctx context.Context) {
	t := time.NewTicker(l.interval)
	defer t.Stop()
	for {
		select {
		case <-ctx.Done():
			close(l.eventCh)
			return
		case <-t.C:
			l.fetchLogs(ctx)
		}
	}
}

// fetchLogs fetches WormholeGUID events from the blockchain and converts them
// into event messages. The converted messages are sent to the eventCh channel.
func (l *TeleportListener) fetchLogs(ctx context.Context) {
	from, to, err := l.nextBlockRange(ctx)
	if err != nil {
		l.logger.
			WithError(err).
			Error("Unable to get latest block number")
		return
	}

	// There is no new blocks to fetch.
	if from == l.lastBlock {
		return
	}

	for _, delta := range l.blocksDelta {
		for _, address := range l.addresses {
			fetchFrom := from - delta
			fetchTo := to - delta

			// Fetch logs.
			l.logger.
				WithFields(log.Fields{
					"from":    fetchFrom,
					"to":      fetchTo,
					"address": address.String(),
				}).
				Info("Fetching logs")
			logs, err := l.filterLogs(ctx, address, fetchFrom, fetchTo)
			if err != nil {
				l.logger.
					WithError(err).
					Error("Unable to fetch logs")
				continue
			}

			// Convert logs to events.
			for _, log := range logs {
				msg, err := logToMessage(log)
				if err != nil {
					l.logger.
						WithError(err).
						Error("Unable to convert log to event")
					continue
				}
				l.eventCh <- msg
			}
		}
	}

	l.lastBlock = to
}

// nextBlockRange returns the range of blocks from which logs should be
// fetched. It returns the range from the latest fetched block stored in the
// lastBlock parameter to the latest block on the blockchain. The maximum
// number of blocks is limited by the blocksLimit parameter.
func (l *TeleportListener) nextBlockRange(ctx context.Context) (uint64, uint64, error) {
	// Get the latest block number.
	to, err := l.getBlockNumber(ctx)
	if err != nil {
		return 0, 0, err
	}

	// Set "from" to the next block. If "from" is greater than "to", then there
	// are no new blocks to fetch.
	from := l.lastBlock + 1
	if from > to {
		return to, to, nil
	}

	// Limit the number of blocks to fetch.
	if to-from > l.blocksLimit {
		from = to - l.blocksLimit + 1
	}

	return from, to, nil
}

// getBlockNumber returns the latest block number on the blockchain.
func (l *TeleportListener) getBlockNumber(ctx context.Context) (uint64, error) {
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

// filterLogs fetches TeleportGUID events from the blockchain.
func (l *TeleportListener) filterLogs(ctx context.Context, addr common.Address, from, to uint64) ([]types.Log, error) {
	var err error
	var res []types.Log
	err = retry.Retry(
		ctx,
		func() error {
			res, err = l.client.FilterLogs(ctx, geth.FilterQuery{
				FromBlock: new(big.Int).SetUint64(from),
				ToBlock:   new(big.Int).SetUint64(to),
				Addresses: []common.Address{addr},
				Topics:    [][]common.Hash{{teleportTopic0}},
			})
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

// intsToUint64s converts int slice to uint64 slice.
func intsToUint64s(i []int) []uint64 {
	u := make([]uint64, len(i))
	for n, v := range i {
		u[n] = uint64(v)
	}
	return u
}
