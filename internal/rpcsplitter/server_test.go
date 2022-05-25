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

package rpcsplitter

import (
	"encoding/json"
	"errors"
	"testing"
	"time"
)

var blockWithHashesResp = json.RawMessage(`
	{
		"difficulty": "0x2d50ba175407",
		"extraData": "0xe4b883e5bda9e7a59ee4bb99e9b1bc",
		"gasLimit": "0x47e7c4",
		"gasUsed": "0x5208",
		"hash": "0xc0f4906fea23cf6f3cce98cb44e8e1449e455b28d684dfa9ff65426495584de6",
		"logsBloom": "0x00000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000",
		"miner": "0x61c808d82a3ac53231750dadc13c777b59310bd9",
		"mixHash": "0xc38853328f753c455edaa4dfc6f62a435e05061beac136c13dbdcd0ff38e5f40",
		"nonce": "0x3b05c6d5524209f1",
		"number": "0x1e8480",
		"parentHash": "0x57ebf07eb9ed1137d41447020a25e51d30a0c272b5896571499c82c33ecb7288",
		"receiptsRoot": "0x84aea4a7aad5c5899bd5cfc7f309cc379009d30179316a2a7baa4a2ea4a438ac",
		"sha3Uncles": "0x1dcc4de8dec75d7aab85b567b6ccd41ad312451b948a7413f0a142fd40d49347",
		"size": "0x28a",
		"stateRoot": "0x96dbad955b166f5119793815c36f11ffa909859bbfeb64b735cca37cbf10bef1",
		"timestamp": "0x57a1118a",
		"totalDifficulty": "0x262c34a6fd1268f6c",
		"transactions": [
		  "0xc55e2b90168af6972193c1f86fa4d7d7b31a29c156665d15b9cd48618b5177ef"
		],
		"transactionsRoot": "0xb31f174d27b99cdae8e746bd138a01ce60d8dd7b224f7c60845914def05ecc58",
		"uncles": [
			"0xc0f4906fea23cf6f3cce98cb44e8e1449e455b28d684dfa9ff65426495584de6"
		]
	}
`)

var blockWithObjectsResp = json.RawMessage(`
	{
		"difficulty": "0x2d50ba175407",
		"extraData": "0xe4b883e5bda9e7a59ee4bb99e9b1bc",
		"gasLimit": "0x47e7c4",
		"gasUsed": "0x5208",
		"hash": "0xc0f4906fea23cf6f3cce98cb44e8e1449e455b28d684dfa9ff65426495584de6",
		"logsBloom": "0x00000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000",
		"miner": "0x61c808d82a3ac53231750dadc13c777b59310bd9",
		"mixHash": "0xc38853328f753c455edaa4dfc6f62a435e05061beac136c13dbdcd0ff38e5f40",
		"nonce": "0x3b05c6d5524209f1",
		"number": "0x1e8480",
		"parentHash": "0x57ebf07eb9ed1137d41447020a25e51d30a0c272b5896571499c82c33ecb7288",
		"receiptsRoot": "0x84aea4a7aad5c5899bd5cfc7f309cc379009d30179316a2a7baa4a2ea4a438ac",
		"sha3Uncles": "0x1dcc4de8dec75d7aab85b567b6ccd41ad312451b948a7413f0a142fd40d49347",
		"size": "0x28a",
		"stateRoot": "0x96dbad955b166f5119793815c36f11ffa909859bbfeb64b735cca37cbf10bef1",
		"timestamp": "0x57a1118a",
		"totalDifficulty": "0x262c34a6fd1268f6c",
		"transactions": [
			{
				"blockHash": "0xc0f4906fea23cf6f3cce98cb44e8e1449e455b28d684dfa9ff65426495584de6",
				"blockNumber": "0x1e8480",
				"from": "0x32be343b94f860124dc4fee278fdcbd38c102d88",
				"gas": "0x51615",
				"gasPrice": "0x6fc23ac00",
				"hash": "0xc55e2b90168af6972193c1f86fa4d7d7b31a29c156665d15b9cd48618b5177ef",
				"input": "0x",
				"nonce": "0x1efc5",
				"to": "0x104994f45d9d697ca104e5704a7b77d7fec3537c",
				"transactionIndex": "0x0",
				"value": "0x821878651a4d70000",
				"v": "0x1b",
				"r": "0x51222d91a379452395d0abaff981af4cfcc242f25cfaf947dea8245a477731f9",
				"s": "0x3a997c910b4701cca5d933fb26064ee5af7fe3236ff0ef2b58aa50b25aff8ca5"
			}
		],
		"transactionsRoot": "0xb31f174d27b99cdae8e746bd138a01ce60d8dd7b224f7c60845914def05ecc58",
		"uncles": [
			"0xc0f4906fea23cf6f3cce98cb44e8e1449e455b28d684dfa9ff65426495584de6"
		]
	}
`)

var transaction1Resp = json.RawMessage(`
	{
		"hash": "0x88df016429689c079f3b2f6ad39fa052532c56795b733da78a91ebe6a713944b",
		"blockHash": "0x1d59ff54b1eb26b013ce3cb5fc9dab3705b415a67127a003c3e61eb445bb8df2",
		"blockNumber": "0x5daf3b",
		"from": "0xa7d9ddbe1f17865597fbd27ec712455208b6b76d",
		"gas": "0xc350",
		"gasPrice": "0x4a817c800",
		"input": "0x68656c6c6f21",
		"nonce": "0x15",
		"r": "0x1b5e176d927f8e9ab405058b2d2457392da3e20f328b16ddabcebc33eaac5fea",
		"s": "0x4ba69724e8f69de52f0125ad8b3c5c2cef33019bac3249e2c0a2192766d1721c",
		"to": "0xf02c1c8e6114b1dbe8937a39260b5b0a374432bb",
		"transactionIndex": "0x41",
		"v": "0x25",
		"value": "0xf3dbb76162000"
	}
`)

var transaction2Resp = json.RawMessage(`
	{
		"blockHash": "0xc0f4906fea23cf6f3cce98cb44e8e1449e455b28d684dfa9ff65426495584de6",
		"blockNumber": "0x1e8480",
		"from": "0x32be343b94f860124dc4fee278fdcbd38c102d88",
		"gas": "0x51615",
		"gasPrice": "0x6fc23ac00",
		"hash": "0xc55e2b90168af6972193c1f86fa4d7d7b31a29c156665d15b9cd48618b5177ef",
		"input": "0x",
		"nonce": "0x1efc5",
		"to": "0x104994f45d9d697ca104e5704a7b77d7fec3537c",
		"transactionIndex": "0x0",
		"value": "0x821878651a4d70000",
		"v": "0x1b",
		"r": "0x51222d91a379452395d0abaff981af4cfcc242f25cfaf947dea8245a477731f9",
		"s": "0x3a997c910b4701cca5d933fb26064ee5af7fe3236ff0ef2b58aa50b25aff8ca5"
	}
`)

var transactionReceipt1Resp = json.RawMessage(`
	{
		"transactionHash": "0xab059a62e22e230fe0f56d8555340a29b2e9532360368f810595453f6fdd213b",
		"blockHash": "0x8243343df08b9751f5ca0c5f8c9c0460d8a9b6351066fae0acbd4d3e776de8bb",
		"blockNumber": "0x429d3b",
		"contractAddress": null,
		"cumulativeGasUsed": "0x64b559",
		"from": "0x00b46c2526e227482e2ebb8f4c69e4674d262e75",
		"gasUsed": "0xcaac",
		"logs": [
			{
				"blockHash": "0x8243343df08b9751f5ca0c5f8c9c0460d8a9b6351066fae0acbd4d3e776de8bb",
				"address": "0xb59f67a8bff5d8cd03f6ac17265c550ed8f33907",
				"logIndex": "0x56",
				"data": "0x000000000000000000000000000000000000000000000000000000012a05f200",
				"removed": false,
				"topics": [
					"0xddf252ad1be2c89b69c2b068fc378daa952ba7f163c4a11628f55a4df523b3ef",
					"0x00000000000000000000000000b46c2526e227482e2ebb8f4c69e4674d262e75",
					"0x00000000000000000000000054a2d42a40f51259dedd1978f6c118a0f0eff078"
				],
				"blockNumber": "0x429d3b",
				"transactionIndex": "0xac",
				"transactionHash": "0xab059a62e22e230fe0f56d8555340a29b2e9532360368f810595453f6fdd213b"
			}
		],
		"logsBloom": "0x00000000040000000000000000000000000000000000000000000000000000080000000010000000000000000000000000000000000040000000000000000000000000000000000000000008000000000000000000000000000000000000000000000000000000000000000000002000000000000000000000000010000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000002000000000000000000000000000000010100000000000000000000000000004000000000000200000000000000000000000000000000000000000000",
		"root": "0x3ccba97c7fcc7e1636ce2d44be1a806a8999df26eab80a928205714a878d5114",
		"status": null,
		"to": "0xb59f67a8bff5d8cd03f6ac17265c550ed8f33907",
		"transactionIndex": "0xac"
	}
`)

