package gofere2e

import (
	"net/http"
	"testing"

	"github.com/chronicleprotocol/infestor"
	"github.com/chronicleprotocol/infestor/origin"

	"github.com/stretchr/testify/suite"
)

func TestPriceBTCUSDE2ESuite(t *testing.T) {
	suite.Run(t, new(PriceBTCUSDE2ESuite))
}

type PriceBTCUSDE2ESuite struct {
	SmockerAPISuite
}

func (s *PriceBTCUSDE2ESuite) TestPrice() {
	err := infestor.NewMocksBuilder().
		Reset().
		Add(origin.NewExchange("bitstamp").WithSymbol("BTC/USD").WithPrice(1)).
		Add(origin.NewExchange("bittrex").WithSymbol("BTC/USD").WithPrice(1)).
		Add(origin.NewExchange("coinbase").WithSymbol("BTC/USD").WithPrice(1)).
		Add(origin.NewExchange("gemini").WithSymbol("BTC/USD").WithPrice(1)).
		Add(origin.NewExchange("kraken").WithSymbol("XXBT/ZUSD").WithPrice(1)).
		Deploy(s.api)

	s.Require().NoError(err)

	out, _, err := callGofer("-c", s.ConfigPath, "--norpc", "price", "BTC/USD")
	s.Require().NoError(err)
	s.Require().NotEmpty(out)

	p, err := parseGoferPrice(out)
	s.Require().NoError(err)
	s.Require().Equal("aggregator", p.Type)
	s.Require().Equal(float64(1), p.Price)
	s.Require().Greater(len(p.Prices), 0)
	s.Require().Equal("median", p.Parameters["method"])
	s.Require().Equal("3", p.Parameters["minimumSuccessfulSources"])
}

func (s *PriceBTCUSDE2ESuite) TestPrice4Correct1Zero() {
	err := infestor.NewMocksBuilder().
		Reset().
		Add(origin.NewExchange("bitstamp").WithSymbol("BTC/USD").WithPrice(1)).
		Add(origin.NewExchange("bittrex").WithSymbol("BTC/USD").WithPrice(1)).
		Add(origin.NewExchange("coinbase").WithSymbol("BTC/USD").WithPrice(1)).
		Add(origin.NewExchange("gemini").WithSymbol("BTC/USD").WithPrice(1)).
		Add(origin.NewExchange("kraken").WithSymbol("XXBT/ZUSD").WithPrice(0)).
		Deploy(s.api)

	s.Require().NoError(err)

	out, _, err := callGofer("-c", s.ConfigPath, "--norpc", "price", "BTC/USD")
	s.Require().NoError(err)
	s.Require().NotEmpty(out)

	p, err := parseGoferPrice(out)
	s.Require().NoError(err)
	s.Require().Equal("aggregator", p.Type)
	s.Require().Equal(float64(1), p.Price)
	s.Require().Greater(len(p.Prices), 0)
	s.Require().Equal("median", p.Parameters["method"])
	s.Require().Equal("3", p.Parameters["minimumSuccessfulSources"])
}

func (s *PriceBTCUSDE2ESuite) TestPrice4Correct1Invalid() {
	err := infestor.NewMocksBuilder().
		Reset().
		Add(origin.NewExchange("bitstamp").WithSymbol("BTC/USD").WithPrice(1)).
		Add(origin.NewExchange("bittrex").WithSymbol("BTC/USD").WithPrice(1)).
		Add(origin.NewExchange("coinbase").WithSymbol("BTC/USD").WithPrice(1)).
		Add(origin.NewExchange("gemini").WithSymbol("BTC/USD").WithPrice(1)).
		Add(origin.NewExchange("kraken").WithSymbol("XXBT/ZUSD").WithStatusCode(http.StatusNotFound)).
		Deploy(s.api)

	s.Require().NoError(err)

	out, _, err := callGofer("-c", s.ConfigPath, "--norpc", "price", "BTC/USD")
	s.Require().NoError(err)
	s.Require().NotEmpty(out)

	p, err := parseGoferPrice(out)
	s.Require().NoError(err)
	s.Require().Equal("aggregator", p.Type)
	s.Require().Equal(float64(1), p.Price)
	s.Require().Greater(len(p.Prices), 0)
	s.Require().Equal("median", p.Parameters["method"])
	s.Require().Equal("3", p.Parameters["minimumSuccessfulSources"])
}

func (s *PriceBTCUSDE2ESuite) TestPrice3Correct2Invalid() {
	err := infestor.NewMocksBuilder().
		Reset().
		Add(origin.NewExchange("bitstamp").WithSymbol("BTC/USD").WithPrice(1)).
		Add(origin.NewExchange("bittrex").WithSymbol("BTC/USD").WithPrice(1)).
		Add(origin.NewExchange("coinbase").WithSymbol("BTC/USD").WithPrice(1)).
		Add(origin.NewExchange("gemini").WithSymbol("BTC/USD").WithStatusCode(http.StatusNotFound)).
		Add(origin.NewExchange("kraken").WithSymbol("XXBT/ZUSD").WithStatusCode(http.StatusNotFound)).
		Deploy(s.api)

	s.Require().NoError(err)

	out, _, err := callGofer("-c", s.ConfigPath, "--norpc", "price", "BTC/USD")
	s.Require().NoError(err)
	s.Require().NotEmpty(out)

	p, err := parseGoferPrice(out)
	s.Require().NoError(err)
	s.Require().Equal("aggregator", p.Type)
	s.Require().Equal(float64(1), p.Price)
	s.Require().Greater(len(p.Prices), 0)
	s.Require().Equal("median", p.Parameters["method"])
	s.Require().Equal("3", p.Parameters["minimumSuccessfulSources"])
}

func (s *PriceBTCUSDE2ESuite) TestPriceMedianCalculationNotEnoughMinSources() {
	err := infestor.NewMocksBuilder().
		Reset().
		Add(origin.NewExchange("bitstamp").WithSymbol("BTC/USD").WithPrice(1)).
		Add(origin.NewExchange("bittrex").WithSymbol("BTC/USD").WithStatusCode(http.StatusNotFound)).
		Add(origin.NewExchange("coinbase").WithSymbol("BTC/USD").WithStatusCode(http.StatusNotFound)).
		Add(origin.NewExchange("gemini").WithSymbol("BTC/USD").WithStatusCode(http.StatusNotFound)).
		Add(origin.NewExchange("kraken").WithSymbol("XXBT/ZUSD").WithStatusCode(http.StatusNotFound)).
		Deploy(s.api)

	s.Require().NoError(err)

	_, code, err := callGofer("-c", s.ConfigPath, "--norpc", "price", "BTC/USD")
	s.Require().Error(err)
	s.Require().Equal(1, code)
}
