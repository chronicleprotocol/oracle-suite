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
	"time"

	"github.com/defiweb/go-eth/abi"
	"github.com/defiweb/go-eth/types"

	"github.com/chronicleprotocol/oracle-suite/pkg/ethereum"
)

//nolint:staticcheck // deprecated ethereum.Client
type RocketPool struct {
	ethClient ethereum.Client
	addrs     ContractAddresses
	abi       *abi.Contract
	blocks    []int64
}

//go:embed rocketpool_abi.json
var rocketPoolABI []byte

func NewRocketPool(cli ethereum.Client, addrs ContractAddresses, blocks []int64) (*RocketPool, error) {
	a, err := abi.ParseJSON(rocketPoolABI)
	if err != nil {
		return nil, err
	}
	return &RocketPool{
		ethClient: cli,
		addrs:     addrs,
		abi:       a,
		blocks:    blocks,
	}, nil
}

func (s RocketPool) PullPrices(pairs []Pair) []FetchResult {
	return callSinglePairOrigin(&s, pairs)
}

func (s RocketPool) callOne(pair Pair) (*Price, error) {
	contract, inverted, err := s.addrs.AddressByPair(pair)
	if err != nil {
		return nil, err
	}

	var callData []byte
	if !inverted {
		callData, err = s.abi.Methods["getExchangeRate"].EncodeArgs()
	} else {
		callData, err = s.abi.Methods["getRethValue"].EncodeArgs(big.NewInt(0).SetUint64(ether))
	}
	if err != nil {
		return nil, fmt.Errorf("failed to get contract args for pair: %s: %w", pair.String(), err)
	}

	resp, err := s.ethClient.CallBlocks(context.Background(), types.Call{To: &contract, Input: callData}, s.blocks)
	if err != nil {
		return nil, err
	}
	price, _ := reduceEtherAverageFloat(resp).Float64()

	return &Price{
		Pair:      pair,
		Price:     price,
		Timestamp: time.Now(),
	}, nil
}
