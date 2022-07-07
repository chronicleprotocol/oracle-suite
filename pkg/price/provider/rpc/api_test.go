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

package rpc

import (
	"context"
	"os"
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/chronicleprotocol/oracle-suite/pkg/price/provider"
	"github.com/chronicleprotocol/oracle-suite/pkg/price/provider/mocks"

	"github.com/chronicleprotocol/oracle-suite/pkg/log/null"
)

var (
	agent     *Agent
	mockGofer *mocks.Provider
	rpcGofer  *Provider
)

func TestMain(m *testing.M) {
	ctx, ctxCancel := context.WithCancel(context.Background())
	var err error

	mockGofer = &mocks.Provider{}
	agent, err = NewAgent(AgentConfig{
		Provider: mockGofer,
		Network:  "tcp",
		Address:  "127.0.0.1:0",
		Logger:   null.New(),
	})
	if err != nil {
		panic(err)
	}
	if err = agent.Start(ctx); err != nil {
		panic(err)
	}
	rpcGofer, err = NewProvider("tcp", agent.listener.Addr().String())
	if err != nil {
		panic(err)
	}
	err = rpcGofer.Start(ctx)
	if err != nil {
		panic(err)
	}

	retCode := m.Run()
	ctxCancel()
	os.Exit(retCode)
}

func TestClient_Models(t *testing.T) {
	pair := provider.Pair{Base: "A", Quote: "B"}
	model := map[provider.Pair]*provider.Model{pair: {Type: "test"}}

	mockGofer.On("Models", pair).Return(model, nil)
	resp, err := rpcGofer.Models(pair)

	assert.Equal(t, model, resp)
	assert.NoError(t, err)
}

func TestClient_Price(t *testing.T) {
	pair := provider.Pair{Base: "A", Quote: "B"}
	prices := map[provider.Pair]*provider.Price{pair: {Type: "test"}}

	mockGofer.On("Prices", pair).Return(prices, nil)
	resp, err := rpcGofer.Price(pair)

	assert.Equal(t, prices[pair], resp)
	assert.NoError(t, err)
}

func TestClient_Prices(t *testing.T) {
	pair := provider.Pair{Base: "A", Quote: "B"}
	prices := map[provider.Pair]*provider.Price{pair: {Type: "test"}}

	mockGofer.On("Prices", pair).Return(prices, nil)
	resp, err := rpcGofer.Prices(pair)

	assert.Equal(t, prices, resp)
	assert.NoError(t, err)
}

func TestClient_Pairs(t *testing.T) {
	pairs := []provider.Pair{{Base: "A", Quote: "B"}}

	mockGofer.On("Pairs").Return(pairs, nil)
	resp, err := rpcGofer.Pairs()

	assert.Equal(t, pairs, resp)
	assert.NoError(t, err)
}
