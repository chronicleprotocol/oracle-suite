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

package provider

import (
	"context"
	"fmt"

	"github.com/chronicleprotocol/oracle-suite/pkg/ethereum"
	"github.com/chronicleprotocol/oracle-suite/pkg/price/provider/hooks"
)

type PostPriceHook struct {
	cli      ethereum.Client
	ctx      context.Context
	handlers map[string]interface{}
}

const RocketPoolPair = "RETH/ETH"

type HookParams map[string]map[string]interface{}

func NewHookParams() HookParams {
	return make(HookParams)
}

func NewPostPriceHook(ctx context.Context, cli ethereum.Client, params HookParams) (*PostPriceHook, error) {
	handlers := make(map[string]interface{})
	for k, v := range params {
		switch k {
		case RocketPoolPair:
			h, err := hooks.NewRocketPoolCircuitBreaker(v)
			if err != nil {
				return nil, err
			}
			handlers[k] = h
		default:
		}
	}

	return &PostPriceHook{
		cli:      cli,
		ctx:      ctx,
		handlers: handlers,
	}, nil
}

func (o *PostPriceHook) Check(prices map[Pair]*Price) error {
	for pair, price := range prices {
		switch pair.String() {
		case RocketPoolPair:
			if _, ok := o.handlers[RocketPoolPair]; !ok {
				return fmt.Errorf("no post price hook handler found for %s", RocketPoolPair)
			}

			var refPrice float64
			for _, origin := range price.Prices {
				if origin.Parameters["origin"] == "rocketpool" {
					refPrice = origin.Price
					break
				}
			}
			if refPrice == 0 {
				prices[pair].Error = fmt.Sprintf("post price hook failed for %s, reference price should be > 0", pair.String())
				return nil
			}

			err := o.handlers[RocketPoolPair].(*hooks.RocketPoolCircuitBreaker).Check(o.ctx, o.cli, price.Price, refPrice)
			if err != nil {
				prices[pair].Error = err.Error()
			}
			return nil
		default:
		}
	}
	return nil
}
