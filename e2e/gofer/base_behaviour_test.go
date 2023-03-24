package gofere2e

import (
	"testing"

	"github.com/stretchr/testify/suite"
)

func TestBaseBehaviourE2ESuite(t *testing.T) {
	suite.Run(t, new(BaseBehaviourE2ESuite))
}

type BaseBehaviourE2ESuite struct {
	SmockerAPISuite
}

func (s *BaseBehaviourE2ESuite) TestVersionCommand() {
	out, _, err := callGofer("--version")

	s.Require().NoError(err)
	s.Require().Contains(out, "gofer")
}

func (s *BaseBehaviourE2ESuite) TestHelpCommand() {
	out, _, err := callGofer("--help")

	s.Require().NoError(err)
	s.Require().NotEmpty(out)
	s.Require().Contains(out, "gofer")
}

func (s *BaseBehaviourE2ESuite) TestPairsCommand() {
	out, _, err := callGofer("--config", s.ConfigPath, "--norpc", "pairs")

	s.Require().NoError(err)
	s.Require().Contains(out, "BTC/USD")
}

func (s *BaseBehaviourE2ESuite) TestPricesInvalidPairCommand() {
	_, code, err := callGofer("--config", s.ConfigPath, "prices", "BTCUSD")

	s.Require().Error(err)
	s.Require().Equal(1, code)
}
