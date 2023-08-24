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

type SDAISuite struct {
	suite.Suite
	addresses ContractAddresses
	client    *ethereumMocks.RPC
	origin    *SDAI
}

func (suite *SDAISuite) SetupTest() {
	suite.client = &ethereumMocks.RPC{}
	o, err := NewSDAI(SDAIConfig{
		Client: suite.client,
		ContractAddresses: ContractAddresses{
			AssetPair{"SDAI", "DAI"}: types.MustAddressFromHex("0x83F20F44975D03b1b09e64809B757c47f942BEeA"),
		},
		Blocks: []int64{0, 10, 20},
		Logger: nil,
	})
	suite.NoError(err)
	suite.origin = o
}
func (suite *SDAISuite) TearDownTest() {
	suite.origin = nil
	suite.client = nil
}

func (suite *SDAISuite) Origin() *SDAI {
	return suite.origin
}

func TestSDAISuite(t *testing.T) {
	suite.Run(t, new(SDAISuite))
}

func (suite *SDAISuite) TestSuccessResponse() {
	resp := [][]byte{
		types.Bytes(big.NewInt(1.02 * ether).Bytes()).PadLeft(32),
		types.Bytes(big.NewInt(1.03 * ether).Bytes()).PadLeft(32),
		types.Bytes(big.NewInt(1.04 * ether).Bytes()).PadLeft(32),
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
			Input: hexutil.MustHexToBytes("252dba4200000000000000000000000000000000000000000000000000000000000000200000000000000000000000000000000000000000000000000000000000000001000000000000000000000000000000000000000000000000000000000000002000000000000000000000000083f20f44975d03b1b09e64809b757c47f942beea000000000000000000000000000000000000000000000000000000000000004000000000000000000000000000000000000000000000000000000000000000244cdad5060000000000000000000000000000000000000000000000000de0b6b3a764000000000000000000000000000000000000000000000000000000000000"),
		},
		types.BlockNumberFromUint64(uint64(100)),
	).Return(respEncoded, nil).Twice()

	respEncoded, _ = abi.EncodeValues(tuple, blockNumber.Uint64(), []any{resp[1]})
	suite.client.On(
		"Call",
		ctx,
		types.Call{
			To:    &contract,
			Input: hexutil.MustHexToBytes("252dba4200000000000000000000000000000000000000000000000000000000000000200000000000000000000000000000000000000000000000000000000000000001000000000000000000000000000000000000000000000000000000000000002000000000000000000000000083f20f44975d03b1b09e64809b757c47f942beea000000000000000000000000000000000000000000000000000000000000004000000000000000000000000000000000000000000000000000000000000000244cdad5060000000000000000000000000000000000000000000000000de0b6b3a764000000000000000000000000000000000000000000000000000000000000"),
		},
		types.BlockNumberFromUint64(uint64(90)),
	).Return(respEncoded, nil).Twice()

	respEncoded, _ = abi.EncodeValues(tuple, blockNumber.Uint64(), []any{resp[2]})
	suite.client.On(
		"Call",
		ctx,
		types.Call{
			To:    &contract,
			Input: hexutil.MustHexToBytes("252dba4200000000000000000000000000000000000000000000000000000000000000200000000000000000000000000000000000000000000000000000000000000001000000000000000000000000000000000000000000000000000000000000002000000000000000000000000083f20f44975d03b1b09e64809b757c47f942beea000000000000000000000000000000000000000000000000000000000000004000000000000000000000000000000000000000000000000000000000000000244cdad5060000000000000000000000000000000000000000000000000de0b6b3a764000000000000000000000000000000000000000000000000000000000000"),
		},
		types.BlockNumberFromUint64(uint64(80)),
	).Return(respEncoded, nil).Twice()

	pair := value.Pair{Base: "SDAI", Quote: "DAI"}
	points, err := suite.origin.FetchDataPoints(ctx, []any{pair})
	suite.Require().NoError(err)
	suite.Equal(1.03, points[pair].Value.(value.Tick).Price.Float64())
	suite.Greater(points[pair].Time.Unix(), int64(0))

	pair = value.Pair{Base: "DAI", Quote: "SDAI"}
	points, err = suite.origin.FetchDataPoints(ctx, []any{pair})
	suite.Require().NoError(err)
	suite.Equal(1/1.03, points[pair].Value.(value.Tick).Price.Float64())
	suite.Greater(points[pair].Time.Unix(), int64(0))
}

func (suite *SDAISuite) TestFailOnWrongPair() {
	pair := value.Pair{Base: "x", Quote: "y"}

	suite.client.On(
		"BlockNumber",
		mock.Anything,
	).Return(big.NewInt(100), nil).Once()

	points, err := suite.origin.FetchDataPoints(context.Background(), []any{pair})
	suite.Require().NoError(err)
	suite.Require().EqualError(points[pair].Error, "failed to get contract address for pair: x/y")
}
