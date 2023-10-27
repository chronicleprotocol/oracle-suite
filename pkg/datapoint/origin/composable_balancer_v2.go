//  Copyright (C) 2021-2023 Chronicle Labs, Inc.
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

package origin

import (
	"context"
	"fmt"
	"sort"
	"time"

	"github.com/defiweb/go-eth/rpc"
	"github.com/defiweb/go-eth/types"
	"golang.org/x/exp/maps"

	"github.com/chronicleprotocol/oracle-suite/pkg/datapoint"
	"github.com/chronicleprotocol/oracle-suite/pkg/datapoint/value"
	"github.com/chronicleprotocol/oracle-suite/pkg/ethereum"
	"github.com/chronicleprotocol/oracle-suite/pkg/log"
	"github.com/chronicleprotocol/oracle-suite/pkg/log/null"
	"github.com/chronicleprotocol/oracle-suite/pkg/util/bn"
	"github.com/chronicleprotocol/oracle-suite/pkg/util/errutil"
)

const ComposableBalancerV2LoggerTag = "COMPOSABLE_BALANCERV2_ORIGIN"

type ComposableBalancerV2Config struct {
	Client            rpc.RPC
	ContractAddresses ContractAddresses
	Logger            log.Logger
	Blocks            []int64
}

type ComposableBalancerV2 struct {
	client            rpc.RPC
	contractAddresses ContractAddresses
	erc20             *ERC20
	blocks            []int64
	logger            log.Logger
}

// NewComposableBalancerV2 create instance for ComposableStableBalancer
// `ComposableStableBalancer` is just a notable name, it is balancer v2 origin specialized for ComposableStablePool implementation
// https://docs.balancer.fi/concepts/pools/composable-stable.html
// WeightedPool or MetaStablePool was implemented in BalancerV2
func NewComposableBalancerV2(config ComposableBalancerV2Config) (*ComposableBalancerV2, error) {
	if config.Client == nil {
		return nil, fmt.Errorf("ethereum client not set")
	}
	if config.Logger == nil {
		config.Logger = null.New()
	}

	erc20, err := NewERC20(config.Client)
	if err != nil {
		return nil, err
	}

	return &ComposableBalancerV2{
		client:            config.Client,
		contractAddresses: config.ContractAddresses,
		erc20:             erc20,
		blocks:            config.Blocks,
		logger:            config.Logger.WithField("composableBalancerV2", ComposableBalancerV2LoggerTag),
	}, nil
}

