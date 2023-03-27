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

	pkgEthereum "github.com/chronicleprotocol/oracle-suite/pkg/ethereum"
)

//go:embed curve_abi.json
var curvePoolABI []byte

type CurveFinance struct {
	ethClient                 pkgEthereum.Client //nolint:staticcheck // ethereum.Client is deprecated
	addrs                     ContractAddresses
	abi                       *abi.Contract
	baseIndex, quoteIndex, dx *big.Int
	blocks                    []int64
}

//nolint:staticcheck // ethereum.Client is deprecated
func NewCurveFinance(cli pkgEthereum.Client, addrs ContractAddresses, blocks []int64) (*CurveFinance, error) {
	a, err := abi.ParseJSON(curvePoolABI)
	if err != nil {
		return nil, err
	}
	return &CurveFinance{
		ethClient:  cli,
		addrs:      addrs,
		abi:        a,
		baseIndex:  big.NewInt(0),
		quoteIndex: big.NewInt(1),
		dx:         new(big.Int).Mul(big.NewInt(1), new(big.Int).SetUint64(ether)),
		blocks:     blocks,
	}, nil
}

func (s CurveFinance) pairsToContractAddress(pair Pair) (types.Address, bool, error) {
	contract, inverted, ok := s.addrs.ByPair(pair)
	if !ok {
		return types.Address{},
			inverted,
			fmt.Errorf("failed to get contract address for pair: %s", pair.String())
	}
	return types.MustAddressFromHex(contract), inverted, nil
}

func (s CurveFinance) PullPrices(pairs []Pair) []FetchResult {
	sort.Slice(pairs, func(i, j int) bool {
		return pairs[i].String() < pairs[j].String()
	})
	var (
		frs []FetchResult
		cds []types.Call
	)
	for _, pair := range pairs {
		contract, inverted, err := s.pairsToContractAddress(pair)
		if err != nil {
			return fetchResultListWithErrors(pairs, err)
		}
		var callData []byte
		if !inverted {
			callData, err = s.abi.Methods["get_dy"].EncodeArgs(s.baseIndex, s.quoteIndex, s.dx)
		} else {
			callData, err = s.abi.Methods["get_dy"].EncodeArgs(s.quoteIndex, s.baseIndex, s.dx)
		}
		if err != nil {
			return fetchResultListWithErrors(pairs, err)
		}
		cds = append(cds, types.Call{To: &contract, Input: callData})
	}
	blockNumber, err := s.ethClient.BlockNumber(context.Background())
	if err != nil {
		return fetchResultListWithErrors(pairs, err)
	}
	resps := make([][][]byte, len(cds))
	for _, blockDelta := range s.blocks {
		ctx := pkgEthereum.WithBlockNumber(context.Background(), big.NewInt(blockNumber.Int64()-blockDelta))
		multiCallResps, err := s.ethClient.MultiCall(ctx, cds)
		if err != nil {
			return fetchResultListWithErrors(pairs, err)
		}
		if len(multiCallResps) != len(resps) {
			return fetchResultListWithErrors(
				pairs,
				fmt.Errorf(
					"multi call response length mismatch, expected %d, got %d",
					len(resps),
					len(multiCallResps),
				),
			)
		}
		for i, r := range multiCallResps {
			resps[i] = append(resps[i], r)
		}
	}
	for i, pair := range pairs {
		price, _ := reduceEtherAverageFloat(resps[i]).Float64()
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
