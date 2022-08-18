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

package hooks

import (
	"context"
	_ "embed"
	"fmt"
	"math"
	"math/big"
	"strings"

	"github.com/chronicleprotocol/oracle-suite/pkg/ethereum"

	"github.com/ethereum/go-ethereum/accounts/abi"
	"github.com/ethereum/go-ethereum/common"
)

//go:embed circuit_abi.json
var circuitABI string

const circuitContractKey = "circuitContract"

type RocketPoolCircuitBreaker struct {
	contract   common.Address
	circuitABI abi.ABI
}

func NewRocketPoolCircuitBreaker(params map[string]interface{}) (*RocketPoolCircuitBreaker, error) {
	var ret RocketPoolCircuitBreaker
	if _, ok := params[circuitContractKey]; !ok {
		return nil, fmt.Errorf("post price hook failed, could not find %s parameter for RETH", circuitContractKey)
	}

	c, ok := params[circuitContractKey].(string)
	if !ok {
		return nil, fmt.Errorf("post price hook failed for RETH incorrect params: %s", c)
	}
	ret.contract = ethereum.HexToAddress(c)

	cabi, err := abi.JSON(strings.NewReader(circuitABI))
	if err != nil {
		return nil, err
	}
	ret.circuitABI = cabi

	return &ret, nil
}

func (o *RocketPoolCircuitBreaker) Check(ctx context.Context,
	cli ethereum.Client, medianPrice, refPrice float64) error {

	// read()
	callData, err := o.circuitABI.Pack("read")
	if err != nil {
		return fmt.Errorf("failed to get contract args for %s: %w", circuitContractKey, err)
	}
	response, err := cli.Call(ctx, ethereum.Call{Address: o.contract, Data: callData})
	if err != nil {
		return err
	}
	val := new(big.Float).SetInt(new(big.Int).SetBytes(response))
	// divisor()
	callData, err = o.circuitABI.Pack("divisor")
	if err != nil {
		return fmt.Errorf("failed to get contract args for %s: %w", circuitContractKey, err)
	}
	response, err = cli.Call(ctx, ethereum.Call{Address: o.contract, Data: callData})
	if err != nil {
		return err
	}
	div := new(big.Float).SetInt(new(big.Int).SetBytes(response))

	v, _ := val.Float64()
	d, _ := div.Float64()
	breaker := v / d

	deviation := math.Abs(1.0 - (refPrice / medianPrice))

	if deviation > breaker {
		return fmt.Errorf("error rETH circuit breaker tripped: %f > %f", deviation, breaker)
	}
	return nil
}
