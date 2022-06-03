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
	ch := make(chan *event)

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

	seq.On("GetLatestBlock", ctx).Return(block, nil).Once()
	seq.On("GetBlockByNumber", ctx, uint64(191492)).Return(block, nil).Once()
	seq.On("GetBlockByNumber", ctx, uint64(191493)).Return(block, nil).Once()
	seq.On("GetBlockByNumber", ctx, uint64(191494)).Return(block, nil).Once()

	// Start listener and collect logs:
	lis.start(ctx)
	var evts []*event
	for len(evts) < 3 {
		evts = append(evts, <-lis.events())
		time.Sleep(time.Millisecond * 10)
	}
	assert.Len(t, evts, 4)
}

func Test_pendingBlockListener(t *testing.T) {
	ctx, cancelFunc := context.WithCancel(context.Background())
	defer cancelFunc()
	seq := &mocks.Sequencer{}
	ch := make(chan *event)

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

	seq.On("GetPendingBlock", ctx).Return(block, nil).Once()

	// Start listener and collect logs:
	lis.start(ctx)
	var evts []*event
	for len(evts) < 1 {
		evts = append(evts, <-lis.events())
		time.Sleep(time.Millisecond * 10)
	}
	assert.Len(t, evts, 1)
}
