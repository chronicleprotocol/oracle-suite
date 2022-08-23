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
	"testing"
	"time"

	geth "github.com/ethereum/go-ethereum"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"

	"github.com/chronicleprotocol/oracle-suite/pkg/ethereum/geth/mocks"
	"github.com/chronicleprotocol/oracle-suite/pkg/log/null"
)

var teleportTestAddress = common.HexToAddress("0x2d800d93b065ce011af83f316cef9f0d005b0aa4")
var teleportTestGUID = common.FromHex("0x111111111111111111111111111111111111111111111111111111111111111122222222222222222222222222222222222222222222222222222222222222220000000000000000000000003333333333333333333333333333333333333333000000000000000000000000444444444444444444444444444444444444444400000000000000000000000000000000000000000000000000000000000000370000000000000000000000000000000000000000000000000000000000000042000000000000000000000000000000000000000000000000000000000000004d")

func Test_teleportListener(t *testing.T) {
	ctx, cancelFunc := context.WithTimeout(context.Background(), time.Second)
	defer cancelFunc()

	cli := &mocks.EthClient{}
	w := New(TeleportEventProviderConfig{
		Client:      cli,
		Addresses:   []common.Address{teleportTestAddress},
		Interval:    time.Millisecond * 100,
		BlocksDelta: []int{0, 10},
		BlocksLimit: 15,
		Logger:      null.New(),
	})

	// Test logs:
	txHash := common.HexToHash("0x66e8ab5a41d4b109c7f6ea5303e3c292771e57fb0b93a8474ca6f72e53eac0e8")
	logs := []types.Log{
		{TxIndex: 1, Data: teleportTestGUID, TxHash: txHash, Address: teleportTestAddress},
		{TxIndex: 2, Data: teleportTestGUID, TxHash: txHash, Address: teleportTestAddress},
	}

	// During the first call we are expecting to fetch up to blocksLimit.
	cli.On("BlockNumber", ctx).Return(uint64(42), nil).Once()
	cli.On("FilterLogs", ctx, mock.Anything).Return([]types.Log{}, nil).Once().Run(func(args mock.Arguments) {
		fq := args.Get(1).(geth.FilterQuery)
		assert.Equal(t, uint64(28), fq.FromBlock.Uint64())
		assert.Equal(t, uint64(42), fq.ToBlock.Uint64())
		assert.Equal(t, []common.Address{teleportTestAddress}, fq.Addresses)
		assert.Equal(t, [][]common.Hash{{teleportTopic0}}, fq.Topics)
	})
	// During the second call, we expect to fetch blocks between the last
	// fetched one and the current one minus the value of blocksDelta..
	cli.On("BlockNumber", ctx).Return(uint64(52), nil).Once()
	cli.On("FilterLogs", ctx, mock.Anything).Return(logs, nil).Once().Run(func(args mock.Arguments) {
		fq := args.Get(1).(geth.FilterQuery)
		assert.Equal(t, uint64(18), fq.FromBlock.Uint64())
		assert.Equal(t, uint64(32), fq.ToBlock.Uint64())
		assert.Equal(t, []common.Address{teleportTestAddress}, fq.Addresses)
		assert.Equal(t, [][]common.Hash{{teleportTopic0}}, fq.Topics)
	})

	require.NoError(t, w.Start(ctx))

	events := 0
	for {
		msg := <-w.Events()
		events++
		assert.Equal(t, txHash.Bytes(), msg.Index)
		assert.Equal(t, common.FromHex("0x69515a78ae1ad8c4650b57eb6dcd0c866b71e828316dabbc64f430588d043452"), msg.Data["hash"])
		assert.Equal(t, teleportTestGUID, msg.Data["event"])
		if events == 2 {
			break
		}
	}
	assert.Equal(t, 2, events)
}
