package gofere2e

import (
	"testing"

	"github.com/chronicleprotocol/infestor"
	"github.com/chronicleprotocol/infestor/origin"

	"github.com/stretchr/testify/suite"
)

func TestPriceRETHUSD2ESuite(t *testing.T) {
	suite.Run(t, new(PriceRETHUSDE2ESuite))
}

type PriceRETHUSDE2ESuite struct {
	SmockerAPISuite
}

func (s *PriceRETHUSDE2ESuite) TestPrice() {
	err := infestor.NewMocksBuilder().
		Reset().
		// "RETH/USD": {
		// [{ "origin": ".", "pair": "RETH/ETH" },{ "origin": ".", "pair": "ETH/USD" }]
		// "minimumSuccessfulSources": 1

		// "RETH/ETH": {
		// [{ "origin": "rocketpool", "pair": "RETH/ETH" }],
		// [{ "origin": "balancerV2", "pair": "RETH/ETH" }],
		// [{ "origin": "curve", "pair": "RETH/WSTETH" },{ "origin": ".", "pair": "WSTETH/ETH" }]
		// "minimumSuccessfulSources": 3
		Add(origin.NewExchange("rocketpool").WithSymbol("RETH/ETH").WithPrice(1)).
		Add(origin.NewExchange("balancerV2").WithSymbol("RETH/ETH").WithPrice(1)).
		Add(origin.NewExchange("curve").WithSymbol("RETH/WSTETH").WithPrice(1)).

		// "WSTETH/ETH": {
		// [{ "origin": "wsteth", "pair": "WSTETH/STETH" },{ "origin": ".", "pair": "STETH/ETH" }]
		// "minimumSuccessfulSources": 1
		Add(origin.NewExchange("wsteth").WithSymbol("WSTETH/STETH").WithPrice(1)).

		// "STETH/ETH": {
		// [{ "origin": "curve", "pair": "STETH/ETH" }],
		// [{ "origin": "balancerV2", "pair": "STETH/ETH" }]
		// "minimumSuccessfulSources": 2
		Add(origin.NewExchange("curve").WithSymbol("STETH/ETH").WithPrice(1)).
		Add(origin.NewExchange("balancerV2").WithSymbol("STETH/ETH").WithPrice(1)).

		// "ETH/USD": {
		// [{ "origin": "binance", "pair": "ETH/BTC" },{ "origin": ".", "pair": "BTC/USD" }],
		// [{ "origin": "bitstamp", "pair": "ETH/USD" }],
		// [{ "origin": "coinbasepro", "pair": "ETH/USD" }],
		// [{ "origin": "ftx", "pair": "ETH/USD" }],
		// [{ "origin": "gemini", "pair": "ETH/USD" }],
		// [{ "origin": "kraken", "pair": "ETH/USD" }],
		// [{ "origin": "uniswapV3", "pair": "ETH/USD" }]
		// "minimumSuccessfulSources": 4
		Add(origin.NewExchange("binance").WithSymbol("ETH/BTC").WithPrice(1)).
		Add(origin.NewExchange("bitstamp").WithSymbol("ETH/USD").WithPrice(1)).
		Add(origin.NewExchange("gemini").WithSymbol("ETH/USD").WithPrice(1)).
		Add(origin.NewExchange("ftx").WithSymbol("ETH/USD").WithPrice(1)).

		// "BTC/USD": {
		// [{ "origin": "bitstamp", "pair": "BTC/USD" }],
		// [{ "origin": "bittrex", "pair": "BTC/USD" }],
		// [{ "origin": "coinbasepro", "pair": "BTC/USD" }],
		// [{ "origin": "gemini", "pair": "BTC/USD" }],
		// [{ "origin": "kraken", "pair": "BTC/USD" }]
		// "minimumSuccessfulSources": 3
		Add(origin.NewExchange("bitstamp").WithSymbol("BTC/USD").WithPrice(1)).
		Add(origin.NewExchange("bittrex").WithSymbol("BTC/USD").WithPrice(1)).
		Add(origin.NewExchange("gemini").WithSymbol("BTC/USD").WithPrice(1)).
		Deploy(s.api)

	s.Require().NoError(err)

	out, _, err := callGofer("-c", s.ConfigPath, "--norpc", "price", "RETH/USD")
	s.Require().NoError(err)
	s.Require().NotEmpty(out)

	p, err := parseGoferPrice(out)
	s.Require().NoError(err)
	s.Require().Equal("aggregator", p.Type)
	s.Require().Equal(1.065747073876745, p.Price)
	s.Require().Greater(len(p.Prices), 0)
	s.Require().Equal("median", p.Parameters["method"])
	s.Require().Equal("1", p.Parameters["minimumSuccessfulSources"])
}
