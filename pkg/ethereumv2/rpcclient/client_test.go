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

package rpcclient

import (
	"bytes"
	"context"
	"io"
	"net/http"
	"testing"

	"github.com/ethereum/go-ethereum/rpc"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/chronicleprotocol/oracle-suite/pkg/ethereumv2/types"
	"github.com/chronicleprotocol/oracle-suite/pkg/util/errutil"
)

const (
	blockNumberResponse           = `{"jsonrpc":"2.0","id":1,"result":"0x1"}`
	blockByNumberTxHashesResponse = `{
		"jsonrpc": "2.0",
		"id": 1,
		"result": {
			"number": "0x1",
			"hash": "0x2",
			"parentHash": "0x3",
			"nonce": "0x4",
			"sha3Uncles": "0x5",
			"logsBloom": "0x6",
			"transactionsRoot": "0x7",
			"stateRoot": "0x8",
			"receiptsRoot": "0x9",
			"miner": "0xa",
			"difficulty": "0xb",
			"totalDifficulty": "0xc",
			"extraData": "0xd",
			"size": "0xe",
			"gasLimit": "0xf",
			"gasUsed": "0x10",
			"timestamp": "0x11",
			"transactions": [
				"0x12",
				"0x13"
			],
			"uncles": [
				"0x14",
				"0x15"
			]
		}
	}`
	blockByNumberObjectsResponse = `{
		"jsonrpc": "2.0",
		"id": 1,
		"result": {
			"number": "0x1",
			"hash": "0x2",
			"parentHash": "0x3",
			"nonce": "0x4",
			"sha3Uncles": "0x5",
			"logsBloom": "0x6",
			"transactionsRoot": "0x7",
			"stateRoot": "0x8",
			"receiptsRoot": "0x9",
			"miner": "0xa",
			"difficulty": "0xb",
			"totalDifficulty": "0xc",
			"extraData": "0xd",
			"size": "0xe",
			"gasLimit": "0xf",
			"gasUsed": "0x10",
			"timestamp": "0x11",
			"transactions": [
				{
					"hash": "0x12",
					"nonce": "0x13",
					"blockHash": "0x14",
					"blockNumber": "0x15",
					"transactionIndex": "0x16",
					"from": "0x17",
					"to": "0x18",
					"value": "0x19",
					"gasPrice": "0x1a",
					"gas": "0x1b",
					"input": "0x1c"
				},
				{
					"hash": "0x1d",
					"nonce": "0x1e",
					"blockHash": "0x1f",
					"blockNumber": "0x20",
					"transactionIndex": "0x21",
					"from": "0x22",
					"to": "0x23",
					"value": "0x24",
					"gasPrice": "0x25",
					"gas": "0x26",
					"input": "0x27"
				}
			],
			"uncles": [
				"0x28",
				"0x29"
			]
		}
	}`
	getTransactionCountResponse = `{"jsonrpc":"2.0","id":1,"result":"0x1"}`
	sendRawTransactionResponse  = `{"jsonrpc":"2.0","id":1,"result":"0x1"}`
	getStorageAtResponse        = `{"jsonrpc":"2.0","id":1,"result":"0x1"}`
	filterLogsResponse          = `{
	   "jsonrpc":"2.0",
	   "id":1,
	   "result":[
		  {
			 "address":"0x00112233445566778899aabbccddeeff00112233",
			 "topics":[
				"0x00112233445566778899aabbccddeeff00112233"
			 ],
			 "data":"0x00112233445566778899aabbccddeeff00112233",
			 "blockNumber":"0x1",
			 "transactionHash":"0x00112233445566778899aabbccddeeff00112233",
			 "transactionIndex":"0x1",
			 "blockHash":"0x00112233445566778899aabbccddeeff00112233",
			 "logIndex":"0x1",
			 "removed":false
		  }
	   ]
	}`
)

type RoundTripFunc func(req *http.Request) (*http.Response, error)

