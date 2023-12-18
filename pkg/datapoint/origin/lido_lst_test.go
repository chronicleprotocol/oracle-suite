package origin

import (
	"context"
	"math/big"
	"testing"

	"github.com/defiweb/go-eth/types"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/suite"

	"github.com/chronicleprotocol/oracle-suite/pkg/datapoint/value"
	ethereumMocks "github.com/chronicleprotocol/oracle-suite/pkg/ethereum/mocks"
	"github.com/chronicleprotocol/oracle-suite/pkg/util/bn"
)

type LidoLSTSuite struct {
	suite.Suite
	addresses map[string]string
	client    *ethereumMocks.RPC
	origin    *LidoLST
}

func (suite *LidoLSTSuite) SetupTest() {
	suite.client = &ethereumMocks.RPC{}
	o, err := NewLidoLST(LidoLSTConfig{
		Client: suite.client,
		ContractAddresses: ContractAddresses{
			AssetPair{"STETH", "ERC20"}: types.MustAddressFromHex("0xae7ab96520DE3A18E5e111B5EaAb095312D7fE84"),
		},
		Blocks: []int64{0, 10, 20},
		Logger: nil,
	})
	suite.NoError(err)
	suite.origin = o
}
func (suite *LidoLSTSuite) TearDownTest() {
	suite.origin = nil
	suite.client = nil
}

func (suite *LidoLSTSuite) Origin() *LidoLST {
	return suite.origin
}

func TestLidoLSTSuite(t *testing.T) {
	suite.Run(t, new(LidoLSTSuite))
}

func (suite *LidoLSTSuite) TestFailOnWrongPair() {
	pair := value.Pair{Base: "x", Quote: "y"}

	suite.client.On(
		"BlockNumber",
		mock.Anything,
	).Return(big.NewInt(100), nil).Once()

	_, err := suite.origin.FetchDataPoints(context.Background(), []any{pair})
	suite.Require().EqualError(err, "quote token should be `nDAYS`, n is digit")
}

func (suite *LidoLSTSuite) TestSuccessOnDaysPair() {
	pair := value.Pair{Base: "x", Quote: "6DAYS"}

	suite.client.On(
		"BlockNumber",
		mock.Anything,
	).Return(big.NewInt(100), nil).Once()

	_, err := suite.origin.FetchDataPoints(context.Background(), []any{pair})
	suite.Require().EqualError(err, "not found TokenRebased event")
}

func (suite *LidoLSTSuite) TestCalculateAPR() {
	// https://etherscan.io/tx/0x251f8cc3fea4be64c1f9a9afd8ba5c03472f61285ffdb895ab32cc9d339c836e#eventlog
	preTotalEther, _ := new(big.Int).SetString("9213013075226300642092863", 10)
	preTotalShares, _ := new(big.Int).SetString("8008758831591232087882503", 10)
	postTotalEther, _ := new(big.Int).SetString("9145130236121447961079791", 10)
	postTotalShares, _ := new(big.Int).SetString("7948878434110346407286571", 10)
	event := rebaseEvent{
		blockNumber:     big.NewInt(18805718),
		reportTimestamp: big.NewInt(1702814411),
		preTotalEther:   preTotalEther,
		preTotalShares:  preTotalShares,
		postTotalEther:  postTotalEther,
		postTotalShares: postTotalShares,
		timeElapsed:     big.NewInt(86400),
	}

	apr := suite.origin.calculateAprFromRebaseEvent(event)
	// Lido Staking Instant APR at 2023-12-17 12:23 is 3.998031409393279 in Dune
	// https://dune.com/queries/570874/1068499
	suite.Require().True(expectEqualWithFloatError(apr, bn.DecFloatPoint("3.998031409393279"), bn.DecFloatPoint(1).SetPrec(8)))
}

func expectEqualWithFloatError(actual *bn.DecFloatPointNumber, expected *bn.DecFloatPointNumber, error *bn.DecFloatPointNumber) bool {
	acceptedError := expected.Mul(error)
	if acceptedError.Cmp(bn.DecFloatPoint(int64(0))) > 0 {
		if actual.Cmp(expected.Sub(acceptedError)) < 0 {
			return false
		}
		if actual.Cmp(expected.Add(acceptedError)) > 0 {
			return false
		}
		return true
	} else {
		if actual.Cmp(expected.Sub(acceptedError)) > 0 {
			return false
		}
		if actual.Cmp(expected.Add(acceptedError)) < 0 {
			return false
		}
		return true
	}
}
