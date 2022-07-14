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

package starknet

import (
	"context"
	"encoding/json"
	"testing"
	"time"

	"github.com/ethereum/go-ethereum/common"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"

	"github.com/stretchr/testify/require"

	"github.com/chronicleprotocol/oracle-suite/pkg/log/null"
	"github.com/chronicleprotocol/oracle-suite/pkg/starknet"
	"github.com/chronicleprotocol/oracle-suite/pkg/starknet/mocks"
)

const testBlockResponse = `
{
  "block_hash": "0x74ff65a69e077e69663539f8a277d3c81965f7eb9a61d039b437e66290f38ea",
  "parent_block_hash": "0x26af2e23367fd4f46198bf469d5dbbe33b29919710b1fa08b65599f79672ecb",
  "block_number": 191504,
  "state_root": "00ee28831898c577fd55991e693865e3c280e3e5051b569bca0c25ccf212310e",
  "status": "ACCEPTED_ON_L1",
  "gas_price": "0x59682f07",
  "transactions": [
    {
      "contract_address": "0x4aef408ac955fd9e5f1397ae9801c1b03344587231f454a0b49d0b99e77c937",
      "entry_point_selector": "0x15d40a3d6ca2ac30f4031e42be28da9b056fef9bb7357ac5e85627ee876e5ad",
      "entry_point_type": "EXTERNAL",
      "calldata": [
        "0x1",
        "0x73314940630fd6dcda0d772d4c972c4e0a9946bef9dabf4ef84eda8ef542b82",
        "0xe48e45e0642d5f170bb832c637926f4c85b77d555848b693304600c4275f26",
        "0x0",
        "0x3",
        "0x3",
        "0x9b872ffeda5ccbe5d3f3fbc40540651654d408f5",
        "0x5543df729c0000",
        "0x0",
        "0x0"
      ],
      "signature": [
        "0x436daf225278c6c10cec26e401b6a68b13b79ee53f46f76188244ed94efa34e",
        "0x2269e54e844d7ac323279249baa5708ebbcd3b98493c7e6626862e17674efd8"
      ],
      "transaction_hash": "0x24a45d2690614692451862c5b0249567854c6731a0a9e9aef236c643ce9abaf",
      "max_fee": "0x0",
      "type": "INVOKE_FUNCTION"
    },
    {
      "contract_address": "0x1068104d5f1be3d69101835c6bf302172744102f8ab0c01f85741fe586a6af8",
      "entry_point_selector": "0x15d40a3d6ca2ac30f4031e42be28da9b056fef9bb7357ac5e85627ee876e5ad",
      "entry_point_type": "EXTERNAL",
      "calldata": [
        "0x1",
        "0x197f9e93cfaf7068ca2daf3ec89c2b91d051505c2231a0a0b9f70801a91fb24",
        "0x3da50d20719cf5809ea34ac89b41e7fceaecbc2204e5da6b33967fd81d47362",
        "0x0",
        "0x4",
        "0x4",
        "0x474f45524c492d4d41535445522d31",
        "0x8aa7c51a6d380f4d9e273add4298d913416031ec",
        "0x8ac7230489e80000",
        "0x8aa7c51a6d380f4d9e273add4298d913416031ec",
        "0x9"
      ],
      "signature": [
        "0x5909ccfd8a2515f8fbbaf0c0e95dab4faf2ed1224d72762c40917311024162f",
        "0x2db1c64ee5c348859e613c4b24029612c82e2c40ed9b07103cc8ef7701bb410"
      ],
      "transaction_hash": "0x57a333bfccf30465cf287460c9c4bb7b21645213bc9cca7fbe99e1b9167d202",
      "max_fee": "0x0",
      "type": "INVOKE_FUNCTION"
    }
  ],
  "timestamp": 1652698140,
  "sequencer_address": "0x46a89ae102987331d369645031b49c27738ed096f2789c24449966da4c6de6b",
  "transaction_receipts": [
    {
      "transaction_index": 0,
      "transaction_hash": "0x24a45d2690614692451862c5b0249567854c6731a0a9e9aef236c643ce9abaf",
      "l2_to_l1_messages": [
        {
          "from_address": "0x73314940630fd6dcda0d772d4c972c4e0a9946bef9dabf4ef84eda8ef542b82",
          "to_address": "0xae0Ee0A63A2cE6BaeEFFE56e7714FB4EFE48D419",
          "payload": [
            "0x0",
            "0x9b872ffeda5ccbe5d3f3fbc40540651654d408f5",
            "0x5543df729c0000",
            "0x0"
          ]
        }
      ],
      "events": [
        {
          "from_address": "0x73314940630fd6dcda0d772d4c972c4e0a9946bef9dabf4ef84eda8ef542b82",
          "keys": [
            "0x194fc63c49b0f07c8e7a78476844837255213824bd6cb81e0ccfb949921aad1"
          ],
          "data": [
            "0x9b872ffeda5ccbe5d3f3fbc40540651654d408f5",
            "0x5543df729c0000",
            "0x0",
            "0x4aef408ac955fd9e5f1397ae9801c1b03344587231f454a0b49d0b99e77c937"
          ]
        },
        {
          "from_address": "0x4aef408ac955fd9e5f1397ae9801c1b03344587231f454a0b49d0b99e77c937",
          "keys": [
            "0x5ad857f66a5b55f1301ff1ed7e098ac6d4433148f0b72ebc4a2945ab85ad53"
          ],
          "data": [
            "0x24a45d2690614692451862c5b0249567854c6731a0a9e9aef236c643ce9abaf",
            "0x0"
          ]
        }
      ],
      "execution_resources": {
        "n_steps": 1348,
        "builtin_instance_counter": {
          "pedersen_builtin": 2,
          "range_check_builtin": 35,
          "output_builtin": 0,
          "ecdsa_builtin": 1,
          "bitwise_builtin": 0,
          "ec_op_builtin": 0
        },
        "n_memory_holes": 44
      },
      "actual_fee": "0x0"
    },
    {
      "transaction_index": 8,
      "transaction_hash": "0x57a333bfccf30465cf287460c9c4bb7b21645213bc9cca7fbe99e1b9167d202",
      "l2_to_l1_messages": [],
      "events": [
        {
          "from_address": "0x52713f43368f9f8ca407174f7bf44f68b6cba77f1fa386d320c0bb096145675",
          "keys": [
            "0x99cd8bde557814842a3121e8ddfd433a539b8c9f14bf31ebf108d12e6196e9"
          ],
          "data": [
            "0x1068104d5f1be3d69101835c6bf302172744102f8ab0c01f85741fe586a6af8",
            "0x0",
            "0x8ac7230489e80000",
            "0x0"
          ]
        },
        {
          "from_address": "0x197f9e93cfaf7068ca2daf3ec89c2b91d051505c2231a0a0b9f70801a91fb24",
          "keys": [
            "0x2f988de39be0ebaa4ef3701988d8affa01403c00f22537d314abcb111ae9c86"
          ],
          "data": [
            "0x474f45524c492d534c4156452d535441524b4e45542d31",
            "0x474f45524c492d4d41535445522d31",
            "0x8aa7c51a6d380f4d9e273add4298d913416031ec",
            "0x8aa7c51a6d380f4d9e273add4298d913416031ec",
            "0x8ac7230489e80000",
            "0xd",
            "0x62822c1c"
          ]
        },
        {
          "from_address": "0x1068104d5f1be3d69101835c6bf302172744102f8ab0c01f85741fe586a6af8",
          "keys": [
            "0x5ad857f66a5b55f1301ff1ed7e098ac6d4433148f0b72ebc4a2945ab85ad53"
          ],
          "data": [
            "0x57a333bfccf30465cf287460c9c4bb7b21645213bc9cca7fbe99e1b9167d202",
            "0x0"
          ]
        }
      ],
      "execution_resources": {
        "n_steps": 1615,
        "builtin_instance_counter": {
          "pedersen_builtin": 8,
          "range_check_builtin": 41,
          "ecdsa_builtin": 1,
          "output_builtin": 0,
          "bitwise_builtin": 0,
          "ec_op_builtin": 0
        },
        "n_memory_holes": 91
      },
      "actual_fee": "0x0"
    }
  ]
}
`

