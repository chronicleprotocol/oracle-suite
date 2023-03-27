package e2e

import (
	"context"
	"net/http"
	"testing"
	"time"

	"github.com/chronicleprotocol/infestor"
	"github.com/chronicleprotocol/infestor/origin"
	"github.com/chronicleprotocol/infestor/smocker"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func Test_Gofer_ETHBTC(t *testing.T) {
	ctx, ctxCancel := context.WithTimeout(context.Background(), 2*time.Minute)
	defer ctxCancel()

	s := smocker.NewAPI("http://127.0.0.1:8081")
	require.NoError(t, s.Reset(ctx))

	err := infestor.NewMocksBuilder().
		Add(origin.NewExchange("binance_us").WithSymbol("ETH/BTC").WithPrice(1)).
		Add(origin.NewExchange("bitstamp").WithSymbol("ETH/BTC").WithPrice(1)).
		Add(origin.NewExchange("coinbasepro").WithSymbol("ETH/BTC").WithPrice(1)).
		Add(origin.NewExchange("gemini").WithSymbol("ETH/BTC").WithPrice(1)).
		Add(origin.NewExchange("kraken").WithSymbol("ETH/BTC").WithPrice(1)).
		Deploy(*s)

	require.NoError(t, err)

	out, err := execCommand(ctx, "..", nil, "./gofer", "-c", "./e2e/testdata/config/gofer.hcl", "-v", "debug", "--norpc", "price", "ETH/BTC")
	require.NoError(t, err)

	p, err := parseGoferPrice(out)
	require.NoError(t, err)

	assert.Equal(t, "aggregator", p.Type)
	assert.Equal(t, float64(1), p.Price)
	assert.Greater(t, len(p.Prices), 0)
}

func Test_Gofer_ETHBTC_4Correct1Zero(t *testing.T) {
	ctx, ctxCancel := context.WithTimeout(context.Background(), 2*time.Minute)
	defer ctxCancel()

	s := smocker.NewAPI("http://127.0.0.1:8081")
	require.NoError(t, s.Reset(ctx))

	err := infestor.NewMocksBuilder().
		Add(origin.NewExchange("binance_us").WithSymbol("ETH/BTC").WithPrice(1)).
		Add(origin.NewExchange("bitstamp").WithSymbol("ETH/BTC").WithPrice(1)).
		Add(origin.NewExchange("coinbasepro").WithSymbol("ETH/BTC").WithPrice(1)).
		Add(origin.NewExchange("gemini").WithSymbol("ETH/BTC").WithPrice(1)).
		Add(origin.NewExchange("kraken").WithSymbol("ETH/BTC").WithPrice(0)).
		Deploy(*s)
	require.NoError(t, err)

	out, err := execCommand(ctx, "..", nil, "./gofer", "-c", "./e2e/testdata/config/gofer.hcl", "-v", "debug", "--norpc", "price", "ETH/BTC")
	require.NoError(t, err)

	p, err := parseGoferPrice(out)
	require.NoError(t, err)

	assert.Equal(t, "aggregator", p.Type)
	assert.Equal(t, float64(1), p.Price)
	assert.Greater(t, len(p.Prices), 0)
}

func Test_Gofer_ETHBTC_4Correct1Invalid(t *testing.T) {
	ctx, ctxCancel := context.WithTimeout(context.Background(), 2*time.Minute)
	defer ctxCancel()

	s := smocker.NewAPI("http://127.0.0.1:8081")
	require.NoError(t, s.Reset(ctx))

	err := infestor.NewMocksBuilder().
		Add(origin.NewExchange("binance_us").WithSymbol("ETH/BTC").WithPrice(1)).
		Add(origin.NewExchange("bitstamp").WithSymbol("ETH/BTC").WithPrice(1)).
		Add(origin.NewExchange("coinbasepro").WithSymbol("ETH/BTC").WithPrice(1)).
		Add(origin.NewExchange("gemini").WithSymbol("ETH/BTC").WithPrice(1)).
		Add(origin.NewExchange("kraken").WithSymbol("ETH/BTC").WithStatusCode(http.StatusNotFound)).
		Deploy(*s)
	require.NoError(t, err)

	out, err := execCommand(ctx, "..", nil, "./gofer", "-c", "./e2e/testdata/config/gofer.hcl", "-v", "debug", "--norpc", "price", "ETH/BTC")
	require.NoError(t, err)

	p, err := parseGoferPrice(out)
	require.NoError(t, err)

	assert.Equal(t, "aggregator", p.Type)
	assert.Equal(t, float64(1), p.Price)
	assert.Greater(t, len(p.Prices), 0)
}

func Test_Gofer_ETHBTC_3Correct2Invalid(t *testing.T) {
	ctx, ctxCancel := context.WithTimeout(context.Background(), 2*time.Minute)
	defer ctxCancel()

	s := smocker.NewAPI("http://127.0.0.1:8081")
	require.NoError(t, s.Reset(ctx))

	err := infestor.NewMocksBuilder().
		Add(origin.NewExchange("binance_us").WithSymbol("ETH/BTC").WithPrice(1)).
		Add(origin.NewExchange("bitstamp").WithSymbol("ETH/BTC").WithPrice(1)).
		Add(origin.NewExchange("coinbasepro").WithSymbol("ETH/BTC").WithPrice(1)).
		Add(origin.NewExchange("gemini").WithSymbol("ETH/BTC").WithStatusCode(http.StatusNotFound)).
		Add(origin.NewExchange("kraken").WithSymbol("ETH/BTC").WithStatusCode(http.StatusNotFound)).
		Deploy(*s)
	require.NoError(t, err)

	out, err := execCommand(ctx, "..", nil, "./gofer", "-c", "./e2e/testdata/config/gofer.hcl", "-v", "debug", "--norpc", "price", "ETH/BTC")
	require.NoError(t, err)

	p, err := parseGoferPrice(out)
	require.NoError(t, err)

	assert.Equal(t, "aggregator", p.Type)
	assert.Equal(t, float64(1), p.Price)
	assert.Greater(t, len(p.Prices), 0)
}

func Test_Gofer_ETHBTC_2Correct3Invalid(t *testing.T) {
	ctx, ctxCancel := context.WithTimeout(context.Background(), 2*time.Minute)
	defer ctxCancel()

	s := smocker.NewAPI("http://127.0.0.1:8081")
	require.NoError(t, s.Reset(ctx))

	err := infestor.NewMocksBuilder().
		Reset().
		Add(origin.NewExchange("binance_us").WithSymbol("ETH/BTC").WithPrice(1)).
		Add(origin.NewExchange("bitstamp").WithSymbol("ETH/BTC").WithPrice(1)).
		Add(origin.NewExchange("coinbasepro").WithSymbol("ETH/BTC").WithStatusCode(http.StatusNotFound)).
		Add(origin.NewExchange("gemini").WithSymbol("ETH/BTC").WithStatusCode(http.StatusNotFound)).
		Add(origin.NewExchange("kraken").WithSymbol("ETH/BTC").WithStatusCode(http.StatusNotFound)).
		Deploy(*s)
	require.NoError(t, err)

	out, err := execCommand(ctx, "..", nil, "./gofer", "-c", "./e2e/testdata/config/gofer.hcl", "-v", "debug", "--norpc", "price", "ETH/BTC")
	require.Error(t, err)

	p, err := parseGoferPrice(out)
	require.NoError(t, err)

	assert.Equal(t, "aggregator", p.Type)
	assert.NotEmpty(t, p.Error)
}
