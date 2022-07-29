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
		Add(origin.NewExchange("rocketpool").WithSymbol("RETH/ETH")).
		Add(origin.NewExchange("balancerV2").WithSymbol("RETH/WETH").
			WithCustom("rate", "0x0000000000000000000000000000000000000000000000000EF976AF325D68E80000000000000000000000000000000000000000000000000000000000002A300000000000000000000000000000000000000000000000000000000062D81469").
			WithCustom("price", "0x0000000000000000000000000000000000000000000000000d925d70884a3395")).
		Add(origin.NewExchange("curve").WithSymbol("RETH/WSTETH")).
		Add(origin.NewExchange("wsteth").WithSymbol("WSTETH/STETH")).
		Add(origin.NewExchange("curve").WithSymbol("STETH/ETH").WithPrice(1.044)).
		Add(origin.NewExchange("balancerV2").WithSymbol("STETH/ETH").
			WithCustom("rate", "0x0000000000000000000000000000000000000000000000000EF976AF325D68E80000000000000000000000000000000000000000000000000000000000002A300000000000000000000000000000000000000000000000000000000062D81469").
			WithCustom("price", "0x0000000000000000000000000000000000000000000000000d925d70884a3395")).
		Add(origin.NewExchange("ethrpc")).
		Deploy(s.api)

	s.Require().NoError(err)
	out, _, err := callGofer("-c", s.ConfigPath, "price", "RETH/ETH")
	s.Require().NoError(err)
	s.Require().NotEmpty(out)

	p, err := parseGoferPrice(out)
	s.Require().NoError(err)
	s.Require().Equal("aggregator", p.Type)
	s.Require().Equal(1.0511045977366833, p.Price)
	s.Require().Greater(len(p.Prices), 0)
	s.Require().Equal("median", p.Parameters["method"])
	s.Require().Equal("3", p.Parameters["minimumSuccessfulSources"])
}

func (s *PriceRETHE2ESuite) TestCircuit() {
	err := infestor.NewMocksBuilder().
		Reset().
		Add(origin.NewExchange("rocketpool").WithSymbol("RETH/ETH").
			WithCustom("price", "0x0000000000000000000000000000000000000000000000001b7314242fae3395")). // 1.97
		Add(origin.NewExchange("balancerV2").WithSymbol("RETH/WETH").
			WithCustom("rate", "0x0000000000000000000000000000000000000000000000000EF976AF325D68E80000000000000000000000000000000000000000000000000000000000002A300000000000000000000000000000000000000000000000000000000062D81469").
			WithCustom("price", "0x0000000000000000000000000000000000000000000000000d925d70884a3395")).
		Add(origin.NewExchange("curve").WithSymbol("RETH/WSTETH")).
		Add(origin.NewExchange("wsteth").WithSymbol("WSTETH/STETH")).
		Add(origin.NewExchange("curve").WithSymbol("STETH/ETH").WithPrice(1.044)).
		Add(origin.NewExchange("balancerV2").WithSymbol("STETH/ETH").
			WithCustom("rate", "0x0000000000000000000000000000000000000000000000000EF976AF325D68E80000000000000000000000000000000000000000000000000000000000002A300000000000000000000000000000000000000000000000000000000062D81469").
			WithCustom("price", "0x0000000000000000000000000000000000000000000000000d925d70884a3395")).
		Add(origin.NewExchange("ethrpc")).
		Deploy(s.api)

	s.Require().NoError(err)
	_, _, err = callGofer("-c", s.ConfigPath, "price", "RETH/ETH")
	s.Require().Error(err)
}