func Test_teleportListener(t *testing.T) {
	ctx, cancelFunc := context.WithTimeout(context.Background(), time.Second)
	defer cancelFunc()

	cli := &mocks.Sequencer{}
	tl := NewTeleportListener(TeleportListenerConfig{
		Sequencer:   cli,
		Addresses:   []*starknet.Felt{starknet.HexToFelt("0x197f9e93cfaf7068ca2daf3ec89c2b91d051505c2231a0a0b9f70801a91fb24")},
		Interval:    time.Millisecond * 100,
		BlocksDelta: []int{10},
		BlocksLimit: 2,
		Logger:      null.New(),
	})

	txHash := starknet.HexToFelt("57a333bfccf30465cf287460c9c4bb7b21645213bc9cca7fbe99e1b9167d202")
	block := &starknet.Block{}
	err := json.Unmarshal([]byte(testBlockResponse), block)
	if err != nil {
		panic(err)
	}

	// Fetch the latest block to determine the latest block number:
	cli.On("GetLatestBlock", ctx, mock.Anything, mock.Anything).Return(block, nil).Once()
	// Because BlocksDelta is set to 10, we should fetch the block 10 blocks before the latest block.
	// Because there were no blocks fetched previously, we are fetching also older blocks but no more
	// that defined in thw BlocksLimit parameter.
	cli.On("GetBlockByNumber", ctx, uint64(191504-11)).Return(block, nil).Once()
	cli.On("GetBlockByNumber", ctx, uint64(191504-10)).Return(block, nil).Once()
	// Finally,fetch pending block.
	cli.On("GetPendingBlock", ctx, mock.Anything, mock.Anything).Return(block, nil).Once()

	require.NoError(t, tl.Start(ctx))

	events := 0
	for {
		msg := <-tl.Events()
		events++
		assert.Equal(t, txHash.Bytes(), msg.Index)
		assert.Equal(t, common.FromHex("0x3507a75b6cda5f180fa8e3ddf7bcb967699061a8f95549b73ecd2673dd14aa97"), msg.Data["hash"])
		if events == 3 {
			break
		}
	}
	assert.Equal(t, 3, events)
}
