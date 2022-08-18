package gofere2e

import (
	"testing"

	"github.com/chronicleprotocol/infestor"
	"github.com/chronicleprotocol/infestor/origin"

	"github.com/stretchr/testify/suite"
)

func TestPriceWSTETH2ESuite(t *testing.T) {
	suite.Run(t, new(PriceWSTETHE2ESuite))
}

type PriceWSTETHE2ESuite struct {
	SmockerAPISuite
}

func (s *PriceWSTETHE2ESuite) TestPrice() {
	err := infestor.NewMocksBuilder().
		Reset().
		Add(origin.NewExchange("wsteth").WithSymbol("WSTETH/ETH").WithPrice(1)).
		Add(origin.NewExchange("balancerV2").WithSymbol("STETH/ETH").
			WithCustom("rate", "0x0000000000000000000000000000000000000000000000000EF976AF325D68E80000000000000000000000000000000000000000000000000000000000002A300000000000000000000000000000000000000000000000000000000062D81469").
			WithCustom("price", "0x0000000000000000000000000000000000000000000000000d925d70884a3395")).
		Add(origin.NewExchange("curve").WithSymbol("STETH/ETH").WithPrice(1)).
		Add(origin.NewExchange("ethrpc")).
		Deploy(s.api)

	s.Require().NoError(err)

	out, _, err := callGofer("-c", s.ConfigPath, "--norpc", "price", "WSTETH/ETH")
	s.Require().NoError(err)
	s.Require().NotEmpty(out)

	p, err := parseGoferPrice(out)
	s.Require().NoError(err)
	s.Require().Equal("aggregator", p.Type)
	s.Require().Equal(1.0561332642043832, p.Price)
	s.Require().Greater(len(p.Prices), 0)
	s.Require().Equal("median", p.Parameters["method"])
	s.Require().Equal("1", p.Parameters["minimumSuccessfulSources"])
}
