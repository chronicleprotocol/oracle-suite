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
	"errors"
	"fmt"
	"math/big"
	"regexp"

	"github.com/ethereum/go-ethereum"
	"github.com/ethereum/go-ethereum/accounts/abi"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/ethereum/go-ethereum/rpc"

	pkgEthereum "github.com/chronicleprotocol/oracle-suite/pkg/ethereum"
)

const (
	mainnetChainID = 1
	kovanChainID   = 42
	rinkebyChainID = 4
	gorliChainID   = 5
	ropstenChainID = 3
	xdaiChainID    = 100
)

// Addresses of multicall contracts. They're used to implement
// the Client.MultiCall function.
//
// https://github.com/makerdao/multicall
var multiCallContracts = map[uint64]common.Address{
	mainnetChainID: common.HexToAddress("0xeefba1e63905ef1d7acba5a8513c70307c1ce441"),
	kovanChainID:   common.HexToAddress("0x2cc8688c5f75e365aaeeb4ea8d6a480405a48d2a"),
	rinkebyChainID: common.HexToAddress("0x42ad527de7d4e9d9d011ac45b31d8551f8fe9821"),
	gorliChainID:   common.HexToAddress("0x77dca2c955b15e9de4dbbcf1246b4b85b651e50e"),
	ropstenChainID: common.HexToAddress("0x53c43764255c17bd724f74c4ef150724ac50a3ed"),
	xdaiChainID:    common.HexToAddress("0xb5b692a88bdfc81ca69dcb1d924f59f0413a602a"),
}

var ErrMulticallNotSupported = errors.New("multicall is not supported on current chain")
var ErrInvalidSignedTxType = errors.New("unable to send transaction, SignedTx field have invalid type")

// ErrRevert may be returned by Client.Call method in case of EVM revert.
type ErrRevert struct {
	Message string
	Err     error
}

func (e ErrRevert) Error() string {
	return fmt.Sprintf("reverted: %s", e.Message)
}

func (e ErrRevert) Unwrap() error {
	return e.Err
}

// EthClient represents the Ethereum client, like the ethclient.Client.
type EthClient interface {
	SendTransaction(ctx context.Context, tx *types.Transaction) error
	StorageAt(ctx context.Context, account common.Address, key common.Hash, block *big.Int) ([]byte, error)
	CallContract(ctx context.Context, call ethereum.CallMsg, block *big.Int) ([]byte, error)
	NonceAt(ctx context.Context, account common.Address, block *big.Int) (uint64, error)
	PendingNonceAt(ctx context.Context, account common.Address) (uint64, error)
	SuggestGasPrice(ctx context.Context) (*big.Int, error)
	SuggestGasTipCap(ctx context.Context) (*big.Int, error)
	NetworkID(ctx context.Context) (*big.Int, error)
	BlockNumber(ctx context.Context) (uint64, error)
	FilterLogs(ctx context.Context, q ethereum.FilterQuery) ([]types.Log, error)
}

// Client implements the ethereum.Client interface.
type Client struct {
	ethClient EthClient
	signer    pkgEthereum.Signer
}

// NewClient returns a new Client instance.
func NewClient(ethClient EthClient, signer pkgEthereum.Signer) *Client {
	return &Client{
		ethClient: ethClient,
		signer:    signer,
	}
}

// BlockNumber implements the ethereum.Client interface.
func (e *Client) BlockNumber(ctx context.Context) (*big.Int, error) {
	n, err := e.ethClient.BlockNumber(ctx)
	if err != nil {
		return nil, err
	}
	return new(big.Int).SetUint64(n), nil
}

// Call implements the ethereum.Client interface.
func (e *Client) Call(ctx context.Context, call pkgEthereum.Call) ([]byte, error) {
	addr := common.Address{}
	if e.signer != nil {
		addr = e.signer.Address()
	}

	cm := ethereum.CallMsg{
		From:     addr,
		To:       &call.Address,
		Gas:      0,
		GasPrice: nil,
		Value:    nil,
		Data:     call.Data,
	}

	resp, err := e.ethClient.CallContract(ctx, cm, pkgEthereum.BlockNumberFromContext(ctx))
	if err := isRevertErr(err); err != nil {
		return nil, err
	}
	if err := isRevertResp(resp); err != nil {
		return nil, err
	}
	if err != nil {
		return nil, err
	}

	return resp, err
}

func (e *Client) CallBlocks(ctx context.Context, call pkgEthereum.Call, blocks []int64) ([][]byte, error) {
	blockNumber, err := e.BlockNumber(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to get block number: %w", err)
	}

	var resps [][]byte
	for _, block := range blocks {
		r, err := e.Call(
			pkgEthereum.WithBlockNumber(ctx, new(big.Int).Sub(blockNumber, big.NewInt(block))),
			call,
		)
		if err != nil {
			return nil, err
		}
		resps = append(resps, r)
	}

	return resps, nil
}

