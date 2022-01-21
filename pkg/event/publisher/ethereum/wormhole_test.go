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

var wormholeTestAddress = common.HexToAddress("0x2d800d93b065ce011af83f316cef9f0d005b0aa4")
var wormholeTestGUID = common.FromHex("0x111111111111111111111111111111111111111111111111111111111111111122222222222222222222222222222222222222222222222222222222222222220000000000000000000000003333333333333333333333333333333333333333000000000000000000000000444444444444444444444444444444444444444400000000000000000000000000000000000000000000000000000000000000370000000000000000000000000000000000000000000000000000000000000042000000000000000000000000000000000000000000000000000000000000004d")

func Test_wormholeListener(t *testing.T) {
	ctx, cancelFunc := context.WithCancel(context.Background())
	cli := &mocks.EthClient{}

	w := NewWormholeListener(WormholeListenerConfig{
		Client:       cli,
		Addresses:    []common.Address{wormholeTestAddress},
		Interval:     time.Millisecond * 100,
		BlocksBehind: 10,
		MaxBlocks:    15,
		Logger:       null.New(),
	})

	txHash := common.HexToHash("0x66e8ab5a41d4b109c7f6ea5303e3c292771e57fb0b93a8474ca6f72e53eac0e8")
	logs := []types.Log{
		{TxIndex: 1, Data: wormholeTestGUID, TxHash: txHash},
		{TxIndex: 2, Data: wormholeTestGUID, TxHash: txHash},
	}

	cli.On("BlockNumber", ctx).Return(uint64(42), nil).Once()
	cli.On("FilterLogs", ctx, mock.Anything).Return(logs, nil).Once().Run(func(args mock.Arguments) {
		fq := args.Get(1).(geth.FilterQuery)
		assert.Equal(t, uint64(17), fq.FromBlock.Uint64())
		assert.Equal(t, uint64(32), fq.ToBlock.Uint64())
		assert.Equal(t, []common.Address{wormholeTestAddress}, fq.Addresses)
		assert.Equal(t, [][]common.Hash{{wormholeTopic0}}, fq.Topics)
	})

	require.NoError(t, w.Start(ctx))
	for {
		if len(cli.Calls) >= 2 { // 2 is the number of mocked calls above.
			cancelFunc()
			break
		}
		time.Sleep(time.Millisecond * 10)
	}
	assert.Len(t, w.Events(), 2) // 2 events are expected
	for len(w.Events()) > 0 {
		msg := <-w.Events()
		assert.Equal(t, txHash.Bytes(), msg.Index)
		assert.Equal(t, common.FromHex("0x69515a78ae1ad8c4650b57eb6dcd0c866b71e828316dabbc64f430588d043452"), msg.Data["hash"])
		assert.Equal(t, wormholeTestGUID, msg.Data["event"])
	}
}

func Test_packWormholeGUID(t *testing.T) {
	g, err := unpackWormholeGUID(wormholeTestGUID)
	require.NoError(t, err)

	b, err := packWormholeGUID(g)
	require.NoError(t, err)
	assert.Equal(t, wormholeTestGUID, b)
}

func Test_unpackWormholeGUID(t *testing.T) {
	g, err := unpackWormholeGUID(wormholeTestGUID)

	require.NoError(t, err)
	assert.Equal(t, common.HexToHash("0x1111111111111111111111111111111111111111111111111111111111111111"), g.sourceDomain)
	assert.Equal(t, common.HexToHash("0x2222222222222222222222222222222222222222222222222222222222222222"), g.targetDomain)
	assert.Equal(t, common.HexToHash("0x0000000000000000000000003333333333333333333333333333333333333333"), g.receiver)
	assert.Equal(t, common.HexToHash("0x0000000000000000000000004444444444444444444444444444444444444444"), g.operator)
	assert.Equal(t, big.NewInt(55), g.amount)
	assert.Equal(t, big.NewInt(66), g.nonce)
	assert.Equal(t, int64(77), g.timestamp)
}
