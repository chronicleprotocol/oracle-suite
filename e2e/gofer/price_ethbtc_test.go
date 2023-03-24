package gofere2e

import (
	"net/http"
	"testing"

	"github.com/chronicleprotocol/infestor"
	"github.com/chronicleprotocol/infestor/origin"
	"github.com/stretchr/testify/suite"
)

func TestPriceETHBTCE2ESuite(t *testing.T) {
	suite.Run(t, new(PriceETHBTCE2ESuite))
}

type PriceETHBTCE2ESuite struct {
	SmockerAPISuite
}

func (s *PriceETHBTCE2ESuite) TestPrice() {
	err := infestor.NewMocksBuilder().
		Reset().
		Add(origin.NewExchange("binance_us").WithSymbol("ETH/BTC").WithPrice(1)).
		Add(origin.NewExchange("bitstamp").WithSymbol("ETH/BTC").WithPrice(1)).
		Add(origin.NewExchange("coinbasepro").WithSymbol("ETH/BTC").WithPrice(1)).
		Add(origin.NewExchange("gemini").WithSymbol("ETH/BTC").WithPrice(1)).
		Add(origin.NewExchange("kraken").WithSymbol("ETH/BTC").WithPrice(1)).
		Deploy(s.api)

	s.Require().NoError(err)

	out, _, err := callGofer("-c", s.ConfigPath, "--norpc", "price", "ETH/BTC")
	s.Require().NoError(err)
	s.Require().NotEmpty(out)

	p, err := parseGoferPrice(out)
	s.Require().NoError(err)
	s.Require().Equal("aggregator", p.Type)
	s.Require().Equal(float64(1), p.Price)
	s.Require().Greater(len(p.Prices), 0)
	// TODO: Check if we still need it
	//s.Require().Equal("median", p.Parameters["method"])
	//s.Require().Equal("4", p.Parameters["minimumSuccessfulSources"])
}

func (s *PriceETHBTCE2ESuite) TestPrice3Correct2Zero() {
	err := infestor.NewMocksBuilder().
		Reset().
		Add(origin.NewExchange("binance_us").WithSymbol("ETH/BTC").WithPrice(1)).
		Add(origin.NewExchange("bitstamp").WithSymbol("ETH/BTC").WithPrice(1)).
		Add(origin.NewExchange("coinbasepro").WithSymbol("ETH/BTC").WithPrice(1)).
		Add(origin.NewExchange("gemini").WithSymbol("ETH/BTC").WithPrice(0)).
		Add(origin.NewExchange("kraken").WithSymbol("ETH/BTC").WithPrice(0)).
		Deploy(s.api)

	s.Require().NoError(err)

	out, _, err := callGofer("-c", s.ConfigPath, "--norpc", "price", "ETH/BTC")
	s.Require().NoError(err)
	s.Require().NotEmpty(out)

	p, err := parseGoferPrice(out)
	s.Require().NoError(err)
	s.Require().Equal("aggregator", p.Type)
	s.Require().Equal(1.0, p.Price)
	s.Require().Greater(len(p.Prices), 0)
	// TODO: Check if we still need it
	//s.Require().Equal("median", p.Parameters["method"])
	//s.Require().Equal("4", p.Parameters["minimumSuccessfulSources"])
}

func (s *PriceETHBTCE2ESuite) TestPrice3Correct2Invalid() {
	err := infestor.NewMocksBuilder().
		Reset().
		Add(origin.NewExchange("binance_us").WithSymbol("ETH/BTC").WithPrice(1)).
		Add(origin.NewExchange("bitstamp").WithSymbol("ETH/BTC").WithPrice(1)).
		Add(origin.NewExchange("coinbasepro").WithSymbol("ETH/BTC").WithPrice(1)).
		Add(origin.NewExchange("gemini").WithSymbol("ETH/BTC").WithStatusCode(http.StatusNotFound)).
		Add(origin.NewExchange("kraken").WithSymbol("ETH/BTC").WithStatusCode(http.StatusNotFound)).
		Deploy(s.api)

	s.Require().NoError(err)

	out, _, err := callGofer("-c", s.ConfigPath, "--norpc", "price", "ETH/BTC")
	s.Require().NoError(err)
	s.Require().NotEmpty(out)

	p, err := parseGoferPrice(out)
	s.Require().NoError(err)
	s.Require().Equal("aggregator", p.Type)
	s.Require().Equal(1.0, p.Price)
	s.Require().Greater(len(p.Prices), 0)
	// TODO: Check if we still need it
	//s.Require().Equal("median", p.Parameters["method"])
	//s.Require().Equal("4", p.Parameters["minimumSuccessfulSources"])
}

func (s *PriceETHBTCE2ESuite) TestPrice2Correct3Invalid() {
	err := infestor.NewMocksBuilder().
		Reset().
		Add(origin.NewExchange("binance_us").WithSymbol("ETH/BTC").WithPrice(1)).
		Add(origin.NewExchange("coinbasepro").WithSymbol("ETH/BTC").WithPrice(1)).
		Add(origin.NewExchange("bitstamp").WithSymbol("ETH/BTC").WithStatusCode(http.StatusNotFound)).
		Add(origin.NewExchange("gemini").WithSymbol("ETH/BTC").WithStatusCode(http.StatusNotFound)).
		Add(origin.NewExchange("kraken").WithSymbol("ETH/BTC").WithStatusCode(http.StatusNotFound)).
		Deploy(s.api)

	s.Require().NoError(err)

	_, code, err := callGofer("-c", s.ConfigPath, "--norpc", "price", "ETH/BTC")
	s.Require().Error(err)
	s.Require().Equal(1, code)
}

func (s *PriceETHBTCE2ESuite) TestPriceMedianCalculationNotEnoughMinSources() {
	err := infestor.NewMocksBuilder().
		Reset().
		Add(origin.NewExchange("binance_us").WithSymbol("ETH/BTC").WithPrice(1)).
		Add(origin.NewExchange("bitstamp").WithSymbol("ETH/BTC").WithStatusCode(http.StatusNotFound)).
		Add(origin.NewExchange("coinbase").WithSymbol("ETH/BTC").WithStatusCode(http.StatusNotFound)).
		Add(origin.NewExchange("gemini").WithSymbol("ETH/BTC").WithStatusCode(http.StatusNotFound)).
		Add(origin.NewExchange("kraken").WithSymbol("ETH/BTC").WithStatusCode(http.StatusNotFound)).
		Deploy(s.api)

	s.Require().NoError(err)

	_, code, err := callGofer("-c", s.ConfigPath, "--norpc", "price", "ETH/BTC")
	s.Require().Error(err)
	s.Require().Equal(1, code)
}
