//  Copyright (C) 2021-2023 Chronicle Labs, Inc.
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

package e2e

import (
	"context"
	"encoding/json"
	"net/http"
	"testing"
	"time"

	"github.com/chronicleprotocol/infestor"
	"github.com/chronicleprotocol/infestor/origin"
	"github.com/chronicleprotocol/infestor/smocker"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

type spireDataPointMessage struct {
	Point spireDataPoint `json:"point"`
	Model string         `json:"model"`
}

type spirePrice struct {
	Pair  string `json:"pair"`
	Price string `json:"price"`
}

type spireDataPoint struct {
	Value spirePrice `json:"value"`
	Time  time.Time  `json:"time"`
}

func parseSpireDataPointMessage(dataPoint []byte) (spireDataPointMessage, error) {
	var p spireDataPointMessage
	err := json.Unmarshal(dataPoint, &p)
	if err != nil {
		return p, err
	}
	return p, nil
}

func Test_Ghost_ValidPrice(t *testing.T) {
	ctx, ctxCancel := context.WithTimeout(context.Background(), 2*time.Minute)
	defer ctxCancel()

	s := smocker.NewAPI("http://127.0.0.1:8081")
	require.NoError(t, s.Reset(ctx))

	err := infestor.NewMocksBuilder().
		Add(origin.NewExchange("kraken").WithSymbol("BTC/USD").WithPrice(1)).
		Add(origin.NewExchange("kraken").WithSymbol("ETH/BTC").WithPrice(1)).
		Add(origin.NewExchange("kraken").WithSymbol("ETH/USD").WithPrice(1)).
		Deploy(*s)
	require.NoError(t, err)

	spireCmd := command(ctx, "..", nil, "./spire", "agent", "-c", "./e2e/testdata/config/spire.hcl", "-v", "debug")
	ghostCmd := command(ctx, "..", nil, "./ghost", "run", "-c", "./e2e/testdata/config/ghost.hcl", "-v", "debug")
	defer func() {
		ctxCancel()
		_ = spireCmd.Wait()
		_ = ghostCmd.Wait()
	}()

	// Start spire.
	require.NoError(t, spireCmd.Start())
	waitForPort(ctx, "localhost", 30100)

	// Start ghost.
	require.NoError(t, ghostCmd.Start())
	waitForPort(ctx, "localhost", 30101)

	time.Sleep(5 * time.Second)

	btcusdMessage, err := execCommand(ctx, "..", nil, "./spire", "-c", "./e2e/testdata/config/spire.hcl", "pull", "price", "BTC/USD", "0x2D800d93B065CE011Af83f316ceF9F0d005B0AA4")
	require.NoError(t, err)

	ethusdMessage, err := execCommand(ctx, "..", nil, "./spire", "-c", "./e2e/testdata/config/spire.hcl", "pull", "price", "ETH/BTC", "0x2D800d93B065CE011Af83f316ceF9F0d005B0AA4")
	require.NoError(t, err)

	// ETHUSD price should not be available because it is missing ghost.pairs config.
	_, err = execCommand(ctx, "..", nil, "./spire", "-c", "./e2e/testdata/config/spire.hcl", "pull", "price", "ETH/USD", "0x2D800d93B065CE011Af83f316ceF9F0d005B0AA4")
	assert.Error(t, err)

	btcusdPrice, err := parseSpireDataPointMessage(btcusdMessage)
	require.NoError(t, err)

	ethusdPrice, err := parseSpireDataPointMessage(ethusdMessage)
	require.NoError(t, err)

	assert.Equal(t, "1", btcusdPrice.Point.Value.Price)
	assert.InDelta(t, time.Now().Unix(), btcusdPrice.Point.Time.Unix(), 10)
	assert.Equal(t, "BTC/USD", btcusdPrice.Point.Value.Pair)

	assert.Equal(t, "1", ethusdPrice.Point.Value.Price)
	assert.InDelta(t, time.Now().Unix(), ethusdPrice.Point.Time.Unix(), 10)
	assert.Equal(t, "ETH/BTC", ethusdPrice.Point.Value.Pair)
}

func Test_Ghost_InvalidPrice(t *testing.T) {
	ctx, ctxCancel := context.WithTimeout(context.Background(), 2*time.Minute)
	defer ctxCancel()

	s := smocker.NewAPI("http://127.0.0.1:8081")
	require.NoError(t, s.Reset(ctx))

	err := infestor.NewMocksBuilder().
		Add(origin.NewExchange("kraken").WithSymbol("BTC/USD").WithStatusCode(http.StatusConflict)).
		Add(origin.NewExchange("kraken").WithSymbol("ETH/BTC").WithStatusCode(http.StatusConflict)).
		Deploy(*s)
	require.NoError(t, err)

	spireCmd := command(ctx, "..", nil, "./spire", "agent", "-c", "./e2e/testdata/config/spire.hcl", "-v", "debug")
	ghostCmd := command(ctx, "..", nil, "./ghost", "run", "-c", "./e2e/testdata/config/ghost.hcl", "-v", "debug")
	defer func() {
		ctxCancel()
		_ = spireCmd.Wait()
		_ = ghostCmd.Wait()
	}()

	require.NoError(t, spireCmd.Start())
	waitForPort(ctx, "localhost", 30100)

	require.NoError(t, ghostCmd.Start())
	waitForPort(ctx, "localhost", 30101)

	time.Sleep(5 * time.Second)

	_, err = execCommand(ctx, "..", nil, "./spire", "-c", "./e2e/testdata/config/spire.hcl", "pull", "price", "BTC/USD", "0x2D800d93B065CE011Af83f316ceF9F0d005B0AA4")
	assert.Error(t, err)

	_, err = execCommand(ctx, "..", nil, "./spire", "-c", "./e2e/testdata/config/spire.hcl", "pull", "price", "ETH/BTC", "0x2D800d93B065CE011Af83f316ceF9F0d005B0AA4")
	assert.Error(t, err)
}