func (f RoundTripFunc) RoundTrip(req *http.Request) (*http.Response, error) {
	return f(req)
}

type testableClient struct {
	*Client
	req *http.Request
	res *http.Response
}

func newTestableClient() *testableClient {
	t := &testableClient{}
	t.Client = New(errutil.Must(rpc.DialHTTPWithClient("http://localhost/", &http.Client{
		Transport: RoundTripFunc(func(req *http.Request) (*http.Response, error) {
			t.req = req
			return t.res, nil
		}),
	})))
	return t
}

func TestClient_BlockNumber(t *testing.T) {
	cli := newTestableClient()
	cli.res = &http.Response{
		StatusCode: http.StatusOK,
		Body:       io.NopCloser(bytes.NewReader([]byte(blockNumberResponse))),
	}
	blockNumber, err := cli.BlockNumber(context.Background())

	require.NoError(t, err)
	assert.Equal(t, uint64(1), blockNumber)
	assert.Equal(t, http.MethodPost, cli.req.Method)
	assert.Equal(t, "/", cli.req.URL.Path)
	assert.Equal(t, "application/json", cli.req.Header.Get("Content-Type"))
	assert.JSONEq(t, `{"jsonrpc":"2.0","method":"eth_blockNumber","id":1}`, readAll(t, cli.req.Body))
}

func TestClient_BlockByNumber(t *testing.T) {
	cli := newTestableClient()
	cli.res = &http.Response{
		StatusCode: http.StatusOK,
		Body:       io.NopCloser(bytes.NewReader([]byte(blockByNumberTxHashesResponse))),
	}
	block, err := cli.BlockByNumber(
		context.Background(),
		types.StringToBlockNumber("latest"),
	)

	require.NoError(t, err)
	assert.Equal(t, types.HexToNumber("0x1"), block.Number)
	assert.Equal(t, types.HexToHash("0x2"), block.Hash)
	assert.Equal(t, types.HexToHash("0x3"), block.ParentHash)
	assert.Equal(t, types.HexToNonce("0x4"), block.Nonce)
	assert.Equal(t, types.HexToHash("0x5"), block.Sha3Uncles)
	assert.Equal(t, types.BytesToBloom([]byte{0x6}), block.LogsBloom)
	assert.Equal(t, types.HexToHash("0x7"), block.TransactionsRoot)
	assert.Equal(t, types.HexToHash("0x8"), block.StateRoot)
	assert.Equal(t, types.HexToHash("0x9"), block.ReceiptsRoot)
	assert.Equal(t, types.HexToAddress("0xa"), block.Miner)
	assert.Equal(t, types.HexToNumber("0xb"), block.Difficulty)
	assert.Equal(t, types.HexToNumber("0xc"), block.TotalDifficulty)
	assert.Equal(t, types.HexToBytes("0xd"), block.ExtraData)
	assert.Equal(t, types.HexToNumber("0xe"), block.Size)
	assert.Equal(t, types.HexToNumber("0xf"), block.GasLimit)
	assert.Equal(t, types.HexToNumber("0x10"), block.GasUsed)
	assert.Equal(t, types.HexToNumber("0x11"), block.Timestamp)
	assert.Equal(t, 2, len(block.Transactions))
	assert.IsType(t, types.Hash{}, block.Transactions[0])
	assert.Equal(t, 2, len(block.Uncles))
	assert.Equal(t, http.MethodPost, cli.req.Method)
	assert.Equal(t, "/", cli.req.URL.Path)
	assert.Equal(t, "application/json", cli.req.Header.Get("Content-Type"))
	assert.JSONEq(t, `{"jsonrpc":"2.0","method":"eth_getBlockByNumber","params":["latest",false],"id":1}`, readAll(t, cli.req.Body))
}