var feeHistory1Resp = json.RawMessage(`
	{
		"oldestBlock": "0xc72641",
		"reward": [
			[
				"0x4a817c7ee",
				"0x4a817c7ee"
			], [
				"0x773593f0",
				"0x773593f5"
			], [
				"0x0",
				"0x0"
			], [
				"0x773593f5",
				"0x773bae75"
			]
		],
		"baseFeePerGas": [
			"0x12",
			"0x10",
			"0x10",
			"0xe",
			"0xd"
		],
		"gasUsedRatio": [
			0.026089875,
			0.406803,
			0,
			0.0866665
		]
	}
`)

var feeHistory2Resp = json.RawMessage(`
	{
		"oldestBlock": "0xC72641",
		"baseFeePerGas": [
			"0x92db30f56",
			"0x9a47da3c5",
			"0x8fb856b5b",
			"0xa1a3c78d9",
			"0x91a6775ac",
			"0x7f71a86f7"
		],
		"gasUsedRatio": [
			0.7022238670892842,
			0.2261976964422899,
			0.9987387,
			0.10431753273738473,
			0
		]
	}
`)

var getLogs1Resp = json.RawMessage(`
[
	{
		"blockHash": "0x8243343df08b9751f5ca0c5f8c9c0460d8a9b6351066fae0acbd4d3e776de8bb",
		"address": "0xb59f67a8bff5d8cd03f6ac17265c550ed8f33907",
		"logIndex": "0x56",
		"data": "0x000000000000000000000000000000000000000000000000000000012a05f200",
		"removed": false,
		"topics": [
			"0xddf252ad1be2c89b69c2b068fc378daa952ba7f163c4a11628f55a4df523b3ef",
			"0x00000000000000000000000000b46c2526e227482e2ebb8f4c69e4674d262e75",
			"0x00000000000000000000000054a2d42a40f51259dedd1978f6c118a0f0eff078"
		],
		"blockNumber": "0x429d3b",
		"transactionIndex": "0xac",
		"transactionHash": "0xab059a62e22e230fe0f56d8555340a29b2e9532360368f810595453f6fdd213b"
	}
]
`)

var getLogs2Resp = json.RawMessage(`
[
	{
		"logIndex": "0x0",
		"removed": false,
		"blockNumber": "0x21c",
		"blockHash": "0xc7e6c9d5b9f522b2c9d2991546be0a8737e587beb6628c056f3c327a44b45132",
		"transactionHash": "0xfd1a40f9fbf89c97b4545ec9db774c85e51dd8a3545f969418a22f9cb79417c5",
		"transactionIndex": "0x0",
		"address": "0x42699a7612a82f1d9c36148af9c77354759b210b",
		"data": "0x0000000000000000000000000000000000000000000000000000000000000005",
		"topics": [
			"0xd3610b1c54575b7f4f0dc03d210b8ac55624ae007679b7a928a4f25a709331a8",
			"0x0000000000000000000000000000000000000000000000000000000000000005"
		]
	}
]
`)

func Test_RPC_BlockNumber(t *testing.T) {
	t.Run("in-range", func(t *testing.T) {
		prepareHandlerTest(t, 3, "eth_blockNumber").
			setOptions(WithRequirements(2, 2)).
			mockClientCall(0, `0x4`, "eth_blockNumber").
			mockClientCall(1, `0x5`, "eth_blockNumber").
			mockClientCall(2, `0x6`, "eth_blockNumber").
			expectedResult(`0x4`).
			test()
	})
	t.Run("outside-range", func(t *testing.T) {
		prepareHandlerTest(t, 3, "eth_blockNumber").
			setOptions(WithRequirements(2, 1)).
			mockClientCall(0, `0x1`, "eth_blockNumber").
			mockClientCall(1, `0x5`, "eth_blockNumber").
			mockClientCall(2, `0x6`, "eth_blockNumber").
			expectedResult(`0x5`).
			test()
	})
	t.Run("one-failed", func(t *testing.T) {
		prepareHandlerTest(t, 3, "eth_blockNumber").
			setOptions(WithRequirements(2, 10)).
			mockClientCall(0, `0x3`, "eth_blockNumber").
			mockClientCall(1, `0x4`, "eth_blockNumber").
			mockClientCall(2, errors.New("error#1"), "eth_blockNumber").
			expectedResult(`0x3`).
			test()
	})
	t.Run("two-failed", func(t *testing.T) {
		prepareHandlerTest(t, 3, "eth_blockNumber").
			setOptions(WithRequirements(2, 10)).
			mockClientCall(0, `0x3`, "eth_blockNumber").
			mockClientCall(1, errors.New("error#1"), "eth_blockNumber").
			mockClientCall(2, errors.New("error#2"), "eth_blockNumber").
			expectedError("error#1").
			expectedError("error#2").
			test()
	})
}

func Test_RPC_GetBlockByHash(t *testing.T) {
	blockHash := newHash("0xc0f4906fea23cf6f3cce98cb44e8e1449e455b28d684dfa9ff65426495584de6")
	t.Run("with-hashes", func(t *testing.T) {
		prepareHandlerTest(t, 3, "eth_getBlockByHash", blockHash, false).
			setOptions(WithRequirements(2, 10)).
			mockClientCall(0, blockWithHashesResp, "eth_getBlockByHash", blockHash, false).
			mockClientCall(1, blockWithHashesResp, "eth_getBlockByHash", blockHash, false).
			mockClientCall(2, blockWithHashesResp, "eth_getBlockByHash", blockHash, false).
			expectedResult(blockWithHashesResp).
			test()
	})
	t.Run("with-objects", func(t *testing.T) {
		prepareHandlerTest(t, 3, "eth_getBlockByHash", blockHash, true).
			setOptions(WithRequirements(2, 10)).
			mockClientCall(0, blockWithObjectsResp, "eth_getBlockByHash", blockHash, true).
			mockClientCall(1, blockWithObjectsResp, "eth_getBlockByHash", blockHash, true).
			mockClientCall(2, blockWithObjectsResp, "eth_getBlockByHash", blockHash, true).
			expectedResult(blockWithObjectsResp).
			test()
	})
	t.Run("one-failed", func(t *testing.T) {
		prepareHandlerTest(t, 3, "eth_getBlockByHash", blockHash, false).
			setOptions(WithRequirements(2, 10)).
			mockClientCall(0, blockWithHashesResp, "eth_getBlockByHash", blockHash, false).
			mockClientCall(1, blockWithHashesResp, "eth_getBlockByHash", blockHash, false).
			mockClientCall(2, errors.New("error#1"), "eth_getBlockByHash", blockHash, false).
			expectedResult(blockWithHashesResp).
			test()
	})
	t.Run("two-failed", func(t *testing.T) {
		prepareHandlerTest(t, 3, "eth_getBlockByHash", blockHash, false).
			setOptions(WithRequirements(2, 10)).
			mockClientCall(0, blockWithHashesResp, "eth_getBlockByHash", blockHash, false).
			mockClientCall(1, errors.New("error#1"), "eth_getBlockByHash", blockHash, false).
			mockClientCall(2, errors.New("error#2"), "eth_getBlockByHash", blockHash, false).
			expectedError("error#1").
			expectedError("error#2").
			test()
	})
	t.Run("different-responses", func(t *testing.T) {
		prepareHandlerTest(t, 2, "eth_getBlockByHash", blockHash, false).
			setOptions(WithRequirements(2, 10)).
			mockClientCall(0, blockWithHashesResp, "eth_getBlockByHash", blockHash, false).
			mockClientCall(1, blockWithObjectsResp, "eth_getBlockByHash", blockHash, false).
			expectedError("").
			test()
	})
}

