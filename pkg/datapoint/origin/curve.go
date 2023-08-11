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

var getDy1 = abi.MustParseMethod("get_dy(int128,int128,uint256)(uint256)")
var getDy2 = abi.MustParseMethod("get_dy(uint256,uint256,uint256)(uint256)")
var coins = abi.MustParseMethod("coins(uint256)(address)")

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
		getDy = getDy1
	} else {
		getDy = getDy2
	}

	// Get all the token addresses and their decimals
	var callsToken []types.Call
	for _, pair := range pairs {
		contract, baseIndex, quoteIndex, err := contractAddresses.ByPair(pair)
		if err != nil {
			continue
		}
		callData, err := coins.EncodeArgs(baseIndex)
		if err != nil {
			continue
		}
		callsToken = append(callsToken, types.Call{
			To:    &contract,
			Input: callData,
		})
		callData, err = coins.EncodeArgs(quoteIndex)
		if err != nil {
			continue
		}
		callsToken = append(callsToken, types.Call{
			To:    &contract,
			Input: callData,
		})
	}

	var tokenDetails map[string]ERC20Details
	if len(callsToken) > 0 {
		resp, err := ethereum.MultiCall(ctx, c.client, callsToken, types.LatestBlockNumber)
		if err != nil {
			return nil, err
		}

		tokensMap := make(map[types.Address]struct{})
		for i := range resp {
			var address types.Address
			if err := coins.DecodeValues(resp[i], &address); err != nil {
				return nil, fmt.Errorf("failed decoding tokens in the pool: %w", err)
			}
			tokensMap[address] = struct{}{}
		}
		tokenDetails, err = c.erc20.GetSymbolAndDecimals(ctx, maps.Keys(tokensMap))
		if err != nil {
			return nil, fmt.Errorf("failed getting symbol & decimals for tokens of pool: %w", err)
		}
	}

	totals := make([]*big.Float, len(pairs))
	var calls []types.Call
	n := 0
	for _, pair := range pairs {
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

		pool, baseIndex, quoteIndex, err := contractAddresses.ByPair(pair)
		if err != nil {
			points[pair] = datapoint.Point{Error: err}
			continue
		}

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
				if points[pair].Error != nil {
					continue
				}
				_, baseIndex, quoteIndex, _ := contractAddresses.ByPair(pair)
				baseToken := tokenDetails[pair.Base]
				quoteToken := tokenDetails[pair.Quote]

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
		if points[pair].Error != nil {
			continue
		}
		avgPrice := new(big.Float).Quo(totals[n], new(big.Float).SetUint64(uint64(len(c.blocks))))
		n++

		// Invert the price if inverted price
		_, baseIndex, quoteIndex, _ := contractAddresses.ByPair(pair)
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

	points := make(map[any]datapoint.Point)

	// Filter pairs into two group according to their data types in `get_dy`; `int128`, `uint256`
	pairs1 := make(map[value.Pair]struct{})
	pairs2 := make(map[value.Pair]struct{})
	for _, pair := range pairs {
		_, baseIndex, quoteIndex, err := c.contractAddresses.ByPair(pair)
		if err == nil && baseIndex >= 0 && quoteIndex >= 0 {
			pairs1[pair] = struct{}{}
			continue
		}
		_, baseIndex, quoteIndex, err = c.contract2Addresses.ByPair(pair)
		if err == nil && baseIndex >= 0 && quoteIndex >= 0 {
			pairs2[pair] = struct{}{}
			continue
		}
		if err != nil {
			points[pair] = datapoint.Point{Error: err}
			continue
		}
	}

	points1, err1 := c.fetchDataPoints(ctx, c.contractAddresses, maps.Keys(pairs1), false, block)
	points2, err2 := c.fetchDataPoints(ctx, c.contract2Addresses, maps.Keys(pairs2), true, block)
	if err1 != nil {
		return points, err1
	}
	if err2 != nil {
		return points, err2
	}
	if points1 == nil && points2 == nil {
		return points, fmt.Errorf("failed to fetch data points")
	}

	for pair, point := range points1 {
		points[pair] = point
	}
	for pair, point := range points2 {
		points[pair] = point
	}
	return points, nil
}
