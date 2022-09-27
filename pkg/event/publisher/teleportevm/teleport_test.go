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
	"math/big"
	"testing"
	"time"

	geth "github.com/ethereum/go-ethereum"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"

	pkgEthereum "github.com/chronicleprotocol/oracle-suite/pkg/ethereum/geth"
	"github.com/chronicleprotocol/oracle-suite/pkg/ethereum/geth/mocks"
	"github.com/chronicleprotocol/oracle-suite/pkg/log/null"
)

var teleportTestAddress = common.HexToAddress("0x2d800d93b065ce011af83f316cef9f0d005b0aa4")
var teleportTestGUID = common.FromHex("0x111111111111111111111111111111111111111111111111111111111111111122222222222222222222222222222222222222222222222222222222222222220000000000000000000000003333333333333333333333333333333333333333000000000000000000000000444444444444444444444444444444444444444400000000000000000000000000000000000000000000000000000000000000370000000000000000000000000000000000000000000000000000000000000042000000000000000000000000000000000000000000000000000000000000004d")

func Test_teleportEventProvider_FetchEventsRoutine(t *testing.T) {
	ctx, cancelFunc := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancelFunc()

	cli := &mocks.EthClient{}
	ep, err := New(Config{
		Client:             pkgEthereum.NewClient(cli, nil),
		Addresses:          []common.Address{teleportTestAddress},
		Interval:           100 * time.Millisecond,
		PrefetchPeriod:     100 * time.Second,
		BlockLimit:         10,
		BlockConfirmations: 1,
		Logger:             null.New(),
	})
	require.NoError(t, err)
	ep.disablePrefetchEventsRoutine = true
	ep.disableFetchEventsRoutine = false

	txHash := common.HexToHash("0x66e8ab5a41d4b109c7f6ea5303e3c292771e57fb0b93a8474ca6f72e53eac0e8")
	logs := []types.Log{
		{TxIndex: 1, Data: teleportTestGUID, TxHash: txHash, Address: teleportTestAddress},
		{TxIndex: 2, Data: teleportTestGUID, TxHash: txHash, Address: teleportTestAddress},
	}

	cli.On("BlockNumber", ctx).Return(uint64(100), nil).Once()
	cli.On("BlockNumber", ctx).Return(uint64(119), nil).Once()
	cli.On("BlockNumber", ctx).Return(uint64(125), nil).Once()

	// First two ranges must be split into two FilterLogs calls to avoid exceeding the block limit.
	cli.On("FilterLogs", ctx, mock.Anything).Return(logs, nil).Once().Run(func(args mock.Arguments) {
		fq := args.Get(1).(geth.FilterQuery)
		assert.Equal(t, uint64(100), fq.FromBlock.Uint64()) // latest block minus block confirmations
		assert.Equal(t, uint64(109), fq.ToBlock.Uint64())   // latest block minus block confirmations minus block limit
		assert.Equal(t, []common.Address{teleportTestAddress}, fq.Addresses)
		assert.Equal(t, [][]common.Hash{{teleportTopic0}}, fq.Topics)
	})
	cli.On("FilterLogs", ctx, mock.Anything).Return(logs, nil).Once().Run(func(args mock.Arguments) {
		fq := args.Get(1).(geth.FilterQuery)
		assert.Equal(t, uint64(110), fq.FromBlock.Uint64())
		assert.Equal(t, uint64(118), fq.ToBlock.Uint64())
		assert.Equal(t, []common.Address{teleportTestAddress}, fq.Addresses)
		assert.Equal(t, [][]common.Hash{{teleportTopic0}}, fq.Topics)
	})
	cli.On("FilterLogs", ctx, mock.Anything).Return(logs, nil).Once().Run(func(args mock.Arguments) {
		fq := args.Get(1).(geth.FilterQuery)
		assert.Equal(t, uint64(119), fq.FromBlock.Uint64())
		assert.Equal(t, uint64(124), fq.ToBlock.Uint64())
		assert.Equal(t, []common.Address{teleportTestAddress}, fq.Addresses)
		assert.Equal(t, [][]common.Hash{{teleportTopic0}}, fq.Topics)
	})

	require.NoError(t, ep.Start(ctx))

	waitForEvents(ctx, t, ep, 6)
}

