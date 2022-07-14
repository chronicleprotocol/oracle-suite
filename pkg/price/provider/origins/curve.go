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
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/params"

	pkgEthereum "github.com/chronicleprotocol/oracle-suite/pkg/ethereum"
)

//go:embed curve_abi.json
var curvePoolABI string

type CurveFinance struct {
	ethClient                 pkgEthereum.Client
	addrs                     ContractAddresses
	abi                       abi.ABI
	baseIndex, quoteIndex, dx *big.Int
	blocks                    []int64
}

func NewCurveFinance(cli pkgEthereum.Client, addrs ContractAddresses, blocks []int64) (*CurveFinance, error) {
	a, err := abi.JSON(strings.NewReader(curvePoolABI))
	if err != nil {
		return nil, err
	}
	return &CurveFinance{
		ethClient:  cli,
		addrs:      addrs,
		abi:        a,
		baseIndex:  big.NewInt(0),
		quoteIndex: big.NewInt(1),
		dx:         new(big.Int).Mul(big.NewInt(1), big.NewInt(params.Ether)),
		blocks:     blocks,
	}, nil
}

func (s CurveFinance) pairsToContractAddress(pair Pair) (common.Address, bool, error) {
	contract, inverted, ok := s.addrs.ByPair(pair)
	if !ok {
		return common.Address{}, inverted, fmt.Errorf("failed to get contract address for pair: %s", pair.String())
	}
	return common.HexToAddress(contract), inverted, nil
}

func (s CurveFinance) PullPrices(pairs []Pair) []FetchResult {
	return callSinglePairOrigin(&s, pairs)
}

func (s CurveFinance) callOne(pair Pair) (*Price, error) {
	contract, inverted, err := s.pairsToContractAddress(pair)
	if err != nil {
		return nil, err
	}

	var callData []byte
	if !inverted {
		callData, err = s.abi.Pack("get_dy", s.baseIndex, s.quoteIndex, s.dx)
	} else {
		callData, err = s.abi.Pack("get_dy", s.quoteIndex, s.baseIndex, s.dx)
	}
	if err != nil {
		return nil, fmt.Errorf("failed to pack contract args for pair: %s", pair.String())
	}

	resp, err := s.ethClient.CallBlocks(
		context.Background(),
		pkgEthereum.Call{Address: contract, Data: callData},
		s.blocks,
	)
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