func Test_RPC_GetBlockByNumber(t *testing.T) {
	blockNumber := newNumber("0x1e8480")
	t.Run("with-hashes", func(t *testing.T) {
		prepareHandlerTest(t, 3, "eth_getBlockByNumber", blockNumber, false).
			setOptions(WithRequirements(2, 10)).
			mockClientCall(0, blockWithHashesResp, "eth_getBlockByNumber", blockNumber, false).
			mockClientCall(1, blockWithHashesResp, "eth_getBlockByNumber", blockNumber, false).
			mockClientCall(2, blockWithHashesResp, "eth_getBlockByNumber", blockNumber, false).
			expectedResult(blockWithHashesResp).
			test()
	})
	t.Run("with-objects", func(t *testing.T) {
		prepareHandlerTest(t, 3, "eth_getBlockByNumber", blockNumber, true).
			setOptions(WithRequirements(2, 10)).
			mockClientCall(0, blockWithObjectsResp, "eth_getBlockByNumber", blockNumber, true).
			mockClientCall(1, blockWithObjectsResp, "eth_getBlockByNumber", blockNumber, true).
			mockClientCall(2, blockWithObjectsResp, "eth_getBlockByNumber", blockNumber, true).
			expectedResult(blockWithObjectsResp).
			test()
	})
	t.Run("one-failed", func(t *testing.T) {
		prepareHandlerTest(t, 3, "eth_getBlockByNumber", blockNumber, false).
			setOptions(WithRequirements(2, 10)).
			mockClientCall(0, blockWithHashesResp, "eth_getBlockByNumber", blockNumber, false).
			mockClientCall(1, blockWithHashesResp, "eth_getBlockByNumber", blockNumber, false).
			mockClientCall(2, errors.New("error#1"), "eth_getBlockByNumber", blockNumber, false).
			expectedResult(blockWithHashesResp).
			test()
	})
	t.Run("two-failed", func(t *testing.T) {
		prepareHandlerTest(t, 3, "eth_getBlockByNumber", blockNumber, false).
			setOptions(WithRequirements(2, 10)).
			mockClientCall(0, blockWithHashesResp, "eth_getBlockByNumber", blockNumber, false).
			mockClientCall(1, errors.New("error#1"), "eth_getBlockByNumber", blockNumber, false).
			mockClientCall(2, errors.New("error#2"), "eth_getBlockByNumber", blockNumber, false).
			expectedError("error#1").
			expectedError("error#2").
			test()
	})
	t.Run("different-responses", func(t *testing.T) {
		prepareHandlerTest(t, 2, "eth_getBlockByNumber", blockNumber, false).
			setOptions(WithRequirements(2, 10)).
			mockClientCall(0, blockWithHashesResp, "eth_getBlockByNumber", blockNumber, false).
			mockClientCall(1, blockWithObjectsResp, "eth_getBlockByNumber", blockNumber, false).
			expectedError("").
			test()
	})
}

func Test_RPC_GetTransactionByHash(t *testing.T) {
	txHash := newHash("0x88df016429689c079f3b2f6ad39fa052532c56795b733da78a91ebe6a713944b")
	t.Run("simple", func(t *testing.T) {
		prepareHandlerTest(t, 3, "eth_getTransactionByHash", txHash).
			setOptions(WithRequirements(2, 10)).
			mockClientCall(0, transaction1Resp, "eth_getTransactionByHash", txHash).
			mockClientCall(1, transaction1Resp, "eth_getTransactionByHash", txHash).
			mockClientCall(2, transaction1Resp, "eth_getTransactionByHash", txHash).
			expectedResult(transaction1Resp).
			test()
	})
	t.Run("one-failed", func(t *testing.T) {
		prepareHandlerTest(t, 3, "eth_getTransactionByHash", txHash).
			setOptions(WithRequirements(2, 10)).
			mockClientCall(0, transaction1Resp, "eth_getTransactionByHash", txHash).
			mockClientCall(1, transaction1Resp, "eth_getTransactionByHash", txHash).
			mockClientCall(2, errors.New("error#1"), "eth_getTransactionByHash", txHash).
			expectedResult(transaction1Resp).
			test()
	})
	t.Run("two-failed", func(t *testing.T) {
		prepareHandlerTest(t, 3, "eth_getTransactionByHash", txHash).
			setOptions(WithRequirements(2, 10)).
			mockClientCall(0, transaction1Resp, "eth_getTransactionByHash", txHash).
			mockClientCall(1, errors.New("error#1"), "eth_getTransactionByHash", txHash).
			mockClientCall(2, errors.New("error#2"), "eth_getTransactionByHash", txHash).
			expectedError("error#1").
			expectedError("error#2").
			test()
	})
	t.Run("different-responses", func(t *testing.T) {
		prepareHandlerTest(t, 2, "eth_getTransactionByHash", txHash).
			setOptions(WithRequirements(2, 10)).
			mockClientCall(0, transaction1Resp, "eth_getTransactionByHash", txHash).
			mockClientCall(1, transaction2Resp, "eth_getTransactionByHash", txHash).
			expectedError("").
			test()
	})
}

func Test_RPC_GetTransactionCount(t *testing.T) {
	address := newAddress("0xc94770007dda54cF92009BFF0dE90c06F603a09f")
	blockNumber := newBlockID("0x10")
	t.Run("simple", func(t *testing.T) {
		prepareHandlerTest(t, 3, "eth_getTransactionCount", address, blockNumber).
			setOptions(WithRequirements(2, 10)).
			mockClientCall(0, `0x5`, "eth_getTransactionCount", address, blockNumber).
			mockClientCall(1, `0x5`, "eth_getTransactionCount", address, blockNumber).
			mockClientCall(2, `0x5`, "eth_getTransactionCount", address, blockNumber).
			expectedResult(`0x5`).
			test()
	})
	t.Run("one-failed", func(t *testing.T) {
		prepareHandlerTest(t, 3, "eth_getTransactionCount", address, blockNumber).
			setOptions(WithRequirements(2, 10)).
			mockClientCall(0, `0x5`, "eth_getTransactionCount", address, blockNumber).
			mockClientCall(1, `0x5`, "eth_getTransactionCount", address, blockNumber).
			mockClientCall(2, errors.New("error#1"), "eth_getTransactionCount", address, blockNumber).
			expectedResult(`0x5`).
			test()
	})
	t.Run("two-failed", func(t *testing.T) {
		prepareHandlerTest(t, 3, "eth_getTransactionCount", address, blockNumber).
			setOptions(WithRequirements(2, 10)).
			mockClientCall(0, `0x5`, "eth_getTransactionCount", address, blockNumber).
			mockClientCall(1, errors.New("error#1"), "eth_getTransactionCount", address, blockNumber).
			mockClientCall(2, errors.New("error#2"), "eth_getTransactionCount", address, blockNumber).
			expectedError("error#1").
			expectedError("error#2").
			test()
	})
	t.Run("latest-block", func(t *testing.T) {
		prepareHandlerTest(t, 2, "eth_getTransactionCount", address, newBlockID("latest")).
			setOptions(WithRequirements(2, 10)).
			mockClientCall(0, blockNumber, "eth_blockNumber").
			mockClientCall(1, blockNumber, "eth_blockNumber").
			mockClientCall(0, `0x5`, "eth_getTransactionCount", address, blockNumber).
			mockClientCall(1, `0x5`, "eth_getTransactionCount", address, blockNumber).
			expectedResult(`0x5`).
			test()
	})
	t.Run("pending-block", func(t *testing.T) {
		prepareHandlerTest(t, 2, "eth_getTransactionCount", address, newBlockID("pending")).
			setOptions(WithRequirements(2, 10)).
			mockClientCall(0, blockNumber, "eth_blockNumber").
			mockClientCall(1, blockNumber, "eth_blockNumber").
			mockClientCall(0, `0x5`, "eth_getTransactionCount", address, blockNumber).
			mockClientCall(1, `0x5`, "eth_getTransactionCount", address, blockNumber).
			expectedResult(`0x5`).
			test()
	})
	t.Run("earliest-block", func(t *testing.T) {
		prepareHandlerTest(t, 2, "eth_getTransactionCount", address, newBlockID("earliest")).
			setOptions(WithRequirements(2, 10)).
			expectedError("").
			test()
	})
}

func Test_RPC_GetTransactionReceipt(t *testing.T) {
	txHash := newHash("0xab059a62e22e230fe0f56d8555340a29b2e9532360368f810595453f6fdd213b")
	t.Run("simple", func(t *testing.T) {
		prepareHandlerTest(t, 3, "eth_getTransactionReceipt", txHash).
			setOptions(WithRequirements(2, 10)).
			mockClientCall(0, transactionReceipt1Resp, "eth_getTransactionReceipt", txHash).
			mockClientCall(1, transactionReceipt1Resp, "eth_getTransactionReceipt", txHash).
			mockClientCall(2, transactionReceipt1Resp, "eth_getTransactionReceipt", txHash).
			expectedResult(transactionReceipt1Resp).
			test()
	})
	t.Run("one-failed", func(t *testing.T) {
		prepareHandlerTest(t, 3, "eth_getTransactionReceipt", txHash).
			setOptions(WithRequirements(2, 10)).
			mockClientCall(0, transactionReceipt1Resp, "eth_getTransactionReceipt", txHash).
			mockClientCall(1, transactionReceipt1Resp, "eth_getTransactionReceipt", txHash).
			mockClientCall(2, errors.New("error#1"), "eth_getTransactionReceipt", txHash).
			expectedResult(transactionReceipt1Resp).
			test()
	})
	t.Run("two-failed", func(t *testing.T) {
		prepareHandlerTest(t, 3, "eth_getTransactionReceipt", txHash).
			setOptions(WithRequirements(2, 10)).
			mockClientCall(0, transactionReceipt1Resp, "eth_getTransactionReceipt", txHash).
			mockClientCall(1, errors.New("error#1"), "eth_getTransactionReceipt", txHash).
			mockClientCall(2, errors.New("error#2"), "eth_getTransactionReceipt", txHash).
			expectedError("error#1").
			expectedError("error#2").
			test()
	})
}

