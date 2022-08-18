package ghost

import (
	"context"
	"net/http"
	"testing"
	"time"

	"github.com/chronicleprotocol/infestor"
	"github.com/chronicleprotocol/infestor/origin"
	"github.com/chronicleprotocol/infestor/smocker"
	"github.com/stretchr/testify/require"
)

func TestPairsPrice(t *testing.T) {
	ctx, ctxCancel := context.WithTimeout(context.Background(), 2*time.Minute)

	s := smocker.NewAPI(env("SMOCKER_URL", "http://127.0.0.1:8081"))
	err := s.Reset(ctx)
	if err != nil {
		require.Fail(t, err.Error())
	}

	// Setup price for BTC/USD
	err = infestor.NewMocksBuilder().
		Reset().
		Add(origin.NewExchange("bitstamp").WithSymbol("BTC/USD").WithPrice(1)).
		Add(origin.NewExchange("bittrex").WithSymbol("BTC/USD").WithPrice(1)).
		Add(origin.NewExchange("coinbase").WithSymbol("BTC/USD").WithPrice(1)).
		Add(origin.NewExchange("gemini").WithSymbol("BTC/USD").WithPrice(1)).
		Add(origin.NewExchange("kraken").WithSymbol("XXBT/ZUSD").WithPrice(1)).
		Deploy(*s)

	require.NoError(t, err)

	ghostCmd := command(ctx, "../..", "./ghost", "run", "-c", "./e2e/ghost/testdata/config/ghost.json", "--gofer.norpc", "-v", "debug")
	spireCmd := command(ctx, "../..", "./spire", "agent", "-c", "./e2e/ghost/testdata/config/spire.json", "-v", "debug")
	defer func() {
		ctxCancel()
		_ = ghostCmd.Wait()
		_ = spireCmd.Wait()
	}()

	if err := ghostCmd.Start(); err != nil {
		require.Fail(t, err.Error())
	}
	waitForPort(ctx, "localhost", 30100)

	if err := spireCmd.Start(); err != nil {
		require.Fail(t, err.Error())
	}
	waitForPort(ctx, "localhost", 30101)

	cli, err := buildTransport(ctx, "./testdata/config/spire_client.json")
	require.NoError(t, err)

	time.Sleep(5 * time.Second)

	price, err := waitForSpire(ctx, cli, "BTCUSD", "0x2D800d93B065CE011Af83f316ceF9F0d005B0AA4")
	require.NoError(t, err)
	require.NotNil(t, price)

	require.Equal(t, "1000000000000000000", price.Price.Val.String())
	require.Equal(t, "BTCUSD", price.Price.Wat)

	// Next price round
	// Setup price for BTC/USD
	err = infestor.NewMocksBuilder().
		Reset().
		Add(origin.NewExchange("bitstamp").WithSymbol("BTC/USD").WithPrice(2)).
		Add(origin.NewExchange("bittrex").WithSymbol("BTC/USD").WithPrice(2)).
		Add(origin.NewExchange("coinbase").WithSymbol("BTC/USD").WithPrice(2)).
		Add(origin.NewExchange("gemini").WithSymbol("BTC/USD").WithPrice(2)).
		Add(origin.NewExchange("kraken").WithSymbol("XXBT/ZUSD").WithPrice(2)).
		Deploy(*s)

	require.NoError(t, err)

	time.Sleep(10 * time.Second)

	price, err = waitForSpire(ctx, cli, "BTCUSD", "0x2D800d93B065CE011Af83f316ceF9F0d005B0AA4")
	require.NoError(t, err)
	require.NotNil(t, price)

	require.Equal(t, "2000000000000000000", price.Price.Val.String())
	require.Equal(t, "BTCUSD", price.Price.Wat)

	// Next price round
	// Setup price for BTC/USD
	err = infestor.NewMocksBuilder().
		Reset().
		Add(origin.NewExchange("bitstamp").WithSymbol("BTC/USD").WithPrice(3)).
		Add(origin.NewExchange("bittrex").WithSymbol("BTC/USD").WithPrice(3)).
		Add(origin.NewExchange("coinbase").WithSymbol("BTC/USD").WithPrice(3)).
		Add(origin.NewExchange("gemini").WithSymbol("BTC/USD").WithPrice(3)).
		Add(origin.NewExchange("kraken").WithSymbol("XXBT/ZUSD").WithPrice(3)).
		Deploy(*s)

	require.NoError(t, err)

	time.Sleep(10 * time.Second)

	price, err = waitForSpire(ctx, cli, "BTCUSD", "0x2D800d93B065CE011Af83f316ceF9F0d005B0AA4")
	require.NoError(t, err)
	require.NotNil(t, price)

	require.Equal(t, "3000000000000000000", price.Price.Val.String())
	require.Equal(t, "BTCUSD", price.Price.Wat)
}

