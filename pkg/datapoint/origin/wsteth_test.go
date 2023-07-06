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

type WrappedStakedETHSuite struct {
	suite.Suite
	addresses ContractAddresses
	client    *ethereumMocks.RPC
	origin    *WrappedStakedETH
}

func (suite *WrappedStakedETHSuite) SetupTest() {
	suite.client = &ethereumMocks.RPC{}
	o, err := NewWrappedStakedETH(WrappedStakedETHConfig{
		Client: suite.client,
		ContractAddresses: map[string]string{
			"WSTETH/STETH": "0x7f39C581F595B53c5cb19bD0b3f8dA6c935E2Ca0",
		},
		Blocks: []int64{0, 10, 20},
		Logger: nil,
	})
	suite.NoError(err)
	suite.origin = o
}
func (suite *WrappedStakedETHSuite) TearDownTest() {
	suite.origin = nil
	suite.client = nil
}

func (suite *WrappedStakedETHSuite) Origin() *WrappedStakedETH {
	return suite.origin
}

func TestWrappedStakedETHSuite(t *testing.T) {
	suite.Run(t, new(WrappedStakedETHSuite))
}

func (suite *WrappedStakedETHSuite) TestSuccessResponse() {
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
			Input: hexutil.MustHexToBytes("252dba420000000000000000000000000000000000000000000000000000000000000020000000000000000000000000000000000000000000000000000000000000000100000000000000000000000000000000000000000000000000000000000000200000000000000000000000007f39c581f595b53c5cb19bd0b3f8da6c935e2ca000000000000000000000000000000000000000000000000000000000000000400000000000000000000000000000000000000000000000000000000000000004035faf8200000000000000000000000000000000000000000000000000000000"),
		},
		types.BlockNumberFromUint64(uint64(100)),
	).Return(respEncoded, nil).Once()

	respEncoded, _ = abi.EncodeValues(tuple, blockNumber.Uint64(), []any{resp[1]})
	suite.client.On(
		"Call",
		ctx,
		types.Call{
			To:    &contract,
			Input: hexutil.MustHexToBytes("252dba420000000000000000000000000000000000000000000000000000000000000020000000000000000000000000000000000000000000000000000000000000000100000000000000000000000000000000000000000000000000000000000000200000000000000000000000007f39c581f595b53c5cb19bd0b3f8da6c935e2ca000000000000000000000000000000000000000000000000000000000000000400000000000000000000000000000000000000000000000000000000000000004035faf8200000000000000000000000000000000000000000000000000000000"),
		},
		types.BlockNumberFromUint64(uint64(90)),
	).Return(respEncoded, nil).Once()

	respEncoded, _ = abi.EncodeValues(tuple, blockNumber.Uint64(), []any{resp[2]})
	suite.client.On(
		"Call",
		ctx,
		types.Call{
			To:    &contract,
			Input: hexutil.MustHexToBytes("252dba420000000000000000000000000000000000000000000000000000000000000020000000000000000000000000000000000000000000000000000000000000000100000000000000000000000000000000000000000000000000000000000000200000000000000000000000007f39c581f595b53c5cb19bd0b3f8da6c935e2ca000000000000000000000000000000000000000000000000000000000000000400000000000000000000000000000000000000000000000000000000000000004035faf8200000000000000000000000000000000000000000000000000000000"),
		},
		types.BlockNumberFromUint64(uint64(80)),
	).Return(respEncoded, nil).Once()

	pair := value.Pair{Base: "WSTETH", Quote: "STETH"}
	points, err := suite.origin.FetchDataPoints(ctx, []any{pair})
	suite.Require().NoError(err)
	suite.Equal(0.97, points[pair].Value.(value.Tick).Price.Float64())
	suite.Greater(points[pair].Time.Unix(), int64(0))
}

func (suite *WrappedStakedETHSuite) TestSuccessResponse_Inverted() {
	suite.client.On(
		"BlockNumber",
		mock.Anything,
	).Return(big.NewInt(100), nil).Once()

	pair := value.Pair{Base: "STETH", Quote: "WSTETH"}
	points, err := suite.origin.FetchDataPoints(context.Background(), []any{pair})
	suite.Require().NoError(err)
	suite.Require().EqualError(points[pair].Error, "cannot use inverted pair to retrieve price: STETH/WSTETH")
}

func (suite *WrappedStakedETHSuite) TestFailOnWrongPair() {
	pair := value.Pair{Base: "x", Quote: "y"}

	suite.client.On(
		"BlockNumber",
		mock.Anything,
	).Return(big.NewInt(100), nil).Once()

	points, err := suite.origin.FetchDataPoints(context.Background(), []any{pair})
	suite.Require().NoError(err)
	suite.Require().EqualError(points[pair].Error, "failed to get contract address for pair: x/y")
}