func Test_RPC_GetBlockTransactionCountByHash(t *testing.T) {
	t.Run("simple", func(t *testing.T) {
		prepareHandlerTest(t, 3, "eth_getBlockTransactionCountByHash").
			setOptions(WithRequirements(2, 10)).
			expectedError("the method eth_getBlockTransactionCountByHash does not exist").
			test()
	})
}

func Test_RPC_GetBlockTransactionCountByNumber(t *testing.T) {
	t.Run("simple", func(t *testing.T) {
		prepareHandlerTest(t, 3, "eth_getBlockTransactionCountByNumber").
			setOptions(WithRequirements(2, 10)).
			expectedError("the method eth_getBlockTransactionCountByNumber does not exist").
			test()
	})
}

func Test_RPC_GetTransactionByBlockHashAndIndex(t *testing.T) {
	t.Run("simple", func(t *testing.T) {
		prepareHandlerTest(t, 3, "eth_getTransactionByBlockHashAndIndex").
			setOptions(WithRequirements(2, 10)).
			expectedError("the method eth_getTransactionByBlockHashAndIndex does not exist").
			test()
	})
}

func Test_RPC_GetTransactionByBlockNumberAndIndex(t *testing.T) {
	t.Run("simple", func(t *testing.T) {
		prepareHandlerTest(t, 3, "eth_getTransactionByBlockNumberAndIndex").
			setOptions(WithRequirements(2, 10)).
			expectedError("the method eth_getTransactionByBlockNumberAndIndex does not exist").
			test()
	})
}

func Test_RPC_SendRawTransaction(t *testing.T) {
	txData := newBytes("0xd46e8dd67c5d32be8d46e8dd67c5d32be8058bb8eb970870f072445675058bb8eb970870f072445675")
	txHash1 := newHash("0xe670ec64341771606e55d6b4ca35a1a6b75ee3d5145a99d05921026d15273310")
	txHash2 := newHash("0xc55e2b90168af6972193c1f86fa4d7d7b31a29c156665d15b9cd48618b5177ef")
	t.Run("simple", func(t *testing.T) {
		prepareHandlerTest(t, 3, "eth_sendRawTransaction", txData).
			setOptions(WithRequirements(2, 10)).
			mockClientCall(0, txHash1, "eth_sendRawTransaction", txData).
			mockClientCall(1, txHash1, "eth_sendRawTransaction", txData).
			mockClientCall(2, txHash1, "eth_sendRawTransaction", txData).
			expectedResult(txHash1).
			test()
	})
	t.Run("one-failed", func(t *testing.T) {
		prepareHandlerTest(t, 3, "eth_sendRawTransaction", txData).
			setOptions(WithRequirements(2, 10)).
			mockClientCall(0, txHash1, "eth_sendRawTransaction", txData).
			mockClientCall(1, txHash1, "eth_sendRawTransaction", txData).
			mockClientCall(2, errors.New("error#1"), "eth_sendRawTransaction", txData).
			expectedResult(txHash1).
			test()
	})
	t.Run("two-failed", func(t *testing.T) {
		prepareHandlerTest(t, 3, "eth_sendRawTransaction", txData).
			setOptions(WithRequirements(2, 10)).
			mockClientCall(0, txHash1, "eth_sendRawTransaction", txData).
			mockClientCall(1, errors.New("error#1"), "eth_sendRawTransaction", txData).
			mockClientCall(2, errors.New("error#2"), "eth_sendRawTransaction", txData).
			expectedError("error#1").
			expectedError("error#2").
			test()
	})
	t.Run("all-failed", func(t *testing.T) {
		prepareHandlerTest(t, 3, "eth_sendRawTransaction", txData).
			setOptions(WithRequirements(2, 10)).
			mockClientCall(0, errors.New("error#1"), "eth_sendRawTransaction", txData).
			mockClientCall(1, errors.New("error#2"), "eth_sendRawTransaction", txData).
			mockClientCall(2, errors.New("error#3"), "eth_sendRawTransaction", txData).
			expectedError("error#1").
			expectedError("error#2").
			expectedError("error#3").
			test()
	})
	t.Run("different-responses", func(t *testing.T) {
		prepareHandlerTest(t, 2, "eth_sendRawTransaction", txData).
			setOptions(WithRequirements(2, 10)).
			mockClientCall(0, txHash1, "eth_sendRawTransaction", txData).
			mockClientCall(1, txHash2, "eth_sendRawTransaction", txData).
			expectedError("").
			test()
	})
}

func Test_RPC_GetBalance(t *testing.T) {
	address := newAddress("0xb59f67a8bff5d8cd03f6ac17265c550ed8f33907")
	balance := newNumber("0x100000000000")
	blockNumber := newBlockID("0x10")
	t.Run("simple", func(t *testing.T) {
		prepareHandlerTest(t, 3, "eth_getBalance", address, blockNumber).
			setOptions(WithRequirements(2, 10)).
			mockClientCall(0, balance, "eth_getBalance", address, blockNumber).
			mockClientCall(1, balance, "eth_getBalance", address, blockNumber).
			mockClientCall(2, balance, "eth_getBalance", address, blockNumber).
			expectedResult(balance).
			test()
	})
	t.Run("one-failed", func(t *testing.T) {
		prepareHandlerTest(t, 3, "eth_getBalance", address, blockNumber).
			setOptions(WithRequirements(2, 10)).
			mockClientCall(0, balance, "eth_getBalance", address, blockNumber).
			mockClientCall(1, balance, "eth_getBalance", address, blockNumber).
			mockClientCall(2, errors.New("error#1"), "eth_getBalance", address, blockNumber).
			expectedResult(balance).
			test()
	})
	t.Run("two-failed", func(t *testing.T) {
		prepareHandlerTest(t, 3, "eth_getBalance", address, blockNumber).
			setOptions(WithRequirements(2, 10)).
			mockClientCall(0, balance, "eth_getBalance", address, blockNumber).
			mockClientCall(1, errors.New("error#1"), "eth_getBalance", address, blockNumber).
			mockClientCall(2, errors.New("error#2"), "eth_getBalance", address, blockNumber).
			expectedError("error#1").
			expectedError("error#2").
			test()
	})
	t.Run("different-responses", func(t *testing.T) {
		prepareHandlerTest(t, 2, "eth_getBalance", address, blockNumber).
			setOptions(WithRequirements(2, 10)).
			mockClientCall(0, newNumber("0x100000000000"), "eth_getBalance", address, blockNumber).
			mockClientCall(1, newNumber("0x100000000001"), "eth_getBalance", address, blockNumber).
			expectedError("").
			test()
	})
	t.Run("latest-block", func(t *testing.T) {
		prepareHandlerTest(t, 2, "eth_getBalance", address, newBlockID("latest")).
			setOptions(WithRequirements(2, 10)).
			mockClientCall(0, blockNumber, "eth_blockNumber").
			mockClientCall(1, blockNumber, "eth_blockNumber").
			mockClientCall(0, balance, "eth_getBalance", address, blockNumber).
			mockClientCall(1, balance, "eth_getBalance", address, blockNumber).
			expectedResult(balance).
			test()
	})
	t.Run("pending-block", func(t *testing.T) {
		prepareHandlerTest(t, 2, "eth_getBalance", address, newBlockID("pending")).
			setOptions(WithRequirements(2, 10)).
			mockClientCall(0, blockNumber, "eth_blockNumber").
			mockClientCall(1, blockNumber, "eth_blockNumber").
			mockClientCall(0, balance, "eth_getBalance", address, blockNumber).
			mockClientCall(1, balance, "eth_getBalance", address, blockNumber).
			expectedResult(balance).
			test()
	})
	t.Run("earliest-block", func(t *testing.T) {
		prepareHandlerTest(t, 2, "eth_getBalance", address, newBlockID("earliest")).
			setOptions(WithRequirements(2, 10)).
			expectedError("").
			test()
	})
}

