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

type DSRSuite struct {
	suite.Suite
	addresses ContractAddresses
	client    *ethereumMocks.RPC
	origin    *DSR
}

func (suite *DSRSuite) SetupTest() {
	suite.client = &ethereumMocks.RPC{}
	o, err := NewDSR(DSRConfig{
		Client: suite.client,
		ContractAddresses: ContractAddresses{
			AssetPair{"DSR", "RATE"}: types.MustAddressFromHex("0x197E90f9FAD81970bA7976f33CbD77088E5D7cf7"),
		},
		Blocks: []int64{0, 10, 20},
		Logger: nil,
	})
	suite.NoError(err)
	suite.origin = o
}
func (suite *DSRSuite) TearDownTest() {
	suite.origin = nil
	suite.client = nil
}

func (suite *DSRSuite) Origin() *DSR {
	return suite.origin
}

func TestDSRSuite(t *testing.T) {
	suite.Run(t, new(DSRSuite))
}

func (suite *DSRSuite) TestSuccessResponse() {
	scaleUp := new(big.Int).Exp(big.NewInt(10), big.NewInt(25), nil)
	resp := [][]byte{
		types.Bytes(new(big.Int).Mul(big.NewInt(102), scaleUp).Bytes()).PadLeft(32),
		types.Bytes(new(big.Int).Mul(big.NewInt(103), scaleUp).Bytes()).PadLeft(32),
		types.Bytes(new(big.Int).Mul(big.NewInt(104), scaleUp).Bytes()).PadLeft(32),
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
			Input: hexutil.MustHexToBytes("252dba42000000000000000000000000000000000000000000000000000000000000002000000000000000000000000000000000000000000000000000000000000000010000000000000000000000000000000000000000000000000000000000000020000000000000000000000000197e90f9fad81970ba7976f33cbd77088e5d7cf700000000000000000000000000000000000000000000000000000000000000400000000000000000000000000000000000000000000000000000000000000004487bf08200000000000000000000000000000000000000000000000000000000"),
		},
		types.BlockNumberFromUint64(uint64(100)),
	).Return(respEncoded, nil).Twice()

	respEncoded, _ = abi.EncodeValues(tuple, blockNumber.Uint64(), []any{resp[1]})
	suite.client.On(
		"Call",
		ctx,
		types.Call{
			To:    &contract,
			Input: hexutil.MustHexToBytes("252dba42000000000000000000000000000000000000000000000000000000000000002000000000000000000000000000000000000000000000000000000000000000010000000000000000000000000000000000000000000000000000000000000020000000000000000000000000197e90f9fad81970ba7976f33cbd77088e5d7cf700000000000000000000000000000000000000000000000000000000000000400000000000000000000000000000000000000000000000000000000000000004487bf08200000000000000000000000000000000000000000000000000000000"),
		},
		types.BlockNumberFromUint64(uint64(90)),
	).Return(respEncoded, nil).Twice()

	respEncoded, _ = abi.EncodeValues(tuple, blockNumber.Uint64(), []any{resp[2]})
	suite.client.On(
		"Call",
		ctx,
		types.Call{
			To:    &contract,
			Input: hexutil.MustHexToBytes("252dba42000000000000000000000000000000000000000000000000000000000000002000000000000000000000000000000000000000000000000000000000000000010000000000000000000000000000000000000000000000000000000000000020000000000000000000000000197e90f9fad81970ba7976f33cbd77088e5d7cf700000000000000000000000000000000000000000000000000000000000000400000000000000000000000000000000000000000000000000000000000000004487bf08200000000000000000000000000000000000000000000000000000000"),
		},
		types.BlockNumberFromUint64(uint64(80)),
	).Return(respEncoded, nil).Twice()

	pair := value.Pair{Base: "DSR", Quote: "RATE"}
	points, err := suite.origin.FetchDataPoints(ctx, []any{pair})
	suite.Require().NoError(err)
	suite.Equal(1.03, points[pair].Value.(value.Tick).Price.Float64())
	suite.Greater(points[pair].Time.Unix(), int64(0))
}

func (suite *DSRSuite) TestFailOnWrongPair() {
	pair := value.Pair{Base: "x", Quote: "y"}

	suite.client.On(
		"BlockNumber",
		mock.Anything,
	).Return(big.NewInt(100), nil).Once()

	points, err := suite.origin.FetchDataPoints(context.Background(), []any{pair})
	suite.Require().NoError(err)
	suite.Require().EqualError(points[pair].Error, "failed to get contract address for pair: x/y")
}
