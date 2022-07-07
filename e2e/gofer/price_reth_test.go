package gofere2e

import (
	"testing"

	"github.com/chronicleprotocol/infestor"
	"github.com/chronicleprotocol/infestor/origin"

	"github.com/stretchr/testify/suite"
)

func TestPriceRETH2ESuite(t *testing.T) {
	suite.Run(t, new(PriceRETHE2ESuite))
}

type PriceRETHE2ESuite struct {
	SmockerAPISuite
}

func (s *PriceRETHE2ESuite) TestPrice() {
	err := infestor.NewMocksBuilder().
		Reset().
		// "RETH/ETH": {
		// [{ "origin": "rocketpool", "pair": "RETH/ETH" }],
		// [{ "origin": "balancerV2", "pair": "RETH/ETH" }],
		// [{ "origin": "curve", "pair": "RETH/WSTETH" },{ "origin": ".", "pair": "WSTETH/ETH" }]
		// "minimumSuccessfulSources": 3
		Add(origin.NewExchange("rocketpool").WithSymbol("RETH/ETH")).
		Add(origin.NewExchange("balancerV2").WithSymbol("RETH/WETH")).
		Add(origin.NewExchange("curve").WithSymbol("RETH/WSTETH")).

		// "WSTETH/ETH": {
		// [{ "origin": "wsteth", "pair": "WSTETH/STETH" },{ "origin": ".", "pair": "STETH/ETH" }]
		// "minimumSuccessfulSources": 1
		Add(origin.NewExchange("wsteth").WithSymbol("WSTETH/STETH")).

		// "STETH/ETH": {
		// [{ "origin": "curve", "pair": "STETH/ETH" }],
		// [{ "origin": "balancerV2", "pair": "STETH/ETH" }]
		// "minimumSuccessfulSources": 2
		Add(origin.NewExchange("curve").WithSymbol("STETH/ETH")).
		Add(origin.NewExchange("balancerV2").WithSymbol("STETH/ETH")).

		// getBlockNumber
		Add(origin.NewExchange("ethrpc")).
		Deploy(s.api)

	s.Require().NoError(err)

	out, _, err := callGofer("-c", s.ConfigPath, "price", "RETH/ETH")
	s.Require().NoError(err)
	s.Require().NotEmpty(out)

	p, err := parseGoferPrice(out)
	s.Require().NoError(err)
	s.Require().Equal("aggregator", p.Type)
	s.Require().Equal(1.0, p.Price)
	s.Require().Greater(len(p.Prices), 0)
	s.Require().Equal("median", p.Parameters["method"])
	s.Require().Equal("3", p.Parameters["minimumSuccessfulSources"])
}
