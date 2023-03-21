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

package origins

import (
	"context"
	_ "embed"
	"fmt"
	"math/big"
	"sort"
	"time"

	"github.com/defiweb/go-eth/abi"
	"github.com/defiweb/go-eth/types"

	"github.com/chronicleprotocol/oracle-suite/pkg/ethereum"
)

// The three values that can be queried:
//
// - PAIR_PRICE: the price of the tokens in the Pool, expressed as the price of the second token in units of the
//   first token. For example, if token A is worth $2, and token B is worth $4, the pair price will be 2.0.
//   Note that the price is computed *including* the tokens decimals. This means that the pair price of a Pool with
//   DAI and USDC will be close to 1.0, despite DAI having 18 decimals and USDC 6.
//
// - BPT_PRICE: the price of the Pool share token (BPT), in units of the first token.
//   Note that the price is computed *including* the tokens decimals. This means that the BPT price of a Pool with
//   USDC in which BPT is worth $5 will be 5.0, despite the BPT having 18 decimals and USDC 6.
//
// - INVARIANT: the value of the Pool's invariant, which serves as a measure of its liquidity.
// enum Variable { PAIR_PRICE, BPT_PRICE, INVARIANT }

const prefixRef = "Ref:"

//go:embed balancerv2_abi.json
var balancerV2PoolABI []byte

type BalancerV2 struct {
	ethClient         ethereum.Client
	ContractAddresses ContractAddresses
	abi               *abi.Contract
	variable          byte
	blocks            []int64
}

func NewBalancerV2(ethClient ethereum.Client, addrs ContractAddresses, blocks []int64) (*BalancerV2, error) {
	a, err := abi.ParseJSON(balancerV2PoolABI)
	if err != nil {
		return nil, err
	}
	return &BalancerV2{
		ethClient:         ethClient,
		ContractAddresses: addrs,
		abi:               a,
		variable:          0, // PAIR_PRICE
		blocks:            blocks,
	}, nil
}

func (s BalancerV2) PullPrices(pairs []Pair) []FetchResult {
	sort.Slice(pairs, func(i, j int) bool {
		return pairs[i].String() < pairs[j].String()
	})
	var frs []FetchResult
	var cds = make([][]types.Call, len(pairs))
	// Prepare list of calls.
	for n, pair := range pairs {
		contract, indirect, err := s.ContractAddresses.AddressByPair(pair)
		if err != nil {
			return fetchResultListWithErrors(pairs, err)
		}
		if indirect {
			return fetchResultListWithErrors(
				pairs,
				fmt.Errorf("cannot use indirect pair to retrieve price: %s", pair.String()),
			)
		}
		callData, err := s.abi.Methods["getLatest"].EncodeArgs(s.variable) // TODO: test
		if err != nil {
			return fetchResultListWithErrors(
				pairs,
				fmt.Errorf("failed to pack contract args for getLatest (pair %s): %w", pair.String(), err),
			)
		}
		cds[n] = append(cds[n], types.Call{
			To:    &contract,
			Input: callData,
		})
		if err != nil {
			return fetchResultListWithErrors(pairs, err)
		}
		token, inverted, ok := s.ContractAddresses.ByPair(Pair{Base: prefixRef + pair.Base, Quote: pair.Quote})
		if ok {
			if inverted {
				return fetchResultListWithErrors(
					pairs,
					fmt.Errorf("cannot use inverted pair to retrieve price: %s", pair.String()),
				)
			}
			callData, err := s.abi.Methods["getPriceRateCache"].EncodeArgs(types.MustAddressFromHex(token))
			if err != nil {
				return fetchResultListWithErrors(
					pairs,
					fmt.Errorf(
						"failed to pack contract args for getPriceRateCache (pair %s): %w",
						pair.String(),
						err,
					),
				)
			}
			cds[n] = append(cds[n], types.Call{
				To:    &contract,
				Input: callData,
			})
		}
	}
	// Execute calls.
	resps, err := nestedMultiCall(s.ethClient, cds, s.blocks)
	if err != nil {
		return fetchResultListWithErrors(pairs, err)
	}
	// Calculate prices.
	for n, pair := range pairs {
		priceFloat := reduceEtherAverageFloat(resps[n][0])
		// If there are two calls, the second one is the reference price for the pair.
		// We need to multiply the pair price by the reference price to get the final price.
		if len(resps[n]) > 1 {
			priceFloat = new(big.Float).Mul(reduceEtherAverageFloat(resps[n][1]), priceFloat)
		}
		price, _ := priceFloat.Float64()
		frs = append(frs, FetchResult{
			Price: Price{
				Pair:      pair,
				Price:     price,
				Timestamp: time.Now(),
			},
		})
	}
	return frs
}

// nestedMultiCall performs a list of calls in a single multicall for each
// block delta defined in the blocks argument. The returned slice have the
// same structure as the calls slice. In the last array dimension, the
// returned slice contains the results of the multicall for each block delta.
//nolint:staticcheck // deprecated ethereum.Client
func nestedMultiCall(ethClient ethereum.Client, calls [][]types.Call, blocks []int64) ([][][][]byte, error) {
	block, err := ethClient.BlockNumber(context.Background())
	if err != nil {
		return nil, err
	}
	rs := make([][][][]byte, len(calls))
	for i, call := range calls {
		rs[i] = make([][][]byte, len(call))
	}
	// Perform multicall for each block delta.
	for _, blockDelta := range blocks {
		// Perform multicall.
		cs := make([]types.Call, 0, len(calls))
		for _, call := range calls {
			cs = append(cs, call...)
		}
		mrs, err := ethClient.MultiCall(
			ethereum.WithBlockNumber(context.Background(), big.NewInt(block.Int64()-blockDelta)),
			cs,
		)
		if err != nil {
			return nil, err
		}
		if len(mrs) != len(cs) {
			return nil, fmt.Errorf(
				"unexpected number of multicall results, expected %d, got %d", len(cs), len(mrs),
			)
		}
		// Recreate the original structure of the calls slice.
		n := 0
		for i, call := range calls {
			for j := range call {
				rs[i][j] = append(rs[i][j], mrs[n])
				n++
			}
		}
	}
	return rs, nil
}
