package origin

import (
	"context"
	"fmt"
	"time"

	"github.com/defiweb/go-eth/abi"
	"github.com/defiweb/go-eth/rpc"
	"github.com/defiweb/go-eth/types"

	"github.com/chronicleprotocol/oracle-suite/pkg/ethereum"
	"github.com/chronicleprotocol/oracle-suite/pkg/pricenext/provider"
	"github.com/chronicleprotocol/oracle-suite/pkg/util/bn"
)

type GenericETHContract struct {
	Method    *abi.Method
	Contract  types.Address
	Arguments []any
	Decimals  uint8
}

// GenericEVM is a generic EVM-based price provider.
//
// It can fetch a price from a smart contract method that returns a price
// as an uint256 value.
type GenericEVM struct {
	client    rpc.RPC
	contracts map[provider.Pair]GenericETHContract
}

func NewGenericEVM(client rpc.RPC, contracts map[provider.Pair]GenericETHContract) (*GenericEVM, error) {
	return &GenericEVM{
		client:    client,
		contracts: contracts,
	}, nil
}

func (g GenericEVM) FetchTicks(ctx context.Context, pairs []provider.Pair) []provider.Tick {
	ticks := make([]provider.Tick, len(pairs))
	calls := make([]types.Call, len(pairs))
	for i, pair := range pairs {
		contract, ok := g.contracts[pair]
		if !ok {
			return withError(pairs, ErrPairNotSupported{Pair: pair})
		}
		input, err := contract.Method.EncodeArgs(contract.Arguments...)
		if err != nil {
			return withError(pairs, fmt.Errorf("failed to encode arguments: %w", err))
		}
		calls[i] = types.Call{
			To:    &contract.Contract,
			Input: input,
		}
	}
	results, err := ethereum.MultiCall(ctx, g.client, calls)
	if err != nil {
		return withError(pairs, fmt.Errorf("failed to call contract: %w", err))
	}
	if len(results) != len(pairs) {
		return withError(pairs, fmt.Errorf("unexpected number of results: %d", len(results)))
	}
	for i, pair := range pairs {
		if abi.IsRevert(results[i]) {
			ticks[i] = provider.Tick{
				Pair:  pairs[i],
				Error: fmt.Errorf("contract reverted: %s", abi.DecodeRevert(results[i])),
			}
			continue
		}
		if abi.IsPanic(results[i]) {
			ticks[i] = provider.Tick{
				Pair:  pairs[i],
				Error: fmt.Errorf("contract panicked: %s", abi.DecodePanic(results[i])),
			}
			continue
		}
		if len(results[i]) != abi.WordLength {
			ticks[i] = provider.Tick{
				Pair:  pairs[i],
				Error: fmt.Errorf("unexpected result length: %d", len(results[i])),
			}
			continue
		}
		ticks[i] = provider.Tick{
			Pair:  pairs[i],
			Price: bn.Float(bn.Int(results[i])).Div(bn.Int(10).Pow(g.contracts[pair].Decimals)),
			Time:  time.Now(),
		}
	}
	return ticks
}
