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

type RocketPoolSuite struct {
	suite.Suite
	addresses map[string]string
	client    *ethereumMocks.RPC
	origin    *RocketPool
}

func (suite *RocketPoolSuite) SetupTest() {
	suite.client = &ethereumMocks.RPC{}
	o, err := NewRocketPool(RocketPoolConfig{
		Client: suite.client,
		ContractAddresses: ContractAddresses{
			AssetPair{"RETH", "ETH"}: types.MustAddressFromHex("0xae78736Cd615f374D3085123A210448E74Fc6393"),
		},
		Blocks: []int64{0, 10, 20},
		Logger: nil,
	})
	suite.NoError(err)
	suite.origin = o
}
func (suite *RocketPoolSuite) TearDownTest() {
	suite.origin = nil
	suite.client = nil
}

func (suite *RocketPoolSuite) Origin() *RocketPool {
	return suite.origin
}

func TestRocketPoolSuite(t *testing.T) {
	suite.Run(t, new(RocketPoolSuite))
}

func (suite *RocketPoolSuite) TestSuccessResponse() {
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
			Input: hexutil.MustHexToBytes("252dba42000000000000000000000000000000000000000000000000000000000000002000000000000000000000000000000000000000000000000000000000000000010000000000000000000000000000000000000000000000000000000000000020000000000000000000000000ae78736cd615f374d3085123a210448e74fc639300000000000000000000000000000000000000000000000000000000000000400000000000000000000000000000000000000000000000000000000000000004e6aa216c00000000000000000000000000000000000000000000000000000000"),
		},
		types.BlockNumberFromUint64(uint64(100)),
	).Return(respEncoded, &types.Call{}, nil).Twice()

	respEncoded, _ = abi.EncodeValues(tuple, blockNumber.Uint64(), []any{resp[1]})
	suite.client.On(
		"Call",
		ctx,
		types.Call{
			To:    &contract,
			Input: hexutil.MustHexToBytes("252dba42000000000000000000000000000000000000000000000000000000000000002000000000000000000000000000000000000000000000000000000000000000010000000000000000000000000000000000000000000000000000000000000020000000000000000000000000ae78736cd615f374d3085123a210448e74fc639300000000000000000000000000000000000000000000000000000000000000400000000000000000000000000000000000000000000000000000000000000004e6aa216c00000000000000000000000000000000000000000000000000000000"),
		},
		types.BlockNumberFromUint64(uint64(90)),
	).Return(respEncoded, &types.Call{}, nil).Twice()

	respEncoded, _ = abi.EncodeValues(tuple, blockNumber.Uint64(), []any{resp[2]})
	suite.client.On(
		"Call",
		ctx,
		types.Call{
			To:    &contract,
			Input: hexutil.MustHexToBytes("252dba42000000000000000000000000000000000000000000000000000000000000002000000000000000000000000000000000000000000000000000000000000000010000000000000000000000000000000000000000000000000000000000000020000000000000000000000000ae78736cd615f374d3085123a210448e74fc639300000000000000000000000000000000000000000000000000000000000000400000000000000000000000000000000000000000000000000000000000000004e6aa216c00000000000000000000000000000000000000000000000000000000"),
		},
		types.BlockNumberFromUint64(uint64(80)),
	).Return(respEncoded, &types.Call{}, nil).Twice()

	pair := value.Pair{Base: "RETH", Quote: "ETH"}
	point, err := suite.origin.FetchDataPoints(ctx, []any{pair})
	suite.Require().NoError(err)
	suite.Equal(0.97, point[pair].Value.(value.Tick).Price.Float64())
	suite.Greater(point[pair].Time.Unix(), int64(0))

	pair = value.Pair{Base: "ETH", Quote: "RETH"}
	point, err = suite.origin.FetchDataPoints(ctx, []any{pair})
	suite.Require().NoError(err)
	suite.Equal(1/0.97, point[pair].Value.(value.Tick).Price.Float64())
	suite.Greater(point[pair].Time.Unix(), int64(0))
}

func (suite *RocketPoolSuite) TestFailOnWrongPair() {
	pair := value.Pair{Base: "x", Quote: "y"}

	suite.client.On(
		"BlockNumber",
		mock.Anything,
	).Return(big.NewInt(100), nil).Once()

	points, err := suite.origin.FetchDataPoints(context.Background(), []any{pair})
	suite.Require().NoError(err)
	suite.Require().EqualError(points[pair].Error, "failed to get contract address for pair: x/y")
}
