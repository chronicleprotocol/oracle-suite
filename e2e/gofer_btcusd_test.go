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

func Test_Gofer_BTCUSD(t *testing.T) {
	ctx, ctxCancel := context.WithTimeout(context.Background(), 2*time.Minute)
	defer ctxCancel()

	s := smocker.NewAPI("http://127.0.0.1:8081")
	require.NoError(t, s.Reset(ctx))

	err := infestor.NewMocksBuilder().
		Debug().
		Add(origin.NewExchange("binance").WithSymbol("BTC/USD").WithPrice(1)).
		Add(origin.NewExchange("bitstamp").WithSymbol("BTC/USD").WithPrice(1)).
		Add(origin.NewExchange("coinbase").WithSymbol("BTC/USD").WithPrice(1)).
		Add(origin.NewExchange("gemini").WithSymbol("BTC/USD").WithPrice(1)).
		Add(origin.NewExchange("kraken").WithSymbol("BTC/USD").WithPrice(1)).
		Deploy(*s)
	require.NoError(t, err)

	out, err := execCommand(ctx, "..", nil, "./gofer", "-c", "./e2e/testdata/config/gofer.hcl", "-v", "debug", "price", "BTC/USD", "--format", "json")
	require.NoError(t, err)

	priceMap, err := parseGoferPrice(out)
	require.NoError(t, err)
	p := priceMap["BTC/USD"]

	assert.Equal(t, "reference", p.Meta["type"])
	assert.Equal(t, float64(1), p.Value.Price)
	assert.Greater(t, len(p.SubPoints), 0)
}

func Test_Gofer_BTCUSD_4Correct1Zero(t *testing.T) {
	ctx, ctxCancel := context.WithTimeout(context.Background(), 2*time.Minute)
	defer ctxCancel()

	s := smocker.NewAPI("http://127.0.0.1:8081")
	err := s.Reset(ctx)
	require.NoError(t, err)

	err = infestor.NewMocksBuilder().
		Add(origin.NewExchange("bitstamp").WithSymbol("BTC/USD").WithPrice(1)).
		Add(origin.NewExchange("binance").WithSymbol("BTC/USD").WithPrice(1)).
		Add(origin.NewExchange("coinbase").WithSymbol("BTC/USD").WithPrice(1)).
		Add(origin.NewExchange("gemini").WithSymbol("BTC/USD").WithPrice(1)).
		Add(origin.NewExchange("kraken").WithSymbol("BTC/USD").WithPrice(0)).
		Deploy(*s)
	require.NoError(t, err)

	out, err := execCommand(ctx, "..", nil, "./gofer", "-c", "./e2e/testdata/config/gofer.hcl", "-v", "debug", "price", "BTC/USD", "--format", "json")
	require.NoError(t, err)

	priceMap, err := parseGoferPrice(out)
	require.NoError(t, err)
	p := priceMap["BTC/USD"]

	assert.Equal(t, "reference", p.Meta["type"])
	assert.Equal(t, float64(1), p.Value.Price)
	assert.Greater(t, len(p.SubPoints), 0)
}

func Test_Gofer_BTCUSD_4Correct1Invalid(t *testing.T) {
	ctx, ctxCancel := context.WithTimeout(context.Background(), 2*time.Minute)
	defer ctxCancel()

	s := smocker.NewAPI("http://127.0.0.1:8081")
	require.NoError(t, s.Reset(ctx))

	err := infestor.NewMocksBuilder().
		Add(origin.NewExchange("bitstamp").WithSymbol("BTC/USD").WithPrice(1)).
		Add(origin.NewExchange("binance").WithSymbol("BTC/USD").WithPrice(1)).
		Add(origin.NewExchange("coinbase").WithSymbol("BTC/USD").WithPrice(1)).
		Add(origin.NewExchange("gemini").WithSymbol("BTC/USD").WithPrice(1)).
		Add(origin.NewExchange("kraken").WithSymbol("BTC/USD").WithStatusCode(http.StatusNotFound)).
		Deploy(*s)
	require.NoError(t, err)

	out, err := execCommand(ctx, "..", nil, "./gofer", "-c", "./e2e/testdata/config/gofer.hcl", "-v", "debug", "price", "BTC/USD", "--format", "json")
	require.NoError(t, err)

	priceMap, err := parseGoferPrice(out)
	require.NoError(t, err)
	p := priceMap["BTC/USD"]

	assert.Equal(t, "reference", p.Meta["type"])
	assert.Equal(t, float64(1), p.Value.Price)
	assert.Greater(t, len(p.SubPoints), 0)
}

func Test_Gofer_BTCUSD_3Correct2Invalid(t *testing.T) {
	ctx, ctxCancel := context.WithTimeout(context.Background(), 2*time.Minute)
	defer ctxCancel()

	s := smocker.NewAPI("http://127.0.0.1:8081")
	require.NoError(t, s.Reset(ctx))

	err := infestor.NewMocksBuilder().
		Reset().
		Add(origin.NewExchange("bitstamp").WithSymbol("BTC/USD").WithPrice(1)).
		Add(origin.NewExchange("binance").WithSymbol("BTC/USD").WithPrice(1)).
		Add(origin.NewExchange("coinbase").WithSymbol("BTC/USD").WithPrice(1)).
		Add(origin.NewExchange("gemini").WithSymbol("BTC/USD").WithStatusCode(http.StatusNotFound)).
		Add(origin.NewExchange("kraken").WithSymbol("BTC/USD").WithStatusCode(http.StatusNotFound)).
		Deploy(*s)
	require.NoError(t, err)

	out, err := execCommand(ctx, "..", nil, "./gofer", "-c", "./e2e/testdata/config/gofer.hcl", "-v", "debug", "price", "BTC/USD", "--format", "json")
	require.NoError(t, err)

	priceMap, err := parseGoferPrice(out)
	require.NoError(t, err)
	p := priceMap["BTC/USD"]

	assert.Equal(t, "reference", p.Meta["type"])
	assert.Equal(t, float64(1), p.Value.Price)
	assert.Greater(t, len(p.SubPoints), 0)
}

func Test_Gofer_BTCUSD_2Correct3Invalid(t *testing.T) {
	ctx, ctxCancel := context.WithTimeout(context.Background(), 2*time.Minute)
	defer ctxCancel()

	s := smocker.NewAPI("http://127.0.0.1:8081")
	require.NoError(t, s.Reset(ctx))

	err := infestor.NewMocksBuilder().
		Add(origin.NewExchange("bitstamp").WithSymbol("BTC/USD").WithPrice(1)).
		Add(origin.NewExchange("binance").WithSymbol("BTC/USD").WithPrice(1)).
		Add(origin.NewExchange("coinbase").WithSymbol("BTC/USD").WithStatusCode(http.StatusNotFound)).
		Add(origin.NewExchange("gemini").WithSymbol("BTC/USD").WithStatusCode(http.StatusNotFound)).
		Add(origin.NewExchange("kraken").WithSymbol("BTC/USD").WithStatusCode(http.StatusNotFound)).
		Deploy(*s)
	require.NoError(t, err)

	out, err := execCommand(ctx, "..", nil, "./gofer", "-c", "./e2e/testdata/config/gofer.hcl", "-v", "debug", "price", "BTC/USD", "--format", "json")
	require.NoError(t, err)

	priceMap, err := parseGoferPrice(out)
	require.NoError(t, err)
	p := priceMap["BTC/USD"]

	assert.Equal(t, "reference", p.Meta["type"])
	assert.NotEmpty(t, p.Error) // todo
}