func Test_RPC_GetCode(t *testing.T) {
	address := newAddress("0xb59f67a8bff5d8cd03f6ac17265c550ed8f33907")
	code1 := newBytes("0x606060405236156100965763ffffffff")
	code2 := newBytes("0x606060405236156100965763ffffff00")
	blockNumber := newBlockID("0x10")
	t.Run("simple", func(t *testing.T) {
		prepareHandlerTest(t, 3, "eth_getCode", address, blockNumber).
			setOptions(WithRequirements(2, 10)).
			mockClientCall(0, code1, "eth_getCode", address, blockNumber).
			mockClientCall(1, code1, "eth_getCode", address, blockNumber).
			mockClientCall(2, code1, "eth_getCode", address, blockNumber).
			expectedResult(code1).
			test()
	})
	t.Run("one-failed", func(t *testing.T) {
		prepareHandlerTest(t, 3, "eth_getCode", address, blockNumber).
			setOptions(WithRequirements(2, 10)).
			mockClientCall(0, code1, "eth_getCode", address, blockNumber).
			mockClientCall(1, code1, "eth_getCode", address, blockNumber).
			mockClientCall(2, errors.New("error#1"), "eth_getCode", address, blockNumber).
			expectedResult(code1).
			test()
	})
	t.Run("two-failed", func(t *testing.T) {
		prepareHandlerTest(t, 3, "eth_getCode", address, blockNumber).
			setOptions(WithRequirements(2, 10)).
			mockClientCall(0, code1, "eth_getCode", address, blockNumber).
			mockClientCall(1, errors.New("error#1"), "eth_getCode", address, blockNumber).
			mockClientCall(2, errors.New("error#2"), "eth_getCode", address, blockNumber).
			expectedError("error#1").
			expectedError("error#2").
			test()
	})
	t.Run("different-responses", func(t *testing.T) {
		prepareHandlerTest(t, 2, "eth_getCode", address, blockNumber).
			setOptions(WithRequirements(2, 10)).
			mockClientCall(0, code1, "eth_getCode", address, blockNumber).
			mockClientCall(1, code2, "eth_getCode", address, blockNumber).
			expectedError("").
			test()
	})
	t.Run("latest-block", func(t *testing.T) {
		prepareHandlerTest(t, 2, "eth_getCode", address, newBlockID("latest")).
			setOptions(WithRequirements(2, 10)).
			mockClientCall(0, blockNumber, "eth_blockNumber").
			mockClientCall(1, blockNumber, "eth_blockNumber").
			mockClientCall(0, code1, "eth_getCode", address, blockNumber).
			mockClientCall(1, code1, "eth_getCode", address, blockNumber).
			expectedResult(code1).
			test()
	})
	t.Run("pending-block", func(t *testing.T) {
		prepareHandlerTest(t, 2, "eth_getCode", address, newBlockID("pending")).
			setOptions(WithRequirements(2, 10)).
			mockClientCall(0, blockNumber, "eth_blockNumber").
			mockClientCall(1, blockNumber, "eth_blockNumber").
			mockClientCall(0, code1, "eth_getCode", address, blockNumber).
			mockClientCall(1, code1, "eth_getCode", address, blockNumber).
			expectedResult(code1).
			test()
	})
	t.Run("earliest-block", func(t *testing.T) {
		prepareHandlerTest(t, 2, "eth_getBalance", address, newBlockID("earliest")).
			setOptions(WithRequirements(2, 10)).
			expectedError("").
			test()
	})
}

func Test_RPC_GetStorageAt(t *testing.T) {
	address := newAddress("0xc94770007dda54cF92009BFF0dE90c06F603a09f")
	position := newNumber("0x0")
	blockNumber := newBlockID("0x10")
	storageHash1 := newHash("0x0000000000000000000000000000000000000000000000000000000000000100")
	storageHash2 := newHash("0x0000000000000000000000000000000000000000000000000000000000000200")
	t.Run("simple", func(t *testing.T) {
		prepareHandlerTest(t, 3, "eth_getStorageAt", address, position, blockNumber).
			setOptions(WithRequirements(2, 10)).
			mockClientCall(0, storageHash1, "eth_getStorageAt", address, position, blockNumber).
			mockClientCall(1, storageHash1, "eth_getStorageAt", address, position, blockNumber).
			mockClientCall(2, storageHash1, "eth_getStorageAt", address, position, blockNumber).
			expectedResult(storageHash1).
			test()
	})
	t.Run("one-failed", func(t *testing.T) {
		prepareHandlerTest(t, 3, "eth_getStorageAt", address, position, blockNumber).
			setOptions(WithRequirements(2, 10)).
			mockClientCall(0, storageHash1, "eth_getStorageAt", address, position, blockNumber).
			mockClientCall(1, storageHash1, "eth_getStorageAt", address, position, blockNumber).
			mockClientCall(2, errors.New("error#1"), "eth_getStorageAt", address, position, blockNumber).
			expectedResult(storageHash1).
			test()
	})
	t.Run("two-failed", func(t *testing.T) {
		prepareHandlerTest(t, 3, "eth_getStorageAt", address, position, blockNumber).
			setOptions(WithRequirements(2, 10)).
			mockClientCall(0, storageHash1, "eth_getStorageAt", address, position, blockNumber).
			mockClientCall(1, errors.New("error#1"), "eth_getStorageAt", address, position, blockNumber).
			mockClientCall(2, errors.New("error#2"), "eth_getStorageAt", address, position, blockNumber).
			expectedError("error#1").
			expectedError("error#2").
			test()
	})
	t.Run("different-responses", func(t *testing.T) {
		prepareHandlerTest(t, 2, "eth_getStorageAt", address, position, blockNumber).
			setOptions(WithRequirements(2, 10)).
			mockClientCall(0, storageHash1, "eth_getStorageAt", address, position, blockNumber).
			mockClientCall(1, storageHash2, "eth_getStorageAt", address, position, blockNumber).
			expectedError("").
			test()
	})
	t.Run("latest-block", func(t *testing.T) {
		prepareHandlerTest(t, 2, "eth_getStorageAt", address, position, newBlockID("latest")).
			setOptions(WithRequirements(2, 10)).
			mockClientCall(0, blockNumber, "eth_blockNumber").
			mockClientCall(1, blockNumber, "eth_blockNumber").
			mockClientCall(0, storageHash1, "eth_getStorageAt", address, position, blockNumber).
			mockClientCall(1, storageHash1, "eth_getStorageAt", address, position, blockNumber).
			expectedResult(storageHash1).
			test()
	})
	t.Run("pending-block", func(t *testing.T) {
		prepareHandlerTest(t, 2, "eth_getStorageAt", address, position, newBlockID("pending")).
			setOptions(WithRequirements(2, 10)).
			mockClientCall(0, blockNumber, "eth_blockNumber").
			mockClientCall(1, blockNumber, "eth_blockNumber").
			mockClientCall(0, storageHash1, "eth_getStorageAt", address, position, blockNumber).
			mockClientCall(1, storageHash1, "eth_getStorageAt", address, position, blockNumber).
			expectedResult(storageHash1).
			test()
	})
	t.Run("earliest-block", func(t *testing.T) {
		prepareHandlerTest(t, 1, "eth_getBalance", address, position, newBlockID("earliest")).
			setOptions(WithRequirements(1, 10)).
			expectedError("").
			test()
	})
}

func Test_RPC_Accounts(t *testing.T) {
	t.Run("simple", func(t *testing.T) {
		prepareHandlerTest(t, 3, "eth_accounts").
			setOptions(WithRequirements(1, 10)).
			expectedError("the method eth_accounts does not exist").
			test()
	})
}

func Test_RPC_GetProof(t *testing.T) {
	t.Run("simple", func(t *testing.T) {
		prepareHandlerTest(t, 3, "eth_getProof").
			setOptions(WithRequirements(1, 10)).
			expectedError("the method eth_getProof does not exist").
			test()
	})
}

