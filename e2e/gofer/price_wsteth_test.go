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
			WithCustom("match", "32296969ef14eb0c6d29669c550d4a044913023").
			WithCustom("response", "0000000000000000000000000000000000000000000000000000000000f1adbd00000000000000000000000000000000000000000000000000000000000000400000000000000000000000000000000000000000000000000000000000000001000000000000000000000000000000000000000000000000000000000000002000000000000000000000000000000000000000000000000000000000000000200000000000000000000000000000000000000000000000000dde952b2b374c17")).
		Add(origin.NewExchange("curve").WithSymbol("STETH/ETH").
			WithCustom("match", "dc24316b9ae028f1497c275eb9192a3ea0f67022").
			WithCustom("response", "0000000000000000000000000000000000000000000000000000000000f1adba00000000000000000000000000000000000000000000000000000000000000400000000000000000000000000000000000000000000000000000000000000001000000000000000000000000000000000000000000000000000000000000002000000000000000000000000000000000000000000000000000000000000000200000000000000000000000000000000000000000000000000ddde23599edb5d5")).
		Add(origin.NewExchange("ethrpc")).
		Deploy(s.api)

	s.Require().NoError(err)

	out, _, err := callGofer("-c", s.ConfigPath, "--norpc", "price", "WSTETH/ETH")
	s.Require().NoError(err)
	s.Require().NotEmpty(out)

	p, err := parseGoferPrice(out)
	s.Require().NoError(err)
	s.Require().Equal("aggregator", p.Type)
	s.Require().Equal(1.0697381612529844, p.Price)
	s.Require().Greater(len(p.Prices), 0)
	s.Require().Equal("median", p.Parameters["method"])
	s.Require().Equal("1", p.Parameters["minimumSuccessfulSources"])
}
