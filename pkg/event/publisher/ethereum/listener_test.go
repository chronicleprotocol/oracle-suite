package ethereum

import (
	"context"
	"testing"
	"time"

	geth "github.com/ethereum/go-ethereum"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"

	"github.com/chronicleprotocol/oracle-suite/pkg/ethereum"
	"github.com/chronicleprotocol/oracle-suite/pkg/ethereum/geth/mocks"
	"github.com/chronicleprotocol/oracle-suite/pkg/log/null"
)

var listenerTestAddress = common.HexToAddress("0x2d800d93b065ce011af83f316cef9f0d005b0aa4")
var listenerTestTopic1 = ethereum.HexToHash("0x0162851814eed5360ac17d1b3d942e6619fa2d803de71e7159ed9bebf724072a")
var listenerTestTopic2 = ethereum.HexToHash("0x618b439c5646a69d0ee5bf11275cce6932a3eaca08b9ad1c68b4531001560de7")

func Test_logListener(t *testing.T) {
	ctx, cancelFunc := context.WithCancel(context.Background())
	cli := &mocks.EthClient{}
	lis := newEthClientLogListener(cli, []common.Address{listenerTestAddress}, []common.Hash{listenerTestTopic1, listenerTestTopic2}, time.Millisecond*100, 10, 15, null.New())

	// During the first call we are expecting to fetch up to maxBlocks.
	cli.On("BlockNumber", ctx).Return(uint64(42), nil).Once()
	cli.On("FilterLogs", ctx, mock.Anything).Return([]types.Log{{Index: 1}, {Index: 2}}, nil).Once().Run(func(args mock.Arguments) {
		fq := args.Get(1).(geth.FilterQuery)
		assert.Equal(t, uint64(17), fq.FromBlock.Uint64()) // currentBlockNumber-blocksBehind-maxBlocks
		assert.Equal(t, uint64(32), fq.ToBlock.Uint64())   // currentBlockNumber-blocksBehind
		assert.Equal(t, []common.Address{listenerTestAddress}, fq.Addresses)
		assert.Equal(t, [][]common.Hash{{listenerTestTopic1}, {listenerTestTopic2}}, fq.Topics)
	})
	// During the second call, we expect to fetch blocks between the last
	// fetched one and the current one minus the value of blocksBehind..
	cli.On("BlockNumber", ctx).Return(uint64(52), nil).Once()
	cli.On("FilterLogs", ctx, mock.Anything).Return([]types.Log{{Index: 3}, {Index: 4}}, nil).Once().Run(func(args mock.Arguments) {
		fq := args.Get(1).(geth.FilterQuery)
		assert.Equal(t, uint64(33), fq.FromBlock.Uint64()) // lastBlockNumber+1
		assert.Equal(t, uint64(42), fq.ToBlock.Uint64())   // currentBlockNumber-blocksBehind
		assert.Equal(t, []common.Address{listenerTestAddress}, fq.Addresses)
		assert.Equal(t, [][]common.Hash{{listenerTestTopic1}, {listenerTestTopic2}}, fq.Topics)
	})
	// If the latest block number is the same as the one during previous call,
	// the fetchLogs method should do nothing.
	cli.On("BlockNumber", ctx).Return(uint64(52), nil).Once()

	// Start listener and collect logs:
	lis.Start(ctx)
	var logs []types.Log
	for {
		if len(cli.Calls) >= 5 { // 5 is the number of mocked calls above.
			cancelFunc()
			break
		}
		time.Sleep(time.Millisecond * 10)
	}
	for len(lis.Logs()) > 0 {
		logs = append(logs, <-lis.Logs())
	}
	assert.Len(t, logs, 4)
}
