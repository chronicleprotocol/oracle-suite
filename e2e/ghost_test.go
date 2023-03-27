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

type spirePriceMessage struct {
	Price spirePrice      `json:"price"`
	Trace json.RawMessage `json:"trace"`
}

type spirePrice struct {
	Wat string `json:"wat"`
	Val string `json:"val"`
	Age int64  `json:"age"`
	V   string `json:"v"`
	R   string `json:"r"`
	S   string `json:"s"`
}

func parseSpirePriceMessage(price []byte) (spirePriceMessage, error) {
	var p spirePriceMessage
	err := json.Unmarshal(price, &p)
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
	ghostCmd := command(ctx, "..", nil, "./ghost", "run", "-c", "./e2e/testdata/config/ghost.hcl", "--gofer.norpc", "-v", "debug")
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

	btcusdMessage, err := execCommand(ctx, "..", nil, "./spire", "-c", "./e2e/testdata/config/spire.hcl", "pull", "price", "BTCUSD", "0x2D800d93B065CE011Af83f316ceF9F0d005B0AA4")
	require.NoError(t, err)

	ethusdMessage, err := execCommand(ctx, "..", nil, "./spire", "-c", "./e2e/testdata/config/spire.hcl", "pull", "price", "ETHBTC", "0x2D800d93B065CE011Af83f316ceF9F0d005B0AA4")
	require.NoError(t, err)

	// ETHUSD price should not be available because it is missing ghost.pairs config.
	_, err = execCommand(ctx, "..", nil, "./spire", "-c", "./e2e/testdata/config/spire.hcl", "pull", "price", "ETHUSD", "0x2D800d93B065CE011Af83f316ceF9F0d005B0AA4")
	assert.Error(t, err)

	btcusdPrice, err := parseSpirePriceMessage(btcusdMessage)
	require.NoError(t, err)

	ethusdPrice, err := parseSpirePriceMessage(ethusdMessage)
	require.NoError(t, err)

	assert.Equal(t, "1000000000000000000", btcusdPrice.Price.Val)
	assert.InDelta(t, time.Now().Unix(), btcusdPrice.Price.Age, 10)
	assert.Equal(t, "BTCUSD", btcusdPrice.Price.Wat)

	assert.Equal(t, "1000000000000000000", ethusdPrice.Price.Val)
	assert.InDelta(t, time.Now().Unix(), ethusdPrice.Price.Age, 10)
	assert.Equal(t, "ETHBTC", ethusdPrice.Price.Wat)
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
	ghostCmd := command(ctx, "..", nil, "./ghost", "run", "-c", "./e2e/testdata/config/ghost.hcl", "--gofer.norpc", "-v", "debug")
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

	_, err = execCommand(ctx, "..", nil, "./spire", "-c", "./e2e/testdata/config/spire.hcl", "pull", "price", "BTCUSD", "0x2D800d93B065CE011Af83f316ceF9F0d005B0AA4")
	assert.Error(t, err)

	_, err = execCommand(ctx, "..", nil, "./spire", "-c", "./e2e/testdata/config/spire.hcl", "pull", "price", "ETHBTC", "0x2D800d93B065CE011Af83f316ceF9F0d005B0AA4")
	assert.Error(t, err)
}
