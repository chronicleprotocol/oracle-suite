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
	"testing"

	"github.com/chronicleprotocol/oracle-suite/pkg/ethereum"
	ethereumMocks "github.com/chronicleprotocol/oracle-suite/pkg/ethereum/mocks"

	"github.com/stretchr/testify/assert"
)

func TestPostPriceHook(t *testing.T) {
	prices := make(map[Pair]*Price)

	pair, err := NewPair("RETH/ETH")
	assert.NoError(t, err)

	prices[pair] = &Price{
		Price: 1.0,
		Prices: []*Price{
			{
				Parameters: map[string]string{
					"origin": "rocketpool",
				},
				Price: 0.99,
			},
		},
	}

	contract := "0xdeadbeef"
	params := make(map[string]interface{})
	params["circuitContract"] = contract
	pairParams := NewHookParams()
	pairParams["RETH/ETH"] = params

	cli := &ethereumMocks.Client{}
	readMethodID := []byte{87, 222, 38, 164}
	divisorMethodID := []byte{31, 45, 197, 239}

	ctx := context.Background()
	cli.On("Call", ctx, ethereum.Call{Address: ethereum.HexToAddress(contract), Data: readMethodID}).Return(
		[]byte{0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 39, 16},
		nil,
	)
	cli.On("Call", ctx, ethereum.Call{Address: ethereum.HexToAddress(contract), Data: divisorMethodID}).Return(
		[]byte{0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 1, 134, 160},
		nil,
	)

	hook, err := NewPostPriceHook(context.Background(), cli, pairParams)
	assert.NoError(t, err)

	// Moderate price deviation, expect success
	err = hook.Check(prices)
	assert.NoError(t, err)
	assert.True(t, prices[pair].Error == "")

	// Extreme price deviation, expect error
	prices[pair].Prices[0].Price = 2.0
	err = hook.Check(prices)
	assert.NoError(t, err)
	assert.True(t, prices[pair].Error != "")
	t.Log(prices[pair].Error)

	// Price pair with no hooks defined succeeds
	prices = make(map[Pair]*Price)
	pair, err = NewPair("WAKA/WAKA")
	assert.NoError(t, err)

	prices[pair] = &Price{Price: 1.0}
	err = hook.Check(prices)
	assert.NoError(t, err)
	assert.True(t, prices[pair].Error == "")
}