//nolint:funlen,gocyclo
func (b *ComposableBalancerV2) FetchDataPoints(ctx context.Context, query []any) (map[any]datapoint.Point, error) {
	pairs, ok := queryToPairs(query)
	if !ok {
		return nil, fmt.Errorf("invalid query type: %T, expected []Pair", query)
	}

	sort.Slice(pairs, func(i, j int) bool {
		return pairs[i].String() < pairs[j].String()
	})

	points := make(map[any]datapoint.Point)

	block, err := b.client.BlockNumber(ctx)
	if err != nil {
		return nil, fmt.Errorf("cannot get block number, %w", err)
	}

	totals := make(map[value.Pair]*bn.DecFloatPointNumber)

	var calls []types.Call
	var pools []*ComposableStablePool
	// Get all the vault address and pool token addresses from vault
	for _, pair := range pairs {
		contract, _, _, err := b.contractAddresses.ByPair(pair)
		if err != nil {
			points[pair] = datapoint.Point{Error: err}
			continue
		}

		c, err := NewComposableStablePool(ComposableStablePoolConfig{
			Pair:            pair,
			ContractAddress: contract,
		})
		if err != nil {
			points[pair] = datapoint.Point{Error: err}
			continue
		}
		pools = append(pools, c)
		calls = append(calls, errutil.Must(c.createInitCalls())...)

		totals[pair] = bn.DecFloatPoint(0)
	}

	if len(calls) < 1 || len(pools) < 1 {
		return nil, fmt.Errorf("not found valid pair")
	}

	// Get pool id and vault address
	resp, err := ethereum.MultiCall(ctx, b.client, calls, types.LatestBlockNumber)
	if err != nil {
		return nil, err
	}
	calls = make([]types.Call, 0)
	n := len(resp) / len(pools)
	for i, c := range pools {
		err := c.decodeInitCalls(resp[i*n : i*n+n])
		if err != nil {
			points[c.config.Pair] = datapoint.Point{Error: err}
			return nil, err
		}
		calls = append(calls, errutil.Must(c.createPoolTokensCall()))
	}
	// Get pool tokens from vault by given pool id
	resp, err = ethereum.MultiCall(ctx, b.client, calls, types.LatestBlockNumber)
	if err != nil {
		return nil, err
	}
	tokensMap := make(map[types.Address]struct{})
	for i, c := range pools {
		err := c.decodePoolTokensCall(resp[i])
		if err != nil {
			points[c.config.Pair] = datapoint.Point{Error: err}
			return nil, err
		}
		for _, address := range c.config.Tokens {
			tokensMap[address] = struct{}{}
		}
	}
	tokenDetails, err := b.erc20.GetSymbolAndDecimals(ctx, maps.Keys(tokensMap))
	if err != nil {
		return nil, fmt.Errorf("failed getting symbol & decimals for tokens of pool: %w", err)
	}

	for _, blockDelta := range b.blocks {
		calls = make([]types.Call, 0)
		for _, c := range pools {
			calls = append(calls, errutil.Must(c.createPoolParamsCalls())...)
		}
		resp, err = ethereum.MultiCall(ctx, b.client, calls, types.BlockNumberFromUint64(uint64(block.Int64()-blockDelta)))
		if err != nil {
			return nil, err
		}
		calls = make([]types.Call, 0)
		n = len(resp) / len(pools)
		for i, c := range pools {
			err := c.decodePoolParamsCalls(resp[i*n : i*n+n])
			if err != nil {
				points[c.config.Pair] = datapoint.Point{Error: err}
				return nil, err
			}
			calls = append(calls, errutil.Must(c.createTokenRateCacheCalls())...)
		}

		if len(calls) > 0 {
			resp, err = ethereum.MultiCall(ctx, b.client, calls, types.BlockNumberFromUint64(uint64(block.Int64()-blockDelta)))
			if err != nil {
				return nil, err
			}
			n = len(resp) / len(pools)
			for i, c := range pools {
				err := c.decodeTokenRateCacheCalls(resp[i*n : i*n+n])
				if err != nil {
					points[c.config.Pair] = datapoint.Point{Error: err}
					return nil, err
				}
			}
		}

		for _, c := range pools {
			baseToken := tokenDetails[c.config.Pair.Base]
			quoteToken := tokenDetails[c.config.Pair.Quote]
			amountIn := bn.DecFloatPoint(10).Exp(bn.DecFloatPoint(baseToken.decimals))
			amountOut, _, err := c.calcAmountOut(baseToken, quoteToken, amountIn)
			if err != nil {
				points[c.config.Pair] = datapoint.Point{Error: err}
				return nil, err
			}
			// price = amountOut / 10 ^ quoteDecimals
			price := amountOut.Div(bn.DecFloatPoint(10).Exp(bn.DecFloatPoint(quoteToken.decimals)))
			totals[c.config.Pair] = totals[c.config.Pair].Add(price)
		}
	}

	for _, pair := range pairs {
		if points[pair].Error != nil {
			continue
		}
		avgPrice := totals[pair].Div(bn.DecFloatPoint(len(b.blocks)))

		// Invert the price if inverted price
		_, baseIndex, quoteIndex, _ := b.contractAddresses.ByPair(pair)
		if baseIndex > quoteIndex {
			avgPrice = bn.DecFloatPoint(1).Div(avgPrice)
		}

		tick := value.NewTick(pair, avgPrice.SetPrec(ComposableStablePrecision), nil)
		points[pair] = datapoint.Point{
			Value: tick,
			Time:  time.Now(),
		}
	}

	return points, nil
}
