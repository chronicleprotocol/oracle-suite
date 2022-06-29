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

type RockerPool struct {
	ethClient ethereum.Client
	addrs     ContractAddresses
	abi       abi.ABI
	blocks    []int64
}

//go:embed rocketpool_abi.json
var rocketPoolABI string

func NewRockerPool(cli ethereum.Client, addrs ContractAddresses, blocks []int64) (*RockerPool, error) {
	a, err := abi.JSON(strings.NewReader(rocketPoolABI))
	if err != nil {
		return nil, err
	}
	return &RockerPool{
		ethClient: cli,
		addrs:     addrs,
		abi:       a,
		blocks:    blocks,
	}, nil
}

func (s RockerPool) PullPrices(pairs []Pair) []FetchResult {
	return callSinglePairOrigin(&s, pairs)
}

func (s RockerPool) callOne(pair Pair) (*Price, error) {
	contract, inverted, err := s.addrs.AddressByPair(pair)
	if err != nil {
		return nil, err
	}

	var callData []byte
	if !inverted {
		callData, err = s.abi.Pack("getExchangeRate")
	} else {
		callData, err = s.abi.Pack("getRethValue", big.NewInt(0).SetUint64(ether))
	}
	if err != nil {
		return nil, fmt.Errorf("failed to get contract args for pair: %s: %w", pair.String(), err)
	}

	resp, err := s.ethClient.CallBlocks(context.Background(), ethereum.Call{Address: contract, Data: callData}, s.blocks)
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