func TestPairsInvalidPrice(t *testing.T) {
	ctx, ctxCancel := context.WithTimeout(context.Background(), 2*time.Minute)

	s := smocker.NewAPI(env("SMOCKER_URL", "http://127.0.0.1:8081"))
	err := s.Reset(ctx)
	if err != nil {
		require.Fail(t, err.Error())
	}

	// Setup price for BTC/USD
	err = infestor.NewMocksBuilder().
		Reset().
		Add(origin.NewExchange("bitstamp").WithSymbol("BTC/USD").WithStatusCode(http.StatusConflict)).
		Add(origin.NewExchange("bittrex").WithSymbol("BTC/USD").WithStatusCode(http.StatusConflict)).
		Add(origin.NewExchange("coinbase").WithSymbol("BTC/USD").WithStatusCode(http.StatusConflict)).
		Add(origin.NewExchange("gemini").WithSymbol("BTC/USD").WithStatusCode(http.StatusConflict)).
		Add(origin.NewExchange("kraken").WithSymbol("XXBT/ZUSD").WithStatusCode(http.StatusConflict)).
		Deploy(*s)

	require.NoError(t, err)

	ghostCmd := command(ctx, "../..", "./ghost", "run", "-c", "./e2e/ghost/testdata/config/ghost.json", "--gofer.norpc", "-v", "debug")
	spireCmd := command(ctx, "../..", "./spire", "agent", "-c", "./e2e/ghost/testdata/config/spire.json", "-v", "debug")
	defer func() {
		ctxCancel()
		_ = ghostCmd.Wait()
		_ = spireCmd.Wait()
	}()

	if err := ghostCmd.Start(); err != nil {
		require.Fail(t, err.Error())
	}
	waitForPort(ctx, "localhost", 30100)

	if err := spireCmd.Start(); err != nil {
		require.Fail(t, err.Error())
	}
	waitForPort(ctx, "localhost", 30101)

	cli, err := buildTransport(ctx, "./testdata/config/spire_client.json")
	require.NoError(t, err)

	time.Sleep(25 * time.Second)

	price, err := waitForSpire(ctx, cli, "BTCUSD", "0x2D800d93B065CE011Af83f316ceF9F0d005B0AA4")
	require.NoError(t, err)
	require.Nil(t, price)
}

func TestPairsMinPrice(t *testing.T) {
	ctx, ctxCancel := context.WithTimeout(context.Background(), 2*time.Minute)

	s := smocker.NewAPI(env("SMOCKER_URL", "http://127.0.0.1:8081"))
	err := s.Reset(ctx)
	if err != nil {
		require.Fail(t, err.Error())
	}

	// Setup price for BTC/USD
	err = infestor.NewMocksBuilder().
		Reset().
		Add(origin.NewExchange("bitstamp").WithSymbol("BTC/USD").WithPrice(1)).
		Add(origin.NewExchange("bittrex").WithSymbol("BTC/USD").WithPrice(1)).
		Add(origin.NewExchange("coinbase").WithSymbol("BTC/USD").WithPrice(1)).
		Add(origin.NewExchange("gemini").WithSymbol("BTC/USD").WithStatusCode(http.StatusConflict)).
		Add(origin.NewExchange("kraken").WithSymbol("XXBT/ZUSD").WithStatusCode(http.StatusConflict)).
		Deploy(*s)

	require.NoError(t, err)

	ghostCmd := command(ctx, "../..", "./ghost", "run", "-c", "./e2e/ghost/testdata/config/ghost.json", "--gofer.norpc", "-v", "debug")
	spireCmd := command(ctx, "../..", "./spire", "agent", "-c", "./e2e/ghost/testdata/config/spire.json", "-v", "debug")
	defer func() {
		ctxCancel()
		_ = ghostCmd.Wait()
		_ = spireCmd.Wait()
	}()

	if err := ghostCmd.Start(); err != nil {
		require.Fail(t, err.Error())
	}
	waitForPort(ctx, "localhost", 30100)

	if err := spireCmd.Start(); err != nil {
		require.Fail(t, err.Error())
	}
	waitForPort(ctx, "localhost", 30101)

	cli, err := buildTransport(ctx, "./testdata/config/spire_client.json")
	require.NoError(t, err)

	time.Sleep(5 * time.Second)

	price, err := waitForSpire(ctx, cli, "BTCUSD", "0x2D800d93B065CE011Af83f316ceF9F0d005B0AA4")
	require.NoError(t, err)
	require.NotNil(t, price)

	require.Equal(t, "1000000000000000000", price.Price.Val.String())
	require.Equal(t, "BTCUSD", price.Price.Wat)
}
