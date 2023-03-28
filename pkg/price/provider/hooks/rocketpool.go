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

	"github.com/defiweb/go-eth/abi"
	"github.com/defiweb/go-eth/types"

	ethereumConfig "github.com/chronicleprotocol/oracle-suite/pkg/config/ethereum"
	"github.com/chronicleprotocol/oracle-suite/pkg/ethereum/geth"
)

//go:embed circuit_abi.json
var circuitABI []byte

const circuitContractKey = "circuit_contract"

type RocketPoolCircuitBreaker struct {
	client     *geth.Client //nolint:staticcheck // ethereum.Client is deprecated
	contract   types.Address
	circuitABI *abi.Contract
}

func NewRocketPoolCircuitBreaker(
	clients ethereumConfig.ClientRegistry,
	params map[string]interface{},
) (
	*RocketPoolCircuitBreaker,
	error,
) {

	var ret RocketPoolCircuitBreaker
	if _, ok := params[circuitContractKey]; !ok {
		return nil, fmt.Errorf("post price hook failed, could not find %s parameter for RETH", circuitContractKey)
	}

	clientName, ok := params["ethereum_client"].(string)
	if !ok {
		return nil, fmt.Errorf("ethereum_client parameter not found for hook")
	}
	client, ok := clients[clientName]
	if !ok {
		return nil, fmt.Errorf("no ethereum client found for hook")
	}
	ret.client = geth.NewClient(client) //nolint:staticcheck // ethereum.Client is deprecated

	c, ok := params[circuitContractKey].(string)
	if !ok {
		return nil, fmt.Errorf("post price hook failed for RETH incorrect params: %s", c)
	}
	ret.contract = types.MustAddressFromHex(c)

	cabi, err := abi.ParseJSON(circuitABI)
	if err != nil {
		return nil, err
	}
	ret.circuitABI = cabi

	return &ret, nil
}

func (o *RocketPoolCircuitBreaker) Check(ctx context.Context, medianPrice, refPrice float64) error {
	// read()
	callData, err := o.circuitABI.Methods["read"].EncodeArgs()
	if err != nil {
		return fmt.Errorf("failed to get contract args for %s: %w", circuitContractKey, err)
	}
	response, err := o.client.Call(ctx, types.Call{To: &o.contract, Input: callData})
	if err != nil {
		return err
	}
	val := new(big.Float).SetInt(new(big.Int).SetBytes(response))
	// divisor()
	callData, err = o.circuitABI.Methods["divisor"].EncodeArgs()
	if err != nil {
		return fmt.Errorf("failed to get contract args for %s: %w", circuitContractKey, err)
	}
	response, err = o.client.Call(ctx, types.Call{To: &o.contract, Input: callData})
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
