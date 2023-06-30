package origin

import (
	"context"
	_ "embed"
	"fmt"
	"math/big"
	"sort"
	"time"

	"github.com/defiweb/go-eth/abi"
	"github.com/defiweb/go-eth/rpc"
	"github.com/defiweb/go-eth/types"
	"golang.org/x/exp/maps"

	"github.com/chronicleprotocol/oracle-suite/pkg/datapoint"
	"github.com/chronicleprotocol/oracle-suite/pkg/datapoint/value"
	"github.com/chronicleprotocol/oracle-suite/pkg/ethereum"
	"github.com/chronicleprotocol/oracle-suite/pkg/log"
	"github.com/chronicleprotocol/oracle-suite/pkg/log/null"
	"github.com/chronicleprotocol/oracle-suite/pkg/util/bn"
)

//go:embed uniswap_v3_pool_abi.json
var uniswapV3PoolABI []byte

const UniswapV3LoggerTag = "UNISWAPV3_ORIGIN"

type UniswapV3Options struct {
	Client            rpc.RPC
	ContractAddresses ContractAddresses
	Logger            log.Logger
	Blocks            []int64
}

type UniswapV3 struct {
	client            rpc.RPC
	contractAddresses ContractAddresses
	erc20             *ERC20
	abi               *abi.Contract
	blocks            []int64
	logger            log.Logger
}

func NewUniswapV3(opts UniswapV3Options) (*UniswapV3, error) {
	if opts.Client == nil {
		return nil, fmt.Errorf("cannot nil ethereum client")
	}
	if opts.Logger == nil {
		opts.Logger = null.New()
	}

	a, err := abi.ParseJSON(uniswapV3PoolABI)
	if err != nil {
		return nil, err
	}

	erc20, err := NewERC20(opts.Client)
	if err != nil {
		return nil, err
	}

	return &UniswapV3{
		client:            opts.Client,
		contractAddresses: opts.ContractAddresses,
		erc20:             erc20,
		abi:               a,
		blocks:            opts.Blocks,
		logger:            opts.Logger.WithField("uniswapV3", UniswapV3LoggerTag),
	}, nil
}

