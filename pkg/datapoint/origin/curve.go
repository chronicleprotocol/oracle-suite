package origin

import (
	"context"
	_ "embed"
	"fmt"
	"math/big"
	"sort"
	"strings"
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

const CurveLoggerTag = "CURVE_ORIGIN"

type CurveConfig struct {
	Client             rpc.RPC
	ContractAddresses  map[string]string
	Contract2Addresses map[string]string
	Logger             log.Logger
	Blocks             []int64
}

type Curve struct {
	client             rpc.RPC
	contractAddresses  ContractAddresses
	contract2Addresses ContractAddresses
	erc20              *ERC20
	blocks             []int64
	logger             log.Logger
}

func NewCurve(config CurveConfig) (*Curve, error) {
	if config.Client == nil {
		return nil, fmt.Errorf("cannot nil ethereum client")
	}
	if config.Logger == nil {
		config.Logger = null.New()
	}

	addresses, err := convertAddressMap(config.ContractAddresses)
	if err != nil {
		return nil, err
	}
	addresses2, err := convertAddressMap(config.Contract2Addresses)
	if err != nil {
		return nil, err
	}
	erc20, err := NewERC20(config.Client)
	if err != nil {
		return nil, err
	}

	return &Curve{
		client:             config.Client,
		contractAddresses:  addresses,
		contract2Addresses: addresses2,
		erc20:              erc20,
		blocks:             config.Blocks,
		logger:             config.Logger.WithField("curve", CurveLoggerTag),
	}, nil
}

// Get all the tokens added in the pools and arrange them according to the pool address
func (c *Curve) getTokensInPools(ctx context.Context, contractAddresses ContractAddresses, pairs []value.Pair) (
	map[types.Address][]types.Address,
	error,
) {

	coins := abi.MustParseMethod("coins(uint256)(address)")

	var callsToken []types.Call
	var pools []types.Address
	for _, pair := range pairs {
		pool, _, err := contractAddresses.ByPair(pair)
		if err != nil {
			continue
		}
		// tricky: Get the key of pool defined in config to figure out number of tokens in pool
		var poolKey string
		for key, value := range contractAddresses {
			if value == pool {
				poolKey = key
			}
		}
		tokens := strings.Split(poolKey, "/")
		maxCoins := len(tokens) // number of tokens in the pool

		for i := 0; i < maxCoins; i++ {
			callData, err := coins.EncodeArgs(i)
			if err != nil {
				continue
			}
			callsToken = append(callsToken, types.Call{
				To:    &pool,
				Input: callData,
			})
			pools = append(pools, pool)
		}
	}
	if len(callsToken) < 1 { // nothing pairs matched in pools
		return nil, nil
	}

	resp, err := ethereum.MultiCall(ctx, c.client, callsToken, types.LatestBlockNumber)
	if err != nil {
		return nil, err
	}

	tokensInPools := make(map[types.Address][]types.Address)
	for i := range resp {
		var address types.Address
		if err := coins.DecodeValues(resp[i], &address); err != nil {
			return nil, fmt.Errorf("failed decoding tokens in the pool: %w", err)
		}
		tokensInPools[pools[i]] = append(tokensInPools[pools[i]], address)
	}
	return tokensInPools, nil
}

func getTokenIndex(tokens []types.Address, finding types.Address) int {
	for i, token := range tokens {
		if token == finding {
			return i
		}
	}
	return -1
}

//nolint:funlen,gocyclo
func (c *Curve) fetchDataPoints(
	ctx context.Context,
	contractAddresses ContractAddresses,
	pairs []value.Pair,
	secondary bool,
	block *big.Int,
) (
	map[value.Pair]datapoint.Point,
	error,
) {

	points := make(map[value.Pair]datapoint.Point)
	var getDy *abi.Method
	if !secondary {
		getDy = abi.MustParseMethod("get_dy(int128,int128,uint256)(uint256)")
	} else {
		getDy = abi.MustParseMethod("get_dy(uint256,uint256,uint256)(uint256)")
	}

	tokensInPools, err := c.getTokensInPools(ctx, contractAddresses, pairs)
	if err != nil {
		return nil, fmt.Errorf("failed getting tokens in pools: %w", err)
	}
	if tokensInPools == nil {
		return nil, nil
	}

	tokensMap := make(map[types.Address]struct{})
	for _, tokens := range tokensInPools {
		for _, address := range tokens {
			tokensMap[address] = struct{}{}
		}
	}
	tokenDetails, err := c.erc20.GetSymbolAndDecimals(ctx, maps.Keys(tokensMap))
	if err != nil {
		return nil, fmt.Errorf("failed getting symbol & decimals for tokens of pool: %w", err)
	}

	totals := make([]*big.Float, len(pairs))
	var calls []types.Call
	n := 0
	for _, pair := range pairs {
		pool, _, err := contractAddresses.ByPair(pair)
		if err != nil {
			continue
		}
		tokensInPool, ok := tokensInPools[pool]
		if !ok {
			points[pair] = datapoint.Point{Error: fmt.Errorf("no tokens in pool")}
			continue
		}

		if _, ok := tokenDetails[pair.Base]; !ok {
			points[pair] = datapoint.Point{Error: fmt.Errorf("not found base token: %s", pair.Base)}
			continue
		}
		if _, ok := tokenDetails[pair.Quote]; !ok {
			points[pair] = datapoint.Point{Error: fmt.Errorf("not found quote token: %s", pair.Quote)}
			continue
		}
		baseToken := tokenDetails[pair.Base]
		quoteToken := tokenDetails[pair.Quote]

		baseIndex := getTokenIndex(tokensInPool, baseToken.address)
		quoteIndex := getTokenIndex(tokensInPool, quoteToken.address)

		var callData types.Bytes
		if baseIndex < quoteIndex {
			callData, err = getDy.EncodeArgs(
				baseIndex,
				quoteIndex,
				new(big.Int).Exp(big.NewInt(10), big.NewInt(int64(baseToken.decimals)), nil),
			)
		} else {
			callData, err = getDy.EncodeArgs(
				quoteIndex,
				baseIndex,
				new(big.Int).Exp(big.NewInt(10), big.NewInt(int64(quoteToken.decimals)), nil),
			)
		}

		if err != nil {
			points[pair] = datapoint.Point{Error: fmt.Errorf(
				"failed to get contract args for pair: %s: %w",
				pair.String(),
				err,
			)}
			continue
		}
		calls = append(calls, types.Call{
			To:    &pool,
			Input: callData,
		})
		totals[n] = new(big.Float).SetInt64(0)
		n++
	}

	if len(calls) > 0 {
		for _, blockDelta := range c.blocks {
			resp, err := ethereum.MultiCall(ctx, c.client, calls, types.BlockNumberFromUint64(uint64(block.Int64()-blockDelta)))
			if err != nil {
				return nil, err
			}

			n = 0
			for _, pair := range pairs {
				pool, _, err := contractAddresses.ByPair(pair)
				if err != nil {
					continue
				}
				if points[pair].Error != nil {
					continue
				}
				tokensInPool := tokensInPools[pool]

				baseToken := tokenDetails[pair.Base]
				quoteToken := tokenDetails[pair.Quote]

				baseIndex := getTokenIndex(tokensInPool, baseToken.address)
				quoteIndex := getTokenIndex(tokensInPool, quoteToken.address)

				price := new(big.Float).SetInt(new(big.Int).SetBytes(resp[n][0:32]))
				// price = price / 10 ^ quoteDecimals
				if baseIndex < quoteIndex {
					price = new(big.Float).Quo(
						price,
						new(big.Float).SetInt(new(big.Int).Exp(big.NewInt(10), big.NewInt(int64(quoteToken.decimals)), nil)),
					)
				} else {
					price = new(big.Float).Quo(
						price,
						new(big.Float).SetInt(new(big.Int).Exp(big.NewInt(10), big.NewInt(int64(baseToken.decimals)), nil)),
					)
				}
				totals[n] = totals[n].Add(totals[n], price)
				n++
			}
		}
	}

	n = 0
	for _, pair := range pairs {
		pool, _, err := contractAddresses.ByPair(pair)
		if err != nil {
			continue
		}
		if points[pair].Error != nil {
			continue
		}
		tokensInPool := tokensInPools[pool]

		baseToken := tokenDetails[pair.Base]
		quoteToken := tokenDetails[pair.Quote]

		baseIndex := getTokenIndex(tokensInPool, baseToken.address)
		quoteIndex := getTokenIndex(tokensInPool, quoteToken.address)

		avgPrice := new(big.Float).Quo(totals[n], new(big.Float).SetUint64(uint64(len(c.blocks))))
		n++

		// Invert the price if inverted price
		if baseIndex > quoteIndex {
			avgPrice = new(big.Float).Quo(new(big.Float).SetUint64(1), avgPrice)
		}

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

func (c *Curve) FetchDataPoints(ctx context.Context, query []any) (map[any]datapoint.Point, error) {
	pairs, ok := queryToPairs(query)
	if !ok {
		return nil, fmt.Errorf("invalid query type: %T, expected []Pair", query)
	}

	sort.Slice(pairs, func(i, j int) bool {
		return pairs[i].String() < pairs[j].String()
	})

	block, err := c.client.BlockNumber(ctx)

	if err != nil {
		return nil, fmt.Errorf("cannot get block number, %w", err)
	}

	points1, err1 := c.fetchDataPoints(ctx, c.contractAddresses, pairs, false, block)
	points2, err2 := c.fetchDataPoints(ctx, c.contract2Addresses, pairs, true, block)
	if err1 != nil {
		return nil, err1
	}
	if err2 != nil {
		return nil, err2
	}
	if points1 == nil && points2 == nil {
		return nil, fmt.Errorf("failed to fetch data points")
	}

	points := make(map[any]datapoint.Point)
	for pair, point := range points1 {
		points[pair] = point
	}
	for pair, point := range points2 {
		points[pair] = point
	}
	return points, nil
}
