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

package teleportevm

import (
	"context"
	"errors"
	"math/big"
	"time"

	geth "github.com/ethereum/go-ethereum"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/types"

	"github.com/chronicleprotocol/oracle-suite/pkg/ethereum"
	"github.com/chronicleprotocol/oracle-suite/pkg/log"
	"github.com/chronicleprotocol/oracle-suite/pkg/transport/messages"
	"github.com/chronicleprotocol/oracle-suite/pkg/util/retry"
)

const TeleportEventType = "teleport_evm"
const LoggerTag = "ETHEREUM_TELEPORT"
const retryAttempts = 3               // The maximum number of attempts to call Client in case of an error.
const retryInterval = 5 * time.Second // The delay between retry attempts.

// teleportTopic0 is Keccak256("TeleportInitialized((bytes32,bytes32,bytes32,bytes32,uint128,uint80,uint48))")
var teleportTopic0 = ethereum.HexToHash("0x61aedca97129bac4264ec6356bd1f66431e65ab80e2d07b7983647d72776f545")

// Client is a Ethereum compatible client.
type Client interface {
	BlockNumber(ctx context.Context) (uint64, error)
	FilterLogs(ctx context.Context, q geth.FilterQuery) ([]types.Log, error)
}

// TeleportEventProviderConfig contains a configuration options for New.
type TeleportEventProviderConfig struct {
	// Client is an instance of Ethereum RPC client.
	Client Client
	// Addresses is a list of contracts from which logs will be fetched.
	Addresses []ethereum.Address
	// Interval specifies how often provider should check for new logs.
	Interval time.Duration
	// BlocksDelta is a list of distances between the latest block on the
	// blockchain and blocks from which logs are to be taken. The purpose of
	// this field is to ensure that older events are resent from time to time.
	BlocksDelta []int
	// BlocksLimit specifies how from many blocks logs can be fetched at once.
	BlocksLimit int
	// Logger is a current logger interface used by the TeleportEventProvider.
	// The Logger is used to monitor asynchronous processes.
	Logger log.Logger
}

// TeleportEventProvider listens to TeleportGUID events on Ethereum compatible
// blockchains.
//
// https://github.com/makerdao/dss-teleport
type TeleportEventProvider struct {
	eventCh chan *messages.Event

	// lastBlock is a number of last block from which events were fetched.
	// it is used in the nextBlockRange function.
	lastBlock uint64

	// Configuration parameters copied from TeleportEventProviderConfig:
	client      Client
	interval    time.Duration
	addresses   []common.Address
	blocksDelta []uint64
	blocksLimit uint64
	log         log.Logger
}

// New returns a new instance of the TeleportEventProvider struct.
func New(cfg TeleportEventProviderConfig) *TeleportEventProvider {
	return &TeleportEventProvider{
		eventCh:     make(chan *messages.Event),
		client:      cfg.Client,
		interval:    cfg.Interval,
		addresses:   cfg.Addresses,
		blocksDelta: intsToUint64s(cfg.BlocksDelta),
		blocksLimit: uint64(cfg.BlocksLimit),
		log:         cfg.Logger.WithField("tag", LoggerTag),
	}
}

// Events implements the publisher.Listener interface.
func (tp *TeleportEventProvider) Events() chan *messages.Event {
	return tp.eventCh
}

// Start implements the publisher.Listener interface.
func (tp *TeleportEventProvider) Start(ctx context.Context) error {
	go tp.fetchLogsRoutine(ctx)
	return nil
}

// fetchLogsRoutine periodically fetches logs from the blockchain.
func (tp *TeleportEventProvider) fetchLogsRoutine(ctx context.Context) {
	t := time.NewTicker(tp.interval)
	defer t.Stop()
	for {
		select {
		case <-ctx.Done():
			close(tp.eventCh)
			return
		case <-t.C:
			tp.fetchLogs(ctx)
		}
	}
}

// fetchLogs fetches TeleportGUID events from the blockchain and converts them
// into event messages. The converted messages are sent to the eventCh channel.
func (tp *TeleportEventProvider) fetchLogs(ctx context.Context) {
	rangeFrom, rangeTo, err := tp.nextBlockRange(ctx)
	if err != nil {
		tp.log.
			WithError(err).
			Error("Unable to get latest block number")
		return
	}
	if rangeFrom == tp.lastBlock {
		return // There is no new blocks to fetch.
	}
	for _, delta := range tp.blocksDelta {
		for _, address := range tp.addresses {
			if ctx.Err() != nil {
				return
			}
			if delta > rangeFrom {
				delta = rangeFrom // To prevent overflow.
			}
			from := rangeFrom - delta
			to := rangeTo - delta
			tp.log.
				WithFields(log.Fields{
					"from":    from,
					"to":      to,
					"address": address.String(),
				}).
				Info("Fetching logs")
			logs, err := tp.filterLogs(ctx, address, from, to, teleportTopic0)
			if errors.Is(err, context.Canceled) {
				continue
			}
			if err != nil {
				tp.log.
					WithError(err).
					Error("Unable to fetch logs")
				continue
			}
			for _, l := range logs {
				if l.Removed {
					continue
				}
				if l.Address != address {
					// This should never happen. All logs returned by
					// eth_filterLogs should be emitted by the specified
					// contract. If it happens, there is a bug somewhere.
					tp.log.
						WithFields(log.Fields{
							"expected": address.String(),
							"actual":   l.Address.String(),
						}).
						Panic("Log emitted by wrong contract")
				}
				msg, err := logToMessage(l)
				if err != nil {
					tp.log.
						WithError(err).
						Error("Unable to convert log to event")
					continue
				}
				tp.eventCh <- msg
			}
		}
	}
	tp.lastBlock = rangeTo
}

// nextBlockRange returns the range of blocks from which logs should be
// fetched.
func (tp *TeleportEventProvider) nextBlockRange(ctx context.Context) (uint64, uint64, error) {
	// Get the latest block number.
	to, err := tp.getBlockNumber(ctx)
	if err != nil {
		return 0, 0, err
	}
	// Set "from" to the next block and check if "from" is greater than "to",
	// if so, then there are no new blocks to fetch.
	from := tp.lastBlock + 1
	if from > to {
		return to, to, nil
	}
	// Limit the number of blocks to fetch.
	if to-from > tp.blocksLimit {
		from = to - tp.blocksLimit + 1
	}
	return from, to, nil
}

// getBlockNumber returns the latest block number on the blockchain.
func (tp *TeleportEventProvider) getBlockNumber(ctx context.Context) (uint64, error) {
	var err error
	var res uint64
	err = retry.Retry(
		ctx,
		func() error {
			res, err = tp.client.BlockNumber(ctx)
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
func (tp *TeleportEventProvider) filterLogs(
	ctx context.Context,
	addr common.Address,
	from, to uint64,
	topic0 common.Hash,
) ([]types.Log, error) {

	var err error
	var res []types.Log
	err = retry.Retry(
		ctx,
		func() error {
			res, err = tp.client.FilterLogs(ctx, geth.FilterQuery{
				FromBlock: new(big.Int).SetUint64(from),
				ToBlock:   new(big.Int).SetUint64(to),
				Addresses: []common.Address{addr},
				Topics:    [][]common.Hash{{topic0}},
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
