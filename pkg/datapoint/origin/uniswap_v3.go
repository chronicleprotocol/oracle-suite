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
	"math/big"
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
)

const UniswapV3LoggerTag = "UNISWAPV3_ORIGIN"

type UniswapV3Config struct {
	Client            rpc.RPC
	ContractAddresses ContractAddresses
	Logger            log.Logger
	Blocks            []int64
}

type UniswapV3 struct {
	client            rpc.RPC
	contractAddresses ContractAddresses
	erc20             *ERC20
	blocks            []int64
	logger            log.Logger
}

func NewUniswapV3(config UniswapV3Config) (*UniswapV3, error) {
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

	return &UniswapV3{
		client:            config.Client,
		contractAddresses: config.ContractAddresses,
		erc20:             erc20,
		blocks:            config.Blocks,
		logger:            config.Logger.WithField("uniswapV3", UniswapV3LoggerTag),
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
		contract, _, _, err := u.contractAddresses.ByPair(pair)
		if err != nil {
			points[pair] = datapoint.Point{Error: err}
			continue
		}

		// Calls for `slot0`
		callData, err := slot0.EncodeArgs()
		if err != nil {
			points[pair] = datapoint.Point{Error: fmt.Errorf("failed to get slot0 for pair: %s: %w",
				pair.String(), err)}
			continue
		}
		calls = append(calls, types.Call{
			To:    &contract,
			Input: callData,
		})
		// Calls for `token0`
		callData, err = token0Abi.EncodeArgs()
		if err != nil {
			points[pair] = datapoint.Point{Error: fmt.Errorf("failed to get token0 for pair: %s: %w",
				pair.String(), err)}
			continue
		}
		callsToken = append(callsToken, types.Call{
			To:    &contract,
			Input: callData,
		})
		// Calls for `token1`
		callData, err = token1Abi.EncodeArgs()
		if err != nil {
			points[pair] = datapoint.Point{Error: fmt.Errorf("failed to get token1 for pair: %s: %w",
				pair.String(), err)}
			continue
		}
		callsToken = append(callsToken, types.Call{
			To:    &contract,
			Input: callData,
		})

		totals[i] = new(big.Float).SetInt64(0)
	}

	// Get decimals for all the tokens
	tokensMap := make(map[types.Address]struct{})
	var tokenDetails map[string]ERC20Details
	if len(callsToken) > 0 {
		resp, err := ethereum.MultiCall(ctx, u.client, callsToken, types.LatestBlockNumber)
		if err != nil {
			return nil, err
		}

		for i := range resp {
			var address types.Address
			if err := token0Abi.DecodeValues(resp[i], &address); err != nil {
				return nil, fmt.Errorf("failed decoding token address of pool: %w", err)
			}
			tokensMap[address] = struct{}{}
		}
		tokenDetails, err = u.erc20.GetSymbolAndDecimals(ctx, maps.Keys(tokensMap))
		if err != nil {
			return nil, fmt.Errorf("failed getting symbol & decimals for tokens of pool: %w", err)
		}
	}

	// 2 ^ 192
	if len(calls) > 0 {
		const x192 = 192
		q192 := new(big.Int).Exp(big.NewInt(2), big.NewInt(x192), nil)
		for _, blockDelta := range u.blocks {
			resp, err := ethereum.MultiCall(ctx, u.client, calls, types.BlockNumberFromUint64(uint64(block.Int64()-blockDelta)))
			if err != nil {
				return nil, err
			}

			n := 0
			for i, pair := range pairs {
				if points[pair].Error != nil {
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

				sqrtRatioX96 := new(big.Int).SetBytes(resp[n][0:32])
				// ratioX192 = sqrtRatioX96 ^ 2
				ratioX192 := new(big.Int).Mul(sqrtRatioX96, sqrtRatioX96)

				baseToken := tokenDetails[pair.Base]
				quoteToken := tokenDetails[pair.Quote]
				// baseAmount = 10 ^ baseDecimals
				baseAmount := new(big.Int).Exp(big.NewInt(10), big.NewInt(int64(baseToken.decimals)), nil)

				// Reference: https://github.com/Uniswap/v3-periphery/blob/main/contracts/libraries/OracleLibrary.sol#L60
				// Reference: https://github.com/Uniswap/v3-subgraph/blob/main/src/utils/pricing.ts#L48
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
				n++
			}
		}
	}

	for i, pair := range pairs {
		if points[pair].Error != nil {
			continue
		}

		avgPrice := new(big.Float).Quo(totals[i], new(big.Float).SetUint64(uint64(len(u.blocks))))

		tick := value.NewTick(pair, avgPrice, nil)
		points[pair] = datapoint.Point{
			Value: tick,
			Time:  time.Now(),
		}
	}

	if len(pairs) == 1 && points[pairs[0]].Error != nil {
		return points, points[pairs[0]].Error
	}
	return points, nil
}
