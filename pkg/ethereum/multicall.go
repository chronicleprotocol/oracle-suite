package ethereum

import (
	"context"
	"fmt"

	"github.com/defiweb/go-eth/abi"
	"github.com/defiweb/go-eth/rpc"
	"github.com/defiweb/go-eth/types"
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
var multiCallContracts = map[uint64]types.Address{
	mainnetChainID: types.MustAddressFromHex("0xeefba1e63905ef1d7acba5a8513c70307c1ce441"),
	kovanChainID:   types.MustAddressFromHex("0x2cc8688c5f75e365aaeeb4ea8d6a480405a48d2a"),
	rinkebyChainID: types.MustAddressFromHex("0x42ad527de7d4e9d9d011ac45b31d8551f8fe9821"),
	gorliChainID:   types.MustAddressFromHex("0x77dca2c955b15e9de4dbbcf1246b4b85b651e50e"),
	ropstenChainID: types.MustAddressFromHex("0x53c43764255c17bd724f74c4ef150724ac50a3ed"),
	xdaiChainID:    types.MustAddressFromHex("0xb5b692a88bdfc81ca69dcb1d924f59f0413a602a"),
}

func MultiCall(
	ctx context.Context,
	client rpc.RPC,
	calls []types.Call,
	blockNumber types.BlockNumber,
) ([][]byte, error) {

	type multicallCall struct {
		Target types.Address `abi:"target"`
		Data   []byte        `abi:"callData"`
	}
	var (
		multicallCalls   []multicallCall
		multicallResults [][]byte
	)
	for _, call := range calls {
		if call.To == nil {
			return nil, fmt.Errorf("multicall: call to nil address")
		}
		multicallCalls = append(multicallCalls, multicallCall{
			Target: *call.To,
			Data:   call.Input,
		})
	}
	chainID, err := client.ChainID(ctx)
	if err != nil {
		return nil, fmt.Errorf("multicall: getting chain id failed: %w", err)
	}
	multicallContract, ok := multiCallContracts[chainID]
	if !ok {
		return nil, fmt.Errorf("multicall: unsupported chain id %d", chainID)
	}
	callata, err := multicallMethod.EncodeArgs(multicallCalls)
	if err != nil {
		return nil, fmt.Errorf("multicall: encoding arguments failed: %w", err)
	}
	resp, err := client.Call(ctx, types.Call{
		To:    &multicallContract,
		Input: callata,
	}, blockNumber)
	if err != nil {
		return nil, fmt.Errorf("multicall: call failed: %w", err)
	}
	if err := multicallMethod.DecodeValues(resp, nil, &multicallResults); err != nil {
		return nil, fmt.Errorf("multicall: decoding results failed: %w", err)
	}
	if len(calls) != len(multicallResults) {
		return nil, fmt.Errorf("unexpected number of multicall results, expected %d, got %d",
			len(calls), len(multicallResults))
	}

	return multicallResults, nil
}

var multicallMethod = abi.MustParseMethod(`
	function aggregate(
		(address target, bytes callData)[] memory calls
	) public returns (
		uint256 blockNumber, 
		bytes[] memory returnData
	)`,
)
