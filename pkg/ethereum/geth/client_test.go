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

package geth

import (
	"context"
	"encoding/hex"
	"math/big"
	"testing"

	"github.com/ethereum/go-ethereum"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"

	pkgEthereum "github.com/chronicleprotocol/oracle-suite/pkg/ethereum"
	"github.com/chronicleprotocol/oracle-suite/pkg/ethereum/geth/mocks"
)

var clientContractAddress = common.HexToAddress("0x0E30F0FC91FDbc4594b1e2E5d64E6F1f94cAB23D")
var clientAddress = common.HexToAddress("0x2d800d93b065ce011af83f316cef9f0d005b0aa4")
var clientCallData = common.Hex2Bytes("095ea7b3000000000000000000000000")
var clientCallResp = common.Hex2Bytes("00000000000000000000000000000000")
var clientCallRevertResp = common.Hex2Bytes("08c379a0000000000000000000000000000000000000000000000000000000000000002000000000000000000000000000000000000000000000000000000000000000094e6f74206f776e65720000000000000000000000000000000000000000000000")
var clientMultiCallResp = common.Hex2Bytes("0000000000000000000000000000000000000000000000000000000001511c6e00000000000000000000000000000000000000000000000000000000000000400000000000000000000000000000000000000000000000000000000000000004000000000000000000000000000000000000000000000000000000000000008000000000000000000000000000000000000000000000000000000000000000c0000000000000000000000000000000000000000000000000000000000000010000000000000000000000000000000000000000000000000000000000000001400000000000000000000000000000000000000000000000000000000000000020000000000000000000000000005b903dadfd96229cba5eb0e5aa75c578e8f968000000000000000000000000000000000000000000000000000000000000002000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000020000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000200000000000000000000000000000000000000000000000000000000000000000")

type DataErr struct {
	Err string
}

func (d DataErr) Error() string {
	return d.Err
}

func (d DataErr) ErrorData() interface{} {
	return d.Err
}

func TestClient_Call(t *testing.T) {
	account, _ := NewAccount("./testdata/keystore", "test123", clientAddress)
	ethClient := &mocks.EthClient{}
	client := NewClient(ethClient, NewSigner(account))

	ethClient.On(
		"CallContract",
		mock.Anything,
		mock.Anything,
		(*big.Int)(nil),
	).Return(clientCallResp, nil)

	resp, err := client.Call(
		context.Background(),
		pkgEthereum.Call{Address: clientContractAddress, Data: clientCallData},
	)

	cm := ethClient.Calls()[0].Arguments.Get(1).(ethereum.CallMsg)

	assert.Equal(t, resp, clientCallResp)
	assert.Equal(t, clientCallData, cm.Data)
	assert.Equal(t, clientAddress, cm.From)
	assert.Equal(t, clientContractAddress, *cm.To)
	assert.NoError(t, err)
}

func TestClient_Call_Reverted(t *testing.T) {
	account, _ := NewAccount("./testdata/keystore", "test123", clientAddress)
	ethClient := &mocks.EthClient{}
	client := NewClient(ethClient, NewSigner(account))

	ethClient.On(
		"CallContract",
		mock.Anything,
		mock.Anything,
		(*big.Int)(nil),
	).Return(clientCallRevertResp, nil)

	resp, err := client.Call(
		context.Background(),
		pkgEthereum.Call{Address: clientContractAddress, Data: clientCallData},
	)

	assert.Nil(t, resp)
	assert.Error(t, err)
	assert.IsType(t, ErrRevert{}, err)
	assert.Equal(t, "reverted: Not owner", err.Error())
}

func TestClient_Call_RevertedDataError(t *testing.T) {
	account, _ := NewAccount("./testdata/keystore", "test123", clientAddress)
	ethClient := &mocks.EthClient{}
	client := NewClient(ethClient, NewSigner(account))

	ethClient.On(
		"CallContract",
		mock.Anything,
		mock.Anything,
		(*big.Int)(nil),
	).Return([]byte(nil), DataErr{Err: "Reverted: 0x" + hex.EncodeToString(clientCallRevertResp)})

	resp, err := client.Call(
		context.Background(),
		pkgEthereum.Call{Address: clientContractAddress, Data: clientCallData},
	)

	assert.Nil(t, resp)
	assert.Error(t, err)
	assert.IsType(t, ErrRevert{}, err)
	assert.Equal(t, "reverted: Not owner", err.Error())
}