func Test_teleportEventProvider_PrefetchEventsRoutine(t *testing.T) {
	ctx, cancelFunc := context.WithTimeout(context.Background(), time.Second)
	defer cancelFunc()

	cli := &mocks.EthClient{}
	ep, err := New(Config{
		Client:             pkgEthereum.NewClient(cli, nil),
		Addresses:          []common.Address{teleportTestAddress},
		Interval:           100 * time.Millisecond,
		PrefetchPeriod:     100 * time.Second,
		BlockLimit:         15,
		BlockConfirmations: 1,
		Logger:             null.New(),
	})
	ep.disablePrefetchEventsRoutine = false
	ep.disableFetchEventsRoutine = true
	require.NoError(t, err)

	txHash := common.HexToHash("0x66e8ab5a41d4b109c7f6ea5303e3c292771e57fb0b93a8474ca6f72e53eac0e8")
	logs := []types.Log{
		{TxIndex: 1, Data: teleportTestGUID, TxHash: txHash, Address: teleportTestAddress},
		{TxIndex: 2, Data: teleportTestGUID, TxHash: txHash, Address: teleportTestAddress},
	}

	now := time.Now().Unix()
	cli.On("BlockByNumber", mock.Anything, big.NewInt(99)).Return(dummyBlock(99, now), nil)
	cli.On("BlockByNumber", mock.Anything, big.NewInt(84)).Return(dummyBlock(84, now-80), nil)
	cli.On("BlockByNumber", mock.Anything, big.NewInt(69)).Return(dummyBlock(69, now-160), nil)
	cli.On("BlockNumber", ctx).Return(uint64(100), nil).Once()
	cli.On("FilterLogs", ctx, mock.Anything).Return([]types.Log{}, nil).Once().Run(func(args mock.Arguments) {
		fq := args.Get(1).(geth.FilterQuery)
		assert.Equal(t, uint64(85), fq.FromBlock.Uint64()) // latest block minus block confirmations minus block limit
		assert.Equal(t, uint64(99), fq.ToBlock.Uint64())   // latest block minus block confirmations
		assert.Equal(t, []common.Address{teleportTestAddress}, fq.Addresses)
		assert.Equal(t, [][]common.Hash{{teleportTopic0}}, fq.Topics)
	})
	cli.On("FilterLogs", ctx, mock.Anything).Return([]types.Log{}, nil).Once().Run(func(args mock.Arguments) {
		fq := args.Get(1).(geth.FilterQuery)
		assert.Equal(t, uint64(70), fq.FromBlock.Uint64())
		assert.Equal(t, uint64(84), fq.ToBlock.Uint64())
		assert.Equal(t, []common.Address{teleportTestAddress}, fq.Addresses)
		assert.Equal(t, [][]common.Hash{{teleportTopic0}}, fq.Topics)
	})
	cli.On("FilterLogs", ctx, mock.Anything).Return(logs, nil).Once().Run(func(args mock.Arguments) {
		fq := args.Get(1).(geth.FilterQuery)
		assert.Equal(t, uint64(55), fq.FromBlock.Uint64())
		assert.Equal(t, uint64(69), fq.ToBlock.Uint64())
		assert.Equal(t, []common.Address{teleportTestAddress}, fq.Addresses)
		assert.Equal(t, [][]common.Hash{{teleportTopic0}}, fq.Topics)
	})

	require.NoError(t, ep.Start(ctx))

	waitForEvents(ctx, t, ep, 2)
}

func waitForEvents(ctx context.Context, t *testing.T, ep *EventProvider, expectedEvents int) {
	events := 0
loop:
	for events < expectedEvents {
		select {
		case msg := <-ep.Events():
			events++
			assert.Equal(t, common.FromHex("0x69515a78ae1ad8c4650b57eb6dcd0c866b71e828316dabbc64f430588d043452"), msg.Data["hash"])
			assert.Equal(t, teleportTestGUID, msg.Data["event"])
		case <-ctx.Done():
			break loop
		}
	}

	assert.Equal(t, expectedEvents, events)
}

func dummyBlock(number uint64, timestamp int64) *types.Block {
	return types.NewBlock(
		&types.Header{
			Number: big.NewInt(int64(number)),
			Time:   uint64(timestamp),
		},
		nil,
		nil,
		nil,
		nil,
	)
}