//nolint:funlen,gocyclo
func (u *UniswapV3) FetchDataPoints(ctx context.Context, query []any) (map[any]datapoint.Point, error) {
	pairs, ok := queryToPairs(query)
	if !ok {
		return nil, fmt.Errorf("invalid query type: %T, expected []Pair", query)
	}

	sort.Slice(pairs, func(i, j int) bool {
		return pairs[i].String() < pairs[j].String()
	})

	points := make(map[any]datapoint.Point)

	block, err := u.client.BlockNumber(ctx)

	if err != nil {
		return nil, fmt.Errorf("cannot get block number, %w", err)
	}

	totals := make([]*big.Float, len(pairs))
	var calls []types.Call
	var callsToken []types.Call
	for i, pair := range pairs {
		contract, _, err := u.contractAddresses.AddressByPair(pair)
		if err != nil {
			return nil, err
		}
		// Calls for `slot0`
		callData, err := u.abi.Methods["slot0"].EncodeArgs()
		if err != nil {
			return nil, fmt.Errorf("failed to get slot0 for pair: %s: %w", pair.String(), err)
		}
		calls = append(calls, types.Call{
			To:    &contract,
			Input: callData,
		})
		// Calls for `token0`
		callData, err = u.abi.Methods["token0"].EncodeArgs()
		if err != nil {
			return nil, fmt.Errorf("failed to get token0 for pair: %s: %w", pair.String(), err)
		}
		callsToken = append(callsToken, types.Call{
			To:    &contract,
			Input: callData,
		})
		// Calls for `token1`
		callData, err = u.abi.Methods["token1"].EncodeArgs()
		if err != nil {
			return nil, fmt.Errorf("failed to get token1 for pair: %s: %w", pair.String(), err)
		}
		callsToken = append(callsToken, types.Call{
			To:    &contract,
			Input: callData,
		})

		totals[i] = new(big.Float).SetInt64(0)
	}

	// Get decimals for all the tokens
	tokensMap := make(map[types.Address]struct{})
	resp, err := ethereum.MultiCall(ctx, u.client, callsToken, types.LatestBlockNumber)
	if err != nil {
		return nil, fmt.Errorf("failed multicall for tokens of pool: %w", err)
	}
	for i := range resp {
		var address types.Address
		if err := u.abi.Methods["token0"].DecodeValues(resp[i], &address); err != nil {
			return nil, fmt.Errorf("failed decoding token address of pool: %w", err)
		}
		tokensMap[address] = struct{}{}
	}
	tokenDetails, err := u.erc20.GetSymbolAndDecimals(ctx, maps.Keys(tokensMap))
	if err != nil {
		return nil, fmt.Errorf("failed getting symbol & decimals for tokens of pool: %w", err)
	}

	// 2 ^ 192
	q192 := new(big.Int).Exp(big.NewInt(2), big.NewInt(192), nil)
	for _, blockDelta := range u.blocks {
		resp, err := ethereum.MultiCall(ctx, u.client, calls, types.BlockNumberFromUint64(uint64(block.Int64()-blockDelta)))
		if err != nil {
			return nil, fmt.Errorf("failed multicall: %w", err)
		}
		if len(calls) != len(resp) {
			return nil, fmt.Errorf("unexpected number of multicall results, expected %d, got %d",
				len(calls), len(resp))
		}
		if len(resp) != len(pairs) {
			return nil, fmt.Errorf("unexpected number of multicall results with pairs, expected %d, got %d",
				len(resp), len(pairs))
		}

		for i, pair := range pairs {
			sqrtRatioX96 := new(big.Int).SetBytes(resp[i][0:32])
			// ratioX192 = sqrtRatioX96 ^ 2
			ratioX192 := new(big.Int).Mul(sqrtRatioX96, sqrtRatioX96)

			if _, ok := tokenDetails[pair.Base]; !ok {
				return nil, fmt.Errorf("not found base token: %s", pair.Base)
			}
			if _, ok := tokenDetails[pair.Quote]; !ok {
				return nil, fmt.Errorf("not found quote token: %s", pair.Quote)
			}

			baseToken := tokenDetails[pair.Base]
			quoteToken := tokenDetails[pair.Quote]
			// baseAmount = 10 ^ baseDecimals
			baseAmount := new(big.Int).Exp(big.NewInt(10), big.NewInt(int64(baseToken.decimals)), nil)

			// Reference: https://github.com/Uniswap/v3-periphery/blob/main/contracts/libraries/OracleLibrary.sol#L60
			var quoteAmount *big.Int
			if baseToken.address.String() < quoteToken.address.String() {
				// quoteAmount = ratioX192 * baseAmount / (2 ^ 192)
				quoteAmount = new(big.Int).Div(new(big.Int).Mul(ratioX192, baseAmount), q192)
			} else {
				// quoteAmount = (2 ^ 192) * baseAmount / ratioX192
				quoteAmount = new(big.Int).Div(new(big.Int).Mul(q192, baseAmount), ratioX192)
			}

			// price = quoteAmount / 10 ^ quoteDecimals
			price := new(big.Float).Quo(
				new(big.Float).SetInt(quoteAmount),
				new(big.Float).SetInt(new(big.Int).Exp(big.NewInt(10), big.NewInt(int64(quoteToken.decimals)), nil)),
			)
			totals[i] = totals[i].Add(totals[i], price)
		}
	}

	for i, pair := range pairs {
		avgPrice := new(big.Float).Quo(totals[i], new(big.Float).SetUint64(uint64(len(u.blocks))))

		tick := value.Tick{
			Pair:      pair,
			Price:     bn.Float(avgPrice),
			Volume24h: nil,
		}
		points[pair] = datapoint.Point{
			Value: tick,
			Time:  time.Now(),
		}
	}

	return points, nil
}
