package origin

import (
	"context"
	_ "embed"
	"fmt"
	"math/big"
	"strings"

	"github.com/defiweb/go-eth/abi"
	"github.com/defiweb/go-eth/rpc"
	"github.com/defiweb/go-eth/types"

	"github.com/chronicleprotocol/oracle-suite/pkg/ethereum"
)

//go:embed erc20_abi.json
var erc20ABI []byte

type ERC20Details struct {
	address  types.Address
	symbol   string
	decimals int
}

type ERC20 struct {
	client rpc.RPC
	abi    *abi.Contract
	cache  map[types.Address]ERC20Details
}

func NewERC20(client rpc.RPC) (*ERC20, error) {
	a, err := abi.ParseJSON(erc20ABI)
	if err != nil {
		return nil, err
	}

	cache := make(map[types.Address]ERC20Details)
	eth := types.MustAddressFromHex("0xeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeee")
	const decimals = 18
	cache[eth] = ERC20Details{
		address:  eth,
		symbol:   "ETH",
		decimals: decimals,
	}
	return &ERC20{
		client: client,
		abi:    a,
		cache:  cache,
	}, nil
}

func (e *ERC20) GetSymbolAndDecimals(ctx context.Context, addresses []types.Address) (map[string]ERC20Details, error) {
	var calls []types.Call
	for i, address := range addresses {
		if _, ok := e.cache[address]; ok {
			continue
		}
		if address == types.ZeroAddress {
			continue
		}

		// Calls for `symbol`
		callData, err := e.abi.Methods["symbol"].EncodeArgs()
		if err != nil {
			return nil, fmt.Errorf("failed to get symbol for address: %s: %w", address.String(), err)
		}
		calls = append(calls, types.Call{
			To:    &addresses[i],
			Input: callData,
		})

		// Calls for `decimals`
		callData, err = e.abi.Methods["decimals"].EncodeArgs()
		if err != nil {
			return nil, fmt.Errorf("failed to get decimals for address: %s: %w", address.String(), err)
		}
		calls = append(calls, types.Call{
			To:    &addresses[i],
			Input: callData,
		})
	}

	var resp [][]byte
	var err error
	if calls != nil {
		resp, err = ethereum.MultiCall(ctx, e.client, calls, types.LatestBlockNumber)
		if err != nil {
			return nil, fmt.Errorf("failed multicall for tokens: %w", err)
		}
	}

	n := 0
	for i, address := range addresses {
		if _, ok := e.cache[address]; ok {
			continue
		}

		var symbol string
		var decimals int
		if strings.ToLower(address.String()) == "0x9f8f72aa9304c8b593d555f12ef6589cc3a579a2" {
			symbol = "MKR"
		} else if err := e.abi.Methods["symbol"].DecodeValues(resp[n*2], &symbol); err != nil {
			return nil, fmt.Errorf("failed decoding token symbol for token %s: %w", address.String(), err)
		}

		decimals = int(new(big.Int).SetBytes(resp[n*2+1]).Int64())
		e.cache[addresses[i]] = ERC20Details{
			address:  addresses[i],
			symbol:   strings.ToUpper(symbol),
			decimals: decimals,
		}
		n++
	}

	details := make(map[string]ERC20Details)
	for _, address := range addresses {
		token := e.cache[address]
		details[token.symbol] = token
	}

	return details, nil
}
