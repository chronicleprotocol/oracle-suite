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
	"strings"
	"time"

	"github.com/ethereum/go-ethereum/accounts/abi"

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
var balancerV2PoolABI string

type BalancerV2 struct {
	ethClient         ethereum.Client
	ContractAddresses ContractAddresses
	abi               abi.ABI
	variable          byte
	blocks            []int64
}

func NewBalancerV2(ethClient ethereum.Client, addrs ContractAddresses, blocks []int64) (*BalancerV2, error) {
	a, err := abi.JSON(strings.NewReader(balancerV2PoolABI))
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
	return callSinglePairOrigin(&s, pairs)
}

func (s BalancerV2) callOne(pair Pair) (*Price, error) {
	contract, inverted, err := s.ContractAddresses.AddressByPair(pair)
	if err != nil {
		return nil, err
	}
	if inverted {
		return nil, fmt.Errorf("cannot use inverted pair to retrieve price: %s", pair.String())
	}

	var priceFloat *big.Float
	{
		callData, err := s.abi.Pack("getLatest", s.variable)
		if err != nil {
			return nil, fmt.Errorf("failed to pack contract args for getLatest (pair %s): %w", pair.String(), err)
		}

		resp, err := s.ethClient.CallBlocks(context.Background(), ethereum.Call{Address: contract, Data: callData}, s.blocks)
		if err != nil {
			return nil, err
		}
		priceFloat = reduceEtherAverageFloat(resp)
	}

	token, inverted, ok := s.ContractAddresses.ByPair(Pair{Base: prefixRef + pair.Base, Quote: pair.Quote})
	if ok && !inverted {
		callData, err := s.abi.Pack("getPriceRateCache", ethereum.HexToAddress(token))
		if err != nil {
			return nil, fmt.Errorf("failed to pack contract args for getPriceRateCache (pair %s): %w", pair.String(), err)
		}
		resp, err := s.ethClient.CallBlocks(context.Background(), ethereum.Call{Address: contract, Data: callData}, s.blocks)
		if err != nil {
			return nil, err
		}
		priceFloat = new(big.Float).Mul(reduceEtherAverageFloat(resp), priceFloat)
	}

	price, _ := priceFloat.Float64()
	return &Price{
		Pair:      pair,
		Price:     price,
		Timestamp: time.Now(),
	}, nil
}