func TestClient_FullBlockByNumber(t *testing.T) {
	cli := newTestableClient()
	cli.res = &http.Response{
		StatusCode: http.StatusOK,
		Body:       io.NopCloser(bytes.NewReader([]byte(blockByNumberObjectsResponse))),
	}
	block, err := cli.FullBlockByNumber(
		context.Background(),
		types.StringToBlockNumber("latest"),
	)

	require.NoError(t, err)
	assert.Equal(t, types.HexToNumber("0x1"), block.Number)
	assert.Equal(t, types.HexToHash("0x2"), block.Hash)
	assert.Equal(t, types.HexToHash("0x3"), block.ParentHash)
	assert.Equal(t, types.HexToNonce("0x4"), block.Nonce)
	assert.Equal(t, types.HexToHash("0x5"), block.Sha3Uncles)
	assert.Equal(t, types.BytesToBloom([]byte{0x6}), block.LogsBloom)
	assert.Equal(t, types.HexToHash("0x7"), block.TransactionsRoot)
	assert.Equal(t, types.HexToHash("0x8"), block.StateRoot)
	assert.Equal(t, types.HexToHash("0x9"), block.ReceiptsRoot)
	assert.Equal(t, types.HexToAddress("0xa"), block.Miner)
	assert.Equal(t, types.HexToNumber("0xb"), block.Difficulty)
	assert.Equal(t, types.HexToNumber("0xc"), block.TotalDifficulty)
	assert.Equal(t, types.HexToBytes("0xd"), block.ExtraData)
	assert.Equal(t, types.HexToNumber("0xe"), block.Size)
	assert.Equal(t, types.HexToNumber("0xf"), block.GasLimit)
	assert.Equal(t, types.HexToNumber("0x10"), block.GasUsed)
	assert.Equal(t, types.HexToNumber("0x11"), block.Timestamp)
	assert.Equal(t, 2, len(block.Transactions))
	assert.IsType(t, types.Transaction{}, block.Transactions[0])
	assert.Equal(t, 2, len(block.Uncles))
	assert.Equal(t, http.MethodPost, cli.req.Method)
	assert.Equal(t, "/", cli.req.URL.Path)
	assert.Equal(t, "application/json", cli.req.Header.Get("Content-Type"))
	assert.JSONEq(t, `{"jsonrpc":"2.0","method":"eth_getBlockByNumber","params":["latest",true],"id":1}`, readAll(t, cli.req.Body))
}

func TestClient_GetTransactionCount(t *testing.T) {
	cli := newTestableClient()
	cli.res = &http.Response{
		StatusCode: http.StatusOK,
		Body:       io.NopCloser(bytes.NewReader([]byte(getTransactionCountResponse))),
	}
	nonce, err := cli.GetTransactionCount(
		context.Background(),
		types.HexToAddress("0x00112233445566778899aabbccddeeff00112233"),
		types.StringToBlockNumber("latest"),
	)

	require.NoError(t, err)
	assert.Equal(t, uint64(1), nonce)
	assert.Equal(t, http.MethodPost, cli.req.Method)
	assert.Equal(t, "/", cli.req.URL.Path)
	assert.Equal(t, "application/json", cli.req.Header.Get("Content-Type"))
	assert.JSONEq(t, `{"jsonrpc":"2.0","method":"eth_getTransactionCount","params":["0x00112233445566778899aabbccddeeff00112233","latest"],"id":1}`, readAll(t, cli.req.Body))
}

func TestClient_SendRawTransaction(t *testing.T) {
	cli := newTestableClient()
	cli.res = &http.Response{
		StatusCode: http.StatusOK,
		Body:       io.NopCloser(bytes.NewReader([]byte(sendRawTransactionResponse))),
	}
	txHash, err := cli.SendRawTransaction(
		context.Background(),
		types.HexToBytes("0x00112233445566778899aabbccddeeff00112233"),
	)

	require.NoError(t, err)
	assert.Equal(t, types.HexToHash("0x1"), *txHash)
	assert.Equal(t, http.MethodPost, cli.req.Method)
	assert.Equal(t, "/", cli.req.URL.Path)
	assert.Equal(t, "application/json", cli.req.Header.Get("Content-Type"))
	assert.JSONEq(t, `{"jsonrpc":"2.0","method":"eth_sendRawTransaction","params":["0x00112233445566778899aabbccddeeff00112233"],"id":1}`, readAll(t, cli.req.Body))
}