func TestClient_MultiCall(t *testing.T) {
	account, _ := NewAccount("./testdata/keystore", "test123", clientAddress)
	ethClient := &mocks.EthClient{}
	client := NewClient(ethClient, NewSigner(account))

	ethClient.On(
		"NetworkID",
		mock.Anything,
	).Return(big.NewInt(mainnetChainID), nil)

	ethClient.On(
		"CallContract",
		mock.Anything,
		mock.Anything,
		(*big.Int)(nil),
	).Return(clientMultiCallResp, nil)

	resp, err := client.MultiCall(
		context.Background(),
		[]pkgEthereum.Call{
			{Address: clientContractAddress, Data: clientCallData},
			{Address: clientContractAddress, Data: clientCallData},
			{Address: clientContractAddress, Data: clientCallData},
			{Address: clientContractAddress, Data: clientCallData},
		},
	)

	cm := ethClient.Calls()[1].Arguments.Get(1).(ethereum.CallMsg)

	assert.NotNil(t, resp)
	assert.Len(t, resp, 4)
	assert.Equal(t, clientAddress, cm.From)
	assert.Equal(t, multiCallContracts[mainnetChainID], *cm.To)
	assert.NoError(t, err)
}

func TestClient_Storage(t *testing.T) {
	ethClient := &mocks.EthClient{}
	client := NewClient(ethClient, NewSigner(nil))

	data := []byte{0}
	hash := common.BytesToHash([]byte{0})

	ethClient.On(
		"StorageAt",
		mock.Anything,
		clientContractAddress,
		hash,
		(*big.Int)(nil),
	).Return(data, nil)

	resp, err := client.Storage(context.Background(), clientContractAddress, hash)

	assert.NoError(t, err)
	assert.Equal(t, data, resp)
}

func TestClient_SendTransaction(t *testing.T) {
	account, _ := NewAccount("./testdata/keystore", "test123", clientAddress)
	ethClient := &mocks.EthClient{}
	client := NewClient(ethClient, NewSigner(account))

	ethClient.On(
		"SendTransaction",
		mock.Anything,
		mock.Anything,
	).Return(nil)

	tx := &pkgEthereum.Transaction{
		Address:     clientContractAddress,
		Nonce:       10,
		PriorityFee: big.NewInt(50),
		MaxFee:      big.NewInt(100),
		GasLimit:    big.NewInt(1000),
		Data:        clientCallData,
		ChainID:     big.NewInt(mainnetChainID),
		SignedTx:    nil,
	}

	hash, err := client.SendTransaction(context.Background(), tx)
	stx := ethClient.Calls()[0].Arguments.Get(1).(*types.Transaction)

	assert.NotNil(t, hash)
	assert.NoError(t, err)
	assert.Equal(t, clientContractAddress, *stx.To())
	assert.Equal(t, uint64(10), stx.Nonce())
	assert.Equal(t, big.NewInt(100), stx.GasFeeCap())
	assert.Equal(t, big.NewInt(50), stx.GasTipCap())
	assert.Equal(t, uint64(1000), stx.Gas())
	assert.Equal(t, big.NewInt(mainnetChainID), stx.ChainId())
	assert.Equal(t, clientCallData, stx.Data())
}

func TestClient_SendTransaction_Minimal(t *testing.T) {
	account, _ := NewAccount("./testdata/keystore", "test123", clientAddress)
	ethClient := &mocks.EthClient{}
	client := NewClient(ethClient, NewSigner(account))

	ethClient.On(
		"PendingNonceAt",
		mock.Anything,
		clientAddress,
	).Return(10, nil)

	ethClient.On(
		"SuggestGasTipCap",
		mock.Anything,
	).Return(big.NewInt(10), nil)

	ethClient.On(
		"SuggestGasPrice",
		mock.Anything,
	).Return(big.NewInt(70), nil)

	ethClient.On(
		"NetworkID",
		mock.Anything,
	).Return(big.NewInt(mainnetChainID), nil)

	ethClient.On(
		"SendTransaction",
		mock.Anything,
		mock.Anything,
	).Return(nil)

	tx := &pkgEthereum.Transaction{
		Address:  clientContractAddress,
		GasLimit: big.NewInt(1000),
		Data:     clientCallData,
		SignedTx: nil,
	}

	hash, err := client.SendTransaction(context.Background(), tx)
	stx := ethClient.Calls()[4].Arguments.Get(1).(*types.Transaction)

	assert.NotNil(t, hash)
	assert.NoError(t, err)
	assert.Equal(t, clientContractAddress, *stx.To())
	assert.Equal(t, uint64(10), stx.Nonce())
	assert.Equal(t, big.NewInt(10), stx.GasTipCap())
	assert.Equal(t, big.NewInt(140), stx.GasFeeCap())
	assert.Equal(t, uint64(1000), stx.Gas())
	assert.Equal(t, big.NewInt(mainnetChainID), stx.ChainId())
}
