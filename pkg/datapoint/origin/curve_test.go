package origin

import (
	"context"
	"math/big"
	"testing"

	"github.com/defiweb/go-eth/abi"
	"github.com/defiweb/go-eth/hexutil"
	"github.com/defiweb/go-eth/types"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/suite"

	"github.com/chronicleprotocol/oracle-suite/pkg/datapoint/value"
	ethereumMocks "github.com/chronicleprotocol/oracle-suite/pkg/ethereum/mocks"
)

type CurveSuite struct {
	suite.Suite
	client *ethereumMocks.RPC
	origin *Curve
}

func (suite *CurveSuite) SetupTest() {
	suite.client = &ethereumMocks.RPC{}
	o, err := NewCurve(CurveConfig{
		Client: suite.client,
		ContractAddresses: map[string]string{
			"ETH/STETH": "0xDC24316b9AE028F1497c275EB9192a3Ea0f67022",
		},
		Blocks: []int64{0, 10, 20},
		Logger: nil,
	})
	suite.NoError(err)
	suite.origin = o
}

func (suite *CurveSuite) TearDownTest() {
	suite.origin = nil
	suite.client = nil
}

func (suite *CurveSuite) Origin() *Curve {
	return suite.origin
}

func TestCurveSuite(t *testing.T) {
	suite.Run(t, new(CurveSuite))
}

func (suite *CurveSuite) TestSuccessResponse() {
	resp := [][]byte{
		types.Bytes(big.NewInt(0.94 * ether).Bytes()).PadLeft(32),
		types.Bytes(big.NewInt(0.98 * ether).Bytes()).PadLeft(32),
		types.Bytes(big.NewInt(0.99 * ether).Bytes()).PadLeft(32),
	}

	ctx := context.Background()
	blockNumber := big.NewInt(100)

	suite.client.On(
		"ChainID",
		ctx,
	).Return(uint64(1), nil)

	suite.client.On(
		"BlockNumber",
		ctx,
	).Return(blockNumber, nil)

	// MultiCall contract
	contract := types.MustAddressFromHex("0xeefba1e63905ef1d7acba5a8513c70307c1ce441")

	// Generate encoded return value of `aggregate` function
	//function aggregate(
	//	(address target, bytes callData)[] memory calls
	//) public returns (
	//	uint256 blockNumber,
	//	bytes[] memory returnData
	//)

	tuple := abi.MustParseType("(uint256,bytes[] memory)")
	respEncoded, _ := abi.EncodeValues(tuple, blockNumber.Uint64(), []any{resp[0]})
	suite.client.On(
		"Call",
		ctx,
		types.Call{
			To:    &contract,
			Input: hexutil.MustHexToBytes("252dba42000000000000000000000000000000000000000000000000000000000000002000000000000000000000000000000000000000000000000000000000000000010000000000000000000000000000000000000000000000000000000000000020000000000000000000000000dc24316b9ae028f1497c275eb9192a3ea0f67022000000000000000000000000000000000000000000000000000000000000004000000000000000000000000000000000000000000000000000000000000000645e0d443f000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000010000000000000000000000000000000000000000000000000de0b6b3a764000000000000000000000000000000000000000000000000000000000000"),
		},
		types.BlockNumberFromUint64(uint64(100)),
	).Return(respEncoded, nil).Once()

	respEncoded, _ = abi.EncodeValues(tuple, blockNumber.Uint64(), []any{resp[1]})
	suite.client.On(
		"Call",
		ctx,
		types.Call{
			To:    &contract,
			Input: hexutil.MustHexToBytes("252dba42000000000000000000000000000000000000000000000000000000000000002000000000000000000000000000000000000000000000000000000000000000010000000000000000000000000000000000000000000000000000000000000020000000000000000000000000dc24316b9ae028f1497c275eb9192a3ea0f67022000000000000000000000000000000000000000000000000000000000000004000000000000000000000000000000000000000000000000000000000000000645e0d443f000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000010000000000000000000000000000000000000000000000000de0b6b3a764000000000000000000000000000000000000000000000000000000000000"),
		},
		types.BlockNumberFromUint64(uint64(90)),
	).Return(respEncoded, nil).Once()

	respEncoded, _ = abi.EncodeValues(tuple, blockNumber.Uint64(), []any{resp[2]})
	suite.client.On(
		"Call",
		ctx,
		types.Call{
			To:    &contract,
			Input: hexutil.MustHexToBytes("252dba42000000000000000000000000000000000000000000000000000000000000002000000000000000000000000000000000000000000000000000000000000000010000000000000000000000000000000000000000000000000000000000000020000000000000000000000000dc24316b9ae028f1497c275eb9192a3ea0f67022000000000000000000000000000000000000000000000000000000000000004000000000000000000000000000000000000000000000000000000000000000645e0d443f000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000010000000000000000000000000000000000000000000000000de0b6b3a764000000000000000000000000000000000000000000000000000000000000"),
		},
		types.BlockNumberFromUint64(uint64(80)),
	).Return(respEncoded, nil).Once()

	pair := value.Pair{Base: "ETH", Quote: "STETH"}
	points, err := suite.origin.FetchDataPoints(context.Background(), []any{pair})
	suite.Require().NoError(err)
	suite.Equal(0.97, points[pair].Value.(value.Tick).Price.Float64())
	suite.Greater(points[pair].Time.Unix(), int64(0))
}

func (suite *CurveSuite) TestSuccessResponse_Inverse() {
	suite.client.On(
		"BlockNumber",
		mock.Anything,
	).Return(big.NewInt(100), nil).Once()

	pair := value.Pair{Base: "STETH", Quote: "ETH"}
	points, err := suite.origin.FetchDataPoints(context.Background(), []any{pair})
	suite.Require().NoError(err)
	suite.Require().EqualError(points[pair].Error, "cannot use inverted pair to retrieve price: STETH/ETH")
}

func (suite *CurveSuite) TestFailOnWrongPair() {
	pair := value.Pair{Base: "x", Quote: "y"}

	suite.client.On(
		"ChainID",
		mock.Anything,
	).Return(uint64(1), nil)

	suite.client.On(
		"BlockNumber",
		mock.Anything,
	).Return(big.NewInt(100), nil).Once()

	points, err := suite.origin.FetchDataPoints(context.Background(), []any{pair})
	suite.Require().NoError(err)
	suite.Require().EqualError(points[pair].Error, "failed to get contract address for pair: x/y")
}
