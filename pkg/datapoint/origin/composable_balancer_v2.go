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

	"github.com/chronicleprotocol/oracle-suite/pkg/datapoint"
	"github.com/chronicleprotocol/oracle-suite/pkg/datapoint/value"
	"github.com/chronicleprotocol/oracle-suite/pkg/log"
	"github.com/chronicleprotocol/oracle-suite/pkg/log/null"
	"github.com/chronicleprotocol/oracle-suite/pkg/util/bn"
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

func (b *ComposableBalancerV2) FetchDataPoints(ctx context.Context, query []any) (map[any]datapoint.Point, error) {
	pairs, ok := queryToPairs(query)
	if !ok {
		return nil, fmt.Errorf("invalid query type: %T, expected []Pair", query)
	}

	sort.Slice(pairs, func(i, j int) bool {
		return pairs[i].String() < pairs[j].String()
	})

	block, err := b.client.BlockNumber(ctx)
	if err != nil {
		return nil, fmt.Errorf("cannot get block number, %w", err)
	}

	points := make(map[any]datapoint.Point)
	totals := make(map[value.Pair]*bn.DecFloatPointNumber)

	var poolConfigs []ComposableStablePoolConfig
	for _, pair := range pairs {
		contract, _, _, err := b.contractAddresses.ByPair(pair)
		if err != nil {
			points[pair] = datapoint.Point{Error: err}
			continue
		}
		poolConfigs = append(poolConfigs, ComposableStablePoolConfig{
			Pair:    pair,
			Address: contract,
		})
		totals[pair] = bn.DecFloatPoint(0)
	}

	for _, pair := range pairs {
		if points[pair].Error != nil {
			continue
		}

		for _, blockDelta := range b.blocks {
			blockNumber := types.BlockNumberFromUint64(uint64(block.Int64() - blockDelta))
			pools, err := NewComposableStablePools(poolConfigs, b.client)
			if err != nil {
				points[pair] = datapoint.Point{Error: err}
				break
			}

			err = pools.InitializePools(ctx, blockNumber)
			if err != nil {
				points[pair] = datapoint.Point{Error: err}
				break
			}

			pool := pools.FindPoolByPair(pair)
			if pool == nil {
				points[pair] = datapoint.Point{Error: fmt.Errorf(
					"not found balancer's weighted pool: %s", pair.String())}
				break
			}

			baseToken := pools.tokenDetails[pair.Base]
			quoteToken := pools.tokenDetails[pair.Quote]
			// amountIn = 10 ^ baseDecimals
			amountIn := _powX(10, int64(baseToken.decimals))
			amountOut, _, err := pool.CalcAmountOut(baseToken.address, quoteToken.address, amountIn)
			if err != nil {
				points[pair] = datapoint.Point{Error: err}
				break
			}
			// price = amountOut / 10 ^ quoteDecimals
			price := bn.DecFloatPoint(amountOut).Div(bn.DecFloatPoint(_powX(10, int64(quoteToken.decimals))))
			totals[pair] = totals[pair].Add(price)
		}
		if points[pair].Error != nil {
			continue
		}

		avgPrice := totals[pair].Div(bn.DecFloatPoint(len(b.blocks)))

		tick := value.NewTick(pair, avgPrice, nil)
		points[pair] = datapoint.Point{
			Value: tick,
			Time:  time.Now(),
		}
	}

	return points, nil
}