func Test_RPC_Call(t *testing.T) {
	call := newJSON(`
		{
			"from": "0xb60e8dd61c5d32be8058bb8eb970870f07233155",
			"to": "0xd46e8dd67c5d32be8058bb8eb970870f07244567",
			"gas": "0x76c0",
			"gasPrice": "0x9184e72a000",
			"value": "0x9184e72a",
			"data": "0xd46e8dd67c5d32be8d46e8dd67c5d32be8058bb8eb970870f072445675058bb8eb970870f072445675"
		}
	`)
	blockNumber := newBlockID("0x10")
	callRes1 := newBytes("0x01")
	callRes2 := newBytes("0x02")
	t.Run("simple", func(t *testing.T) {
		prepareHandlerTest(t, 3, "eth_call", call, blockNumber).
			setOptions(WithRequirements(2, 10)).
			mockClientCall(0, callRes1, "eth_call", call, blockNumber).
			mockClientCall(1, callRes1, "eth_call", call, blockNumber).
			mockClientCall(2, callRes1, "eth_call", call, blockNumber).
			expectedResult(callRes1).
			test()
	})
	t.Run("one-failed", func(t *testing.T) {
		prepareHandlerTest(t, 3, "eth_call", call, blockNumber).
			setOptions(WithRequirements(2, 10)).
			mockClientCall(0, callRes1, "eth_call", call, blockNumber).
			mockClientCall(1, callRes1, "eth_call", call, blockNumber).
			mockClientCall(2, errors.New("error#1"), "eth_call", call, blockNumber).
			expectedResult(callRes1).
			test()
	})
	t.Run("two-failed", func(t *testing.T) {
		prepareHandlerTest(t, 3, "eth_call", call, blockNumber).
			setOptions(WithRequirements(2, 10)).
			mockClientCall(0, callRes1, "eth_call", call, blockNumber).
			mockClientCall(1, errors.New("error#1"), "eth_call", call, blockNumber).
			mockClientCall(2, errors.New("error#2"), "eth_call", call, blockNumber).
			expectedError("error#1").
			expectedError("error#2").
			test()
	})
	t.Run("different-responses", func(t *testing.T) {
		prepareHandlerTest(t, 2, "eth_call", call, blockNumber).
			setOptions(WithRequirements(2, 10)).
			mockClientCall(0, callRes1, "eth_call", call, blockNumber).
			mockClientCall(1, callRes2, "eth_call", call, blockNumber).
			expectedError("").
			test()
	})
	t.Run("latest-block", func(t *testing.T) {
		prepareHandlerTest(t, 2, "eth_call", call, newBlockID("latest")).
			setOptions(WithRequirements(2, 10)).
			mockClientCall(0, blockNumber, "eth_blockNumber").
			mockClientCall(1, blockNumber, "eth_blockNumber").
			mockClientCall(0, callRes1, "eth_call", call, blockNumber).
			mockClientCall(1, callRes1, "eth_call", call, blockNumber).
			expectedResult(callRes1).
			test()
	})
	t.Run("pending-block", func(t *testing.T) {
		prepareHandlerTest(t, 2, "eth_call", call, newBlockID("pending")).
			setOptions(WithRequirements(2, 10)).
			mockClientCall(0, blockNumber, "eth_blockNumber").
			mockClientCall(1, blockNumber, "eth_blockNumber").
			mockClientCall(0, callRes1, "eth_call", call, blockNumber).
			mockClientCall(1, callRes1, "eth_call", call, blockNumber).
			expectedResult(callRes1).
			test()
	})
	t.Run("earliest-block", func(t *testing.T) {
		prepareHandlerTest(t, 2, "eth_call", call, newBlockID("earliest")).
			setOptions(WithRequirements(2, 10)).
			expectedError("").
			test()
	})
}

func Test_RPC_GetLogs(t *testing.T) {
	address := newAddresses("0xc94770007dda54cF92009BFF0dE90c06F603a09f")
	blockHash := newHash("0xab059a62e22e230fe0f56d8555340a29b2e9532360368f810595453f6fdd213b")
	topics := []hashesType{newHashes("0xc0f4906fea23cf6f3cce98cb44e8e1449e455b28d684dfa9ff65426495584de6")}
	filter := &logFilterType{
		Address:   &address,
		FromBlock: newBlockID("0x5"),
		ToBlock:   newBlockID("0x5"),
		Topics:    topics,
		BlockHash: &blockHash,
	}
	filterLatest := &logFilterType{
		Address:   &address,
		FromBlock: newBlockID("latest"),
		ToBlock:   newBlockID("latest"),
		Topics:    topics,
		BlockHash: &blockHash,
	}
	filterPending := &logFilterType{
		Address:   &address,
		FromBlock: newBlockID("pending"),
		ToBlock:   newBlockID("pending"),
		Topics:    topics,
		BlockHash: &blockHash,
	}
	filterEarliest := &logFilterType{
		Address:   &address,
		FromBlock: newBlockID("earliest"),
		ToBlock:   newBlockID("earliest"),
		Topics:    topics,
		BlockHash: &blockHash,
	}
	t.Run("simple", func(t *testing.T) {
		prepareHandlerTest(t, 3, "eth_getLogs", filter).
			setOptions(WithRequirements(2, 10)).
			mockClientCall(0, getLogs1Resp, "eth_getLogs", filter).
			mockClientCall(1, getLogs1Resp, "eth_getLogs", filter).
			mockClientCall(2, getLogs1Resp, "eth_getLogs", filter).
			expectedResult(getLogs1Resp).
			test()
	})
	t.Run("one-failed", func(t *testing.T) {
		prepareHandlerTest(t, 3, "eth_getLogs", filter).
			setOptions(WithRequirements(2, 10)).
			mockClientCall(0, getLogs1Resp, "eth_getLogs", filter).
			mockClientCall(1, getLogs1Resp, "eth_getLogs", filter).
			mockClientCall(2, errors.New("error#1"), "eth_getLogs", filter).
			expectedResult(getLogs1Resp).
			test()
	})
	t.Run("two-failed", func(t *testing.T) {
		prepareHandlerTest(t, 3, "eth_getLogs", filter).
			setOptions(WithRequirements(2, 10)).
			mockClientCall(0, getLogs1Resp, "eth_getLogs", filter).
			mockClientCall(1, errors.New("error#1"), "eth_getLogs", filter).
			mockClientCall(2, errors.New("error#2"), "eth_getLogs", filter).
			expectedError("error#1").
			expectedError("error#2").
			test()
	})
	t.Run("different-responses", func(t *testing.T) {
		prepareHandlerTest(t, 2, "eth_getLogs", filter).
			setOptions(WithRequirements(2, 10)).
			mockClientCall(0, getLogs1Resp, "eth_getLogs", filter).
			mockClientCall(1, getLogs2Resp, "eth_getLogs", filter).
			expectedError("").
			test()
	})
	t.Run("latest-block", func(t *testing.T) {
		f := *filterLatest
		f.FromBlock = newBlockID("0x5")
		f.ToBlock = newBlockID("0x6")
		prepareHandlerTest(t, 2, "eth_getLogs", filterLatest).
			setOptions(WithRequirements(2, 10)).
			mockClientCall(0, newBlockID("0x5"), "eth_blockNumber").
			mockClientCall(0, newBlockID("0x6"), "eth_blockNumber").
			mockClientCall(1, newBlockID("0x5"), "eth_blockNumber").
			mockClientCall(1, newBlockID("0x6"), "eth_blockNumber").
			mockClientCall(0, getLogs1Resp, "eth_getLogs", f).
			mockClientCall(1, getLogs1Resp, "eth_getLogs", f).
			expectedResult(getLogs1Resp).
			test()
	})
	t.Run("pending-block", func(t *testing.T) {
		f := *filterLatest
		f.FromBlock = newBlockID("0x5")
		f.ToBlock = newBlockID("0x6")
		prepareHandlerTest(t, 2, "eth_getLogs", filterPending).
			setOptions(WithRequirements(2, 10)).
			mockClientCall(0, newBlockID("0x5"), "eth_blockNumber").
			mockClientCall(0, newBlockID("0x6"), "eth_blockNumber").
			mockClientCall(1, newBlockID("0x5"), "eth_blockNumber").
			mockClientCall(1, newBlockID("0x6"), "eth_blockNumber").
			mockClientCall(0, getLogs1Resp, "eth_getLogs", f).
			mockClientCall(1, getLogs1Resp, "eth_getLogs", f).
			expectedResult(getLogs1Resp).
			test()
	})
	t.Run("earliest-block", func(t *testing.T) {
		prepareHandlerTest(t, 2, "eth_getLogs", filterEarliest).
			setOptions(WithRequirements(2, 10)).
			expectedError("").
			test()
	})
}

func Test_RPC_ProtocolVersion(t *testing.T) {
	t.Run("simple", func(t *testing.T) {
		prepareHandlerTest(t, 3, "eth_protocolVersion").
			setOptions(WithRequirements(2, 10)).
			expectedError("the method eth_protocolVersion does not exist").
			test()
	})
}