func TestClient_GetStorageAt(t *testing.T) {
	cli := newTestableClient()
	cli.res = &http.Response{
		StatusCode: http.StatusOK,
		Body:       io.NopCloser(bytes.NewReader([]byte(getStorageAtResponse))),
	}
	storage, err := cli.GetStorageAt(
		context.Background(),
		types.HexToAddress("0x00112233445566778899aabbccddeeff00112233"),
		types.HexToHash("0x00000000000000000000000000112233445566778899aabbccddeeff00112233"),
		types.StringToBlockNumber("latest"),
	)

	require.NoError(t, err)
	assert.Equal(t, types.HexToHash("0x1"), *storage)
	assert.Equal(t, http.MethodPost, cli.req.Method)
	assert.Equal(t, "/", cli.req.URL.Path)
	assert.Equal(t, "application/json", cli.req.Header.Get("Content-Type"))
	assert.JSONEq(t, `{"jsonrpc":"2.0","method":"eth_getStorageAt","params":["0x00112233445566778899aabbccddeeff00112233","0x00000000000000000000000000112233445566778899aabbccddeeff00112233","latest"],"id":1}`, readAll(t, cli.req.Body))
}

func TestClient_FilterLogs(t *testing.T) {
	cli := newTestableClient()
	cli.res = &http.Response{
		StatusCode: http.StatusOK,
		Body:       io.NopCloser(bytes.NewReader([]byte(filterLogsResponse))),
	}
	logs, err := cli.FilterLogs(
		context.Background(),
		types.FilterLogsQuery{
			FromBlock: ptr(types.Uint64ToBlockNumber(10000)),
			ToBlock:   ptr(types.StringToBlockNumber("latest")),
			Address: types.Addresses{
				types.HexToAddress("0x00112233445566778899aabbccddeeff00112233"),
			},
			Topics: []types.Hashes{
				{
					types.HexToHash("0x00112233445566778899aabbccddeeff00112233"),
				},
			},
		},
	)

	require.NoError(t, err)
	assert.Equal(t, []types.Log{
		{
			Address:     types.HexToAddress("0x00112233445566778899aabbccddeeff00112233"),
			Topics:      []types.Hash{types.HexToHash("0x00112233445566778899aabbccddeeff00112233")},
			Data:        types.HexToBytes("0x00112233445566778899aabbccddeeff00112233"),
			BlockNumber: types.Uint64ToNumber(1),
			TxHash:      types.HexToHash("0x00112233445566778899aabbccddeeff00112233"),
			TxIndex:     types.Uint64ToNumber(1),
			BlockHash:   types.HexToHash("0x00112233445566778899aabbccddeeff00112233"),
			LogIndex:    types.Uint64ToNumber(1),
			Removed:     false,
		},
	}, logs)
	assert.Equal(t, http.MethodPost, cli.req.Method)
	assert.Equal(t, "/", cli.req.URL.Path)
	assert.Equal(t, "application/json", cli.req.Header.Get("Content-Type"))
	assert.JSONEq(t, `{"jsonrpc":"2.0","method":"eth_getLogs","params":[{"fromBlock":"0x2710","toBlock":"latest","address":["0x00112233445566778899aabbccddeeff00112233"],"topics":[["0x00000000000000000000000000112233445566778899aabbccddeeff00112233"]]}],"id":1}`, readAll(t, cli.req.Body))
}

func readAll(t *testing.T, r io.Reader) string {
	buf := bytes.NewBuffer(nil)
	_, err := buf.ReadFrom(r)
	require.NoError(t, err)
	return buf.String()
}

func ptr[T any](v T) *T {
	return &v
}
