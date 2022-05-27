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

	"github.com/chronicleprotocol/oracle-suite/internal/starknet"
	"github.com/chronicleprotocol/oracle-suite/internal/starknet/mocks"
	"github.com/chronicleprotocol/oracle-suite/pkg/log/null"
)

const testResponse = `
{
  "block_hash": "0x477c7df1ffe6f73df5760dc566b3d250ac2030566dee729af6e5a1acc41fed2",
  "parent_hash": "0x38ac438303497fb08ccfa4edbfece5a6aaf7e0b86b51e90306e04d028bb1d7a",
  "block_number": 2427,
  "status": "ACCEPTED_ON_L2",
  "sequencer": "0x21f4b90b0377c82bf330b7b5295820769e72d79d8acd0effa0ebde6e9988bc5",
  "new_root": "0x69d5c24991aac56beabd37a0598a561cd3b0becd419c38f7ab2cd496d889d2f",
  "old_root": "0x5f83732887317cf0262b726924af6c42181207dc02fe5ae7ec3acd5a51aab62",
  "accepted_time": 1653611681,
  "gas_price": "0x8d6e3b615",
  "transactions": [
    {
      "txn_hash": "0x57a333bfccf30465cf287460c9c4bb7b21645213bc9cca7fbe99e1b9167d202",
      "contract_address": "0x1068104d5f1be3d69101835c6bf302172744102f8ab0c01f85741fe586a6af8",
      "entry_point_selector": "0x15d40a3d6ca2ac30f4031e42be28da9b056fef9bb7357ac5e85627ee876e5ad",
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
      "status": "ACCEPTED_ON_L2",
      "status_data": "",
      "messages_sent": [],
      "l1_origin_message": {},
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
      ]
    }
  ]
}
`

func Test_wormholeListener(t *testing.T) {
	ctx, cancelFunc := context.WithCancel(context.Background())
	cli := &mocks.Client{}

	w := NewWormholeListener(WormholeListenerConfig{
		Client:       cli,
		Addresses:    []*starknet.Felt{starknet.HexToFelt("0x197f9e93cfaf7068ca2daf3ec89c2b91d051505c2231a0a0b9f70801a91fb24")},
		Interval:     time.Millisecond * 100,
		BlocksBehind: 0,
		MaxBlocks:    1,
		Logger:       null.New(),
	})

	txHash := starknet.HexToFelt("57a333bfccf30465cf287460c9c4bb7b21645213bc9cca7fbe99e1b9167d202")
	block := &starknet.Block{}
	err := json.Unmarshal([]byte(testResponse), block)
	if err != nil {
		panic(err)
	}

	cli.On("BlockNumber", ctx).Return(uint64(42), nil).Once()
	cli.On("GetBlockByNumber", ctx, mock.Anything, mock.Anything).Return(block, nil).Once().Run(func(args mock.Arguments) {
		bn := args.Get(1).(uint64)
		sc := args.Get(2).(starknet.Scope)
		assert.Equal(t, uint64(42), bn)
		assert.Equal(t, starknet.ScopeFullTXNAndReceipts, sc)
	})

	require.NoError(t, w.Start(ctx))
	for {
		if len(cli.Calls) >= 2 { // 2 is the number of mocked calls above.
			cancelFunc()
			break
		}
		time.Sleep(time.Millisecond * 10)
	}
	events := 0
	for len(w.Events()) > 0 {
		events++
		msg := <-w.Events()
		assert.Equal(t, txHash.Bytes(), msg.Index)
		assert.Equal(t, common.FromHex("0x3507a75b6cda5f180fa8e3ddf7bcb967699061a8f95549b73ecd2673dd14aa97"), msg.Data["hash"])
	}
	assert.Equal(t, 1, events)
}