func Test_RPC_GasPrice(t *testing.T) {
	t.Run("four-responses", func(t *testing.T) {
		prepareHandlerTest(t, 4, "eth_gasPrice").
			setOptions(WithRequirements(3, 10)).
			mockClientCall(0, `0x1`, "eth_gasPrice").
			mockClientCall(1, `0x5`, "eth_gasPrice").
			mockClientCall(2, `0x7`, "eth_gasPrice").
			mockClientCall(3, `0x8`, "eth_gasPrice").
			expectedResult(`0x6`).
			test()
	})
	t.Run("three-responses", func(t *testing.T) {
		prepareHandlerTest(t, 3, "eth_gasPrice").
			setOptions(WithRequirements(2, 10)).
			mockClientCall(0, `0x1`, "eth_gasPrice").
			mockClientCall(1, `0x5`, "eth_gasPrice").
			mockClientCall(2, `0x6`, "eth_gasPrice").
			expectedResult(`0x5`).
			test()
	})
	t.Run("two-responses", func(t *testing.T) {
		prepareHandlerTest(t, 2, "eth_gasPrice").
			setOptions(WithRequirements(2, 10)).
			mockClientCall(0, `0x4`, "eth_gasPrice").
			mockClientCall(1, `0x2`, "eth_gasPrice").
			expectedResult(`0x2`).
			test()
	})
	t.Run("one-failed", func(t *testing.T) {
		prepareHandlerTest(t, 3, "eth_gasPrice").
			setOptions(WithRequirements(2, 10)).
			mockClientCall(0, `0x2`, "eth_gasPrice").
			mockClientCall(1, `0x4`, "eth_gasPrice").
			mockClientCall(2, errors.New("error#1"), "eth_gasPrice").
			expectedResult(`0x2`).
			test()
	})
	t.Run("two-failed", func(t *testing.T) {
		prepareHandlerTest(t, 3, "eth_gasPrice").
			setOptions(WithRequirements(2, 10)).
			mockClientCall(0, `0x3`, "eth_gasPrice").
			mockClientCall(1, errors.New("error#1"), "eth_gasPrice").
			mockClientCall(2, errors.New("error#2"), "eth_gasPrice").
			expectedError("error#1").
			expectedError("error#2").
			test()
	})
}

func Test_RPC_EstimateGas(t *testing.T) {
	call := newJSON(`
		{
			"from": "0xb60e8dd61c5d32be8058bb8eb970870f07233155",
			"to": "0xd46e8dd67c5d32be8058bb8eb970870f07244567",
			"gas": "0x76c0",
			"gasPrice": "0x9184e72a000",
			"value": "0x9184e72a",
			"data": "0xd46e8dd67c5d32be8d46e8dd67c5d32be8058bb8eb970870f072445675058bb8eb970870f072445675"
		}
	`)
	blockNumber := newBlockID("0x10")
	t.Run("four-responses", func(t *testing.T) {
		prepareHandlerTest(t, 4, "eth_estimateGas", call, blockNumber).
			setOptions(WithRequirements(3, 10)).
			mockClientCall(0, `0x1`, "eth_estimateGas", call, blockNumber).
			mockClientCall(1, `0x5`, "eth_estimateGas", call, blockNumber).
			mockClientCall(2, `0x7`, "eth_estimateGas", call, blockNumber).
			mockClientCall(3, `0x8`, "eth_estimateGas", call, blockNumber).
			expectedResult(`0x6`).
			test()
	})
	t.Run("three-responses", func(t *testing.T) {
		prepareHandlerTest(t, 3, "eth_estimateGas", call, blockNumber).
			setOptions(WithRequirements(2, 10)).
			mockClientCall(0, `0x1`, "eth_estimateGas", call, blockNumber).
			mockClientCall(1, `0x5`, "eth_estimateGas", call, blockNumber).
			mockClientCall(2, `0x6`, "eth_estimateGas", call, blockNumber).
			expectedResult(`0x5`).
			test()
	})
	t.Run("two-responses", func(t *testing.T) {
		prepareHandlerTest(t, 2, "eth_estimateGas", call, blockNumber).
			setOptions(WithRequirements(2, 10)).
			mockClientCall(0, `0x2`, "eth_estimateGas", call, blockNumber).
			mockClientCall(1, `0x4`, "eth_estimateGas", call, blockNumber).
			expectedResult(`0x2`).
			test()
	})
	t.Run("one-failed", func(t *testing.T) {
		prepareHandlerTest(t, 3, "eth_estimateGas", call, blockNumber).
			setOptions(WithRequirements(2, 10)).
			mockClientCall(0, `0x2`, "eth_estimateGas", call, blockNumber).
			mockClientCall(1, `0x4`, "eth_estimateGas", call, blockNumber).
			mockClientCall(2, errors.New("error#1"), "eth_estimateGas", call, blockNumber).
			expectedResult(`0x2`).
			test()
	})
	t.Run("two-failed", func(t *testing.T) {
		prepareHandlerTest(t, 3, "eth_estimateGas", call, blockNumber).
			setOptions(WithRequirements(2, 10)).
			mockClientCall(0, `0x3`, "eth_estimateGas", call, blockNumber).
			mockClientCall(1, errors.New("error#1"), "eth_estimateGas", call, blockNumber).
			mockClientCall(2, errors.New("error#2"), "eth_estimateGas", call, blockNumber).
			expectedError("error#1").
			expectedError("error#2").
			test()
	})
	t.Run("latest-block", func(t *testing.T) {
		prepareHandlerTest(t, 2, "eth_estimateGas", call, newBlockID("latest")).
			setOptions(WithRequirements(2, 10)).
			mockClientCall(0, blockNumber, "eth_blockNumber").
			mockClientCall(1, blockNumber, "eth_blockNumber").
			mockClientCall(0, `0x4`, "eth_estimateGas", call, blockNumber).
			mockClientCall(1, `0x4`, "eth_estimateGas", call, blockNumber).
			expectedResult(`0x4`).
			test()
	})
	t.Run("pending-block", func(t *testing.T) {
		prepareHandlerTest(t, 2, "eth_estimateGas", call, newBlockID("pending")).
			setOptions(WithRequirements(2, 10)).
			mockClientCall(0, blockNumber, "eth_blockNumber").
			mockClientCall(1, blockNumber, "eth_blockNumber").
			mockClientCall(0, `0x4`, "eth_estimateGas", call, blockNumber).
			mockClientCall(1, `0x4`, "eth_estimateGas", call, blockNumber).
			expectedResult(`0x4`).
			test()
	})
	t.Run("earliest-block", func(t *testing.T) {
		prepareHandlerTest(t, 2, "eth_estimateGas", call, newBlockID("earliest")).
			setOptions(WithRequirements(2, 10)).
			expectedError("").
			test()
	})
}

func Test_RPC_FeeHistory(t *testing.T) {
	blockCount := newNumber("0x5")
	newestBlock := newBlockID("0x10")
	percentiles := newJSON("[25, 75]")
	t.Run("simple", func(t *testing.T) {
		prepareHandlerTest(t, 3, "eth_feeHistory", blockCount, newestBlock, percentiles).
			setOptions(WithRequirements(2, 10)).
			mockClientCall(0, feeHistory1Resp, "eth_feeHistory", blockCount, newestBlock, percentiles).
			mockClientCall(1, feeHistory1Resp, "eth_feeHistory", blockCount, newestBlock, percentiles).
			mockClientCall(2, feeHistory1Resp, "eth_feeHistory", blockCount, newestBlock, percentiles).
			expectedResult(feeHistory1Resp).
			test()
	})
	t.Run("one-failed", func(t *testing.T) {
		prepareHandlerTest(t, 3, "eth_feeHistory", blockCount, newestBlock, percentiles).
			setOptions(WithRequirements(2, 10)).
			mockClientCall(0, feeHistory1Resp, "eth_feeHistory", blockCount, newestBlock, percentiles).
			mockClientCall(1, feeHistory1Resp, "eth_feeHistory", blockCount, newestBlock, percentiles).
			mockClientCall(2, errors.New("error#1"), "eth_feeHistory", blockCount, newestBlock, percentiles).
			expectedResult(feeHistory1Resp).
			test()
	})
	t.Run("two-failed", func(t *testing.T) {
		prepareHandlerTest(t, 3, "eth_feeHistory", blockCount, newestBlock, percentiles).
			setOptions(WithRequirements(2, 10)).
			mockClientCall(0, feeHistory1Resp, "eth_feeHistory", blockCount, newestBlock, percentiles).
			mockClientCall(1, errors.New("error#1"), "eth_feeHistory", blockCount, newestBlock, percentiles).
			mockClientCall(2, errors.New("error#2"), "eth_feeHistory", blockCount, newestBlock, percentiles).
			expectedError("error#1").
			expectedError("error#2").
			test()
	})
	t.Run("different-responses", func(t *testing.T) {
		prepareHandlerTest(t, 2, "eth_feeHistory", blockCount, newestBlock, percentiles).
			setOptions(WithRequirements(2, 10)).
			mockClientCall(0, feeHistory1Resp, "eth_feeHistory", blockCount, newestBlock, percentiles).
			mockClientCall(1, feeHistory2Resp, "eth_feeHistory", blockCount, newestBlock, percentiles).
			expectedError("").
			test()
	})
	t.Run("latest-block", func(t *testing.T) {
		prepareHandlerTest(t, 2, "eth_feeHistory", blockCount, newBlockID("latest"), percentiles).
			setOptions(WithRequirements(2, 10)).
			mockClientCall(0, newestBlock, "eth_blockNumber").
			mockClientCall(1, newestBlock, "eth_blockNumber").
			mockClientCall(0, feeHistory1Resp, "eth_feeHistory", blockCount, newestBlock, percentiles).
			mockClientCall(1, feeHistory1Resp, "eth_feeHistory", blockCount, newestBlock, percentiles).
			expectedResult(feeHistory1Resp).
			test()
	})
	t.Run("pending-block", func(t *testing.T) {
		prepareHandlerTest(t, 2, "eth_feeHistory", blockCount, newBlockID("pending"), percentiles).
			setOptions(WithRequirements(2, 10)).
			mockClientCall(0, newestBlock, "eth_blockNumber").
			mockClientCall(1, newestBlock, "eth_blockNumber").
			mockClientCall(0, feeHistory1Resp, "eth_feeHistory", blockCount, newestBlock, percentiles).
			mockClientCall(1, feeHistory1Resp, "eth_feeHistory", blockCount, newestBlock, percentiles).
			expectedResult(feeHistory1Resp).
			test()
	})
	t.Run("earliest-block", func(t *testing.T) {
		prepareHandlerTest(t, 2, "eth_getBalance", blockCount, newBlockID("earliest"), percentiles).
			setOptions(WithRequirements(2, 10)).
			expectedError("").
			test()
	})
}

