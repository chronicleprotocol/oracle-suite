package starknet

import (
	"context"
	"encoding/json"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"

	"github.com/chronicleprotocol/oracle-suite/internal/starknet"
	"github.com/chronicleprotocol/oracle-suite/internal/starknet/mocks"
	"github.com/chronicleprotocol/oracle-suite/pkg/log/null"
)

func Test_acceptedBlockListener(t *testing.T) {
	ctx, cancelFunc := context.WithCancel(context.Background())
	defer cancelFunc()
	seq := &mocks.Sequencer{}
	ch := make(chan *event, 10)

	lis := acceptedBlockListener{
		sequencer:    seq,
		addresses:    []*starknet.Felt{starknet.HexToFelt("0x197f9e93cfaf7068ca2daf3ec89c2b91d051505c2231a0a0b9f70801a91fb24")},
		interval:     10 * time.Second,
		maxBlocks:    3,
		blocksBehind: []uint64{10},
		eventsCh:     ch,
		log:          null.New(),
	}

	block := &starknet.Block{}
	err := json.Unmarshal([]byte(testBlockResponse), block)
	if err != nil {
		panic(err)
	}

	// During the first interval, fetch the three black blocks defined in
	// maxBlocks, the blocks must be 10 blocks from the last block.
	seq.On("GetLatestBlock", ctx).Return(block, nil).Once()
	seq.On("GetBlockByNumber", ctx, uint64(191492)).Return(block, nil).Once()
	seq.On("GetBlockByNumber", ctx, uint64(191493)).Return(block, nil).Once()
	seq.On("GetBlockByNumber", ctx, uint64(191494)).Return(block, nil).Once()
	lis.fetchEvents(ctx)
	time.Sleep(time.Millisecond * 100)
	assert.Len(t, ch, 3)

	// During the second interval, there is no new blocks.
	seq.On("GetLatestBlock", ctx).Return(block, nil).Once()
	lis.fetchEvents(ctx)
	time.Sleep(time.Millisecond * 100)
	assert.Len(t, ch, 3)
}

func Test_pendingBlockListener(t *testing.T) {
	ctx, cancelFunc := context.WithCancel(context.Background())
	defer cancelFunc()
	seq := &mocks.Sequencer{}
	ch := make(chan *event, 10)

	lis := pendingBlockListener{
		sequencer: seq,
		addresses: []*starknet.Felt{starknet.HexToFelt("0x197f9e93cfaf7068ca2daf3ec89c2b91d051505c2231a0a0b9f70801a91fb24")},
		interval:  10 * time.Second,
		eventsCh:  ch,
		log:       null.New(),
	}

	block := &starknet.Block{}
	err := json.Unmarshal([]byte(testBlockResponse), block)
	if err != nil {
		panic(err)
	}

	// First interval.
	seq.On("GetPendingBlock", ctx).Return(block, nil).Once()
	lis.fetchEvents(ctx)
	time.Sleep(time.Millisecond * 100)
	assert.Len(t, ch, 1)

	// Second interval.
	seq.On("GetPendingBlock", ctx).Return(block, nil).Once()
	lis.fetchEvents(ctx)
	time.Sleep(time.Millisecond * 100)
	assert.Len(t, ch, 2)
}