// MultiCall implements the ethereum.Client interface.
func (e *Client) MultiCall(ctx context.Context, calls []pkgEthereum.Call) ([][]byte, error) {
	type abiCall struct {
		Address common.Address `abi:"target"`
		Data    []byte         `abi:"callData"`
	}
	var abiCalls []abiCall
	for _, c := range calls {
		abiCalls = append(abiCalls, abiCall{
			Address: c.Address,
			Data:    c.Data,
		})
	}

	chainID, err := e.ethClient.NetworkID(ctx)
	if err != nil {
		return nil, err
	}
	multicallAddr, ok := multiCallContracts[chainID.Uint64()]
	if !ok {
		return nil, ErrMulticallNotSupported
	}
	callData, err := multiCallABI.Pack("aggregate", abiCalls)
	if err != nil {
		return nil, err
	}
	response, err := e.Call(ctx, pkgEthereum.Call{Address: multicallAddr, Data: callData})
	if err != nil {
		return nil, err
	}
	results, err := multiCallABI.Unpack("aggregate", response)
	if err != nil {
		return nil, err
	}

	return results[1].([][]byte), nil
}

// Storage implements the ethereum.Client interface.
func (e *Client) Storage(ctx context.Context, address pkgEthereum.Address, key pkgEthereum.Hash) ([]byte, error) {
	return e.ethClient.StorageAt(ctx, address, key, pkgEthereum.BlockNumberFromContext(ctx))
}

// SendTransaction implements the ethereum.Client interface.
func (e *Client) SendTransaction(ctx context.Context, transaction *pkgEthereum.Transaction) (*pkgEthereum.Hash, error) {
	var err error

	// We don't want to modify passed structure because that would be rude, so
	// we copy it here:
	tx := &pkgEthereum.Transaction{
		Address:     transaction.Address,
		Nonce:       transaction.Nonce,
		PriorityFee: transaction.PriorityFee,
		MaxFee:      transaction.MaxFee,
		GasLimit:    transaction.GasLimit,
		ChainID:     transaction.ChainID,
		SignedTx:    transaction.SignedTx,
	}
	tx.Data = make([]byte, len(transaction.Data))
	copy(tx.Data, transaction.Data)

	// Fill optional values if necessary:
	if tx.Nonce == 0 {
		tx.Nonce, err = e.ethClient.PendingNonceAt(ctx, e.signer.Address())
		if err != nil {
			return nil, err
		}
	}
	if tx.PriorityFee == nil {
		suggestedGasTipPrice, err := e.ethClient.SuggestGasTipCap(ctx)
		if err != nil {
			return nil, err
		}
		tx.PriorityFee = suggestedGasTipPrice
	}
	if tx.MaxFee == nil {
		suggestedGasPrice, err := e.ethClient.SuggestGasPrice(ctx)
		if err != nil {
			return nil, err
		}
		tx.MaxFee = new(big.Int).Mul(suggestedGasPrice, big.NewInt(2))
	}
	if tx.ChainID == nil {
		tx.ChainID, err = e.ethClient.NetworkID(ctx)
		if err != nil {
			return nil, err
		}
	}
	if tx.SignedTx == nil {
		err = e.signer.SignTransaction(tx)
		if err != nil {
			return nil, err
		}
	}

	// Send transaction:
	if stx, ok := tx.SignedTx.(*types.Transaction); ok {
		hash := stx.Hash()
		return &hash, e.ethClient.SendTransaction(ctx, stx)
	}
	return nil, ErrInvalidSignedTxType
}

func isRevertResp(resp []byte) error {
	revert, err := abi.UnpackRevert(resp)
	if err != nil {
		return nil
	}

	return ErrRevert{Message: revert, Err: nil}
}

func isRevertErr(vmErr error) error {
	if terr, is := vmErr.(rpc.DataError); is {
		// Some RPC servers returns "revert" data as a hex encoded string, here
		// we're trying to parse it:
		if str, ok := terr.ErrorData().(string); ok {
			re := regexp.MustCompile("(0x[a-zA-Z0-9]+)")
			match := re.FindStringSubmatch(str)

			if len(match) == 2 && len(match[1]) > 2 {
				bytes, err := hex.DecodeString(match[1][2:])
				if err != nil {
					return nil
				}

				revert, err := abi.UnpackRevert(bytes)
				if err != nil {
					return nil
				}

				return ErrRevert{Message: revert, Err: vmErr}
			}
		}
	}

	return nil
}