func Test_RPC_MaxPriorityFeePerGas(t *testing.T) {
	t.Run("four-responses", func(t *testing.T) {
		prepareHandlerTest(t, 4, "eth_maxPriorityFeePerGas").
			setOptions(WithRequirements(3, 10)).
			mockClientCall(0, `0x1`, "eth_maxPriorityFeePerGas").
			mockClientCall(1, `0x5`, "eth_maxPriorityFeePerGas").
			mockClientCall(2, `0x7`, "eth_maxPriorityFeePerGas").
			mockClientCall(3, `0x8`, "eth_maxPriorityFeePerGas").
			expectedResult(`0x6`).
			test()
	})
	t.Run("three-responses", func(t *testing.T) {
		prepareHandlerTest(t, 3, "eth_maxPriorityFeePerGas").
			setOptions(WithRequirements(2, 10)).
			mockClientCall(0, `0x1`, "eth_maxPriorityFeePerGas").
			mockClientCall(1, `0x5`, "eth_maxPriorityFeePerGas").
			mockClientCall(2, `0x6`, "eth_maxPriorityFeePerGas").
			expectedResult(`0x5`).
			test()
	})
	t.Run("two-responses", func(t *testing.T) {
		prepareHandlerTest(t, 2, "eth_maxPriorityFeePerGas").
			setOptions(WithRequirements(2, 10)).
			mockClientCall(0, `0x2`, "eth_maxPriorityFeePerGas").
			mockClientCall(1, `0x4`, "eth_maxPriorityFeePerGas").
			expectedResult(`0x2`).
			test()
	})
	t.Run("one-failed", func(t *testing.T) {
		prepareHandlerTest(t, 3, "eth_maxPriorityFeePerGas").
			setOptions(WithRequirements(2, 10)).
			mockClientCall(0, `0x2`, "eth_maxPriorityFeePerGas").
			mockClientCall(1, `0x4`, "eth_maxPriorityFeePerGas").
			mockClientCall(2, errors.New("error#1"), "eth_maxPriorityFeePerGas").
			expectedResult(`0x2`).
			test()
	})
	t.Run("two-failed", func(t *testing.T) {
		prepareHandlerTest(t, 3, "eth_maxPriorityFeePerGas").
			setOptions(WithRequirements(2, 10)).
			mockClientCall(0, `0x3`, "eth_maxPriorityFeePerGas").
			mockClientCall(1, errors.New("error#1"), "eth_maxPriorityFeePerGas").
			mockClientCall(2, errors.New("error#2"), "eth_maxPriorityFeePerGas").
			expectedError("error#1").
			expectedError("error#2").
			test()
	})
}

func Test_RPC_ChainId(t *testing.T) {
	t.Run("simple", func(t *testing.T) {
		prepareHandlerTest(t, 3, "eth_chainId").
			setOptions(WithRequirements(2, 10)).
			mockClientCall(0, `0x1`, "eth_chainId").
			mockClientCall(1, `0x1`, "eth_chainId").
			mockClientCall(2, `0x1`, "eth_chainId").
			expectedResult(`0x1`).
			test()
	})
	t.Run("one-failed", func(t *testing.T) {
		prepareHandlerTest(t, 3, "eth_chainId").
			setOptions(WithRequirements(2, 10)).
			mockClientCall(0, `0x1`, "eth_chainId").
			mockClientCall(1, `0x1`, "eth_chainId").
			mockClientCall(2, errors.New("error#1"), "eth_chainId").
			expectedResult(`0x1`).
			test()
	})
	t.Run("two-failed", func(t *testing.T) {
		prepareHandlerTest(t, 3, "eth_chainId").
			setOptions(WithRequirements(2, 10)).
			mockClientCall(0, `0x1`, "eth_chainId").
			mockClientCall(1, errors.New("error#1"), "eth_chainId").
			mockClientCall(2, errors.New("error#2"), "eth_chainId").
			expectedError("error#1").
			expectedError("error#2").
			test()
	})
	t.Run("different-responses", func(t *testing.T) {
		prepareHandlerTest(t, 2, "eth_chainId").
			setOptions(WithRequirements(2, 10)).
			mockClientCall(0, `0x1`, "eth_chainId").
			mockClientCall(1, `0x2`, "eth_chainId").
			expectedError("").
			test()
	})
}

func Test_RPC_Version(t *testing.T) {
	t.Run("simple", func(t *testing.T) {
		prepareHandlerTest(t, 3, "net_version").
			setOptions(WithRequirements(2, 10)).
			mockClientCall(0, 1, "net_version").
			mockClientCall(1, 1, "net_version").
			mockClientCall(2, 1, "net_version").
			expectedResult(1).
			test()
	})
	t.Run("one-failed", func(t *testing.T) {
		prepareHandlerTest(t, 3, "net_version").
			setOptions(WithRequirements(2, 10)).
			mockClientCall(0, 1, "net_version").
			mockClientCall(1, 1, "net_version").
			mockClientCall(2, errors.New("error#1"), "net_version").
			expectedResult(1).
			test()
	})
	t.Run("two-failed", func(t *testing.T) {
		prepareHandlerTest(t, 3, "net_version").
			setOptions(WithRequirements(2, 10)).
			mockClientCall(0, 1, "net_version").
			mockClientCall(1, errors.New("error#1"), "net_version").
			mockClientCall(2, errors.New("error#2"), "net_version").
			expectedError("error#1").
			expectedError("error#2").
			test()
	})
	t.Run("different-responses", func(t *testing.T) {
		prepareHandlerTest(t, 2, "net_version").
			setOptions(WithRequirements(2, 10)).
			mockClientCall(0, `0x1`, "net_version").
			mockClientCall(1, `0x2`, "net_version").
			expectedError("").
			test()
	})
}

func Test_RPC_Timeout(t *testing.T) {
	t.Run("total-timeout", func(t *testing.T) {
		prepareHandlerTest(t, 3, "eth_blockNumber").
			setOptions(WithRequirements(2, 10), WithTotalTimeout(100*time.Millisecond), WithGracefulTimeout(100*time.Millisecond)).
			mockClientSlowCall(time.Millisecond*50, 0, 1, "eth_blockNumber").
			mockClientSlowCall(time.Millisecond*150, 1, 1, "eth_blockNumber").
			mockClientSlowCall(time.Millisecond*150, 2, 1, "eth_blockNumber").
			expectedError("context cancelled").
			test()
	})
	t.Run("graceful-timeout", func(t *testing.T) {
		prepareHandlerTest(t, 3, "eth_blockNumber").
			setOptions(WithRequirements(2, 10), WithTotalTimeout(100*time.Millisecond), WithGracefulTimeout(50*time.Millisecond)).
			mockClientCall(0, 1, "eth_blockNumber").
			mockClientCall(1, 1, "eth_blockNumber").
			mockClientSlowCall(time.Millisecond*150, 2, 1, "eth_blockNumber").
			expectedResult(`0x1`).
			test()
	})
}
