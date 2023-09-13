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
	"github.com/chronicleprotocol/oracle-suite/pkg/util/bn"
)

type BalancerV2Suite struct {
	suite.Suite
	addresses ContractAddresses
	client    *ethereumMocks.RPC
	origin    *BalancerV2
}

func (suite *BalancerV2Suite) SetupTest() {
	suite.client = &ethereumMocks.RPC{}
	o, err := NewBalancerV2(BalancerV2Config{
		Client: suite.client,
		ContractAddresses: ContractAddresses{
			AssetPair{"RETH", "WETH"}:  types.MustAddressFromHex("0x1E19CF2D73a72Ef1332C882F20534B6519Be0276"),
			AssetPair{"STETH", "WETH"}: types.MustAddressFromHex("0x32296969ef14eb0c6d29669c550d4a0449130230"),
			AssetPair{"WETH", "YFI"}:   types.MustAddressFromHex("0x186084ff790c65088ba694df11758fae4943ee9e"),
		},
		ReferenceAddresses: ContractAddresses{
			AssetPair{"RETH", "WETH"}: types.MustAddressFromHex("0xae78736Cd615f374D3085123A210448E74Fc6393"),
		},
		Blocks: []int64{0, 10, 20},
		Logger: nil,
	})
	suite.NoError(err)
	suite.origin = o
}

func (suite *BalancerV2Suite) TearDownTest() {
	suite.origin = nil
	suite.client = nil
}

func (suite *BalancerV2Suite) Origin() *BalancerV2 {
	return suite.origin
}

func TestBalancerV2Suite(t *testing.T) {
	suite.Run(t, new(BalancerV2Suite))
}

func (suite *BalancerV2Suite) TestSuccessResponse() {
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

	// Mock for `getLatest`
	respEncoded, _ := abi.EncodeValues(tuple, blockNumber.Uint64(), []any{resp[0]})
	suite.client.On(
		"Call",
		ctx,
		types.Call{
			To:    &contract,
			Input: hexutil.MustHexToBytes("0x252dba4200000000000000000000000000000000000000000000000000000000000000200000000000000000000000000000000000000000000000000000000000000001000000000000000000000000000000000000000000000000000000000000002000000000000000000000000032296969ef14eb0c6d29669c550d4a044913023000000000000000000000000000000000000000000000000000000000000000400000000000000000000000000000000000000000000000000000000000000024b10be739000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000"),
		},
		types.BlockNumberFromUint64(uint64(100)),
	).Return(respEncoded, &types.Call{}, nil).Twice()

	// Mock for `getLatest`
	respEncoded, _ = abi.EncodeValues(tuple, blockNumber.Uint64(), []any{resp[1]})
	suite.client.On(
		"Call",
		ctx,
		types.Call{
			To:    &contract,
			Input: hexutil.MustHexToBytes("0x252dba4200000000000000000000000000000000000000000000000000000000000000200000000000000000000000000000000000000000000000000000000000000001000000000000000000000000000000000000000000000000000000000000002000000000000000000000000032296969ef14eb0c6d29669c550d4a044913023000000000000000000000000000000000000000000000000000000000000000400000000000000000000000000000000000000000000000000000000000000024b10be739000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000"),
		},
		types.BlockNumberFromUint64(uint64(90)),
	).Return(respEncoded, &types.Call{}, nil).Twice()

	// Mock for `getLatest`
	respEncoded, _ = abi.EncodeValues(tuple, blockNumber.Uint64(), []any{resp[2]})
	suite.client.On(
		"Call",
		ctx,
		types.Call{
			To:    &contract,
			Input: hexutil.MustHexToBytes("0x252dba4200000000000000000000000000000000000000000000000000000000000000200000000000000000000000000000000000000000000000000000000000000001000000000000000000000000000000000000000000000000000000000000002000000000000000000000000032296969ef14eb0c6d29669c550d4a044913023000000000000000000000000000000000000000000000000000000000000000400000000000000000000000000000000000000000000000000000000000000024b10be739000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000"),
		},
		types.BlockNumberFromUint64(uint64(80)),
	).Return(respEncoded, &types.Call{}, nil).Twice()

	pair := value.Pair{Base: "STETH", Quote: "WETH"}
	points, err := suite.origin.FetchDataPoints(ctx, []any{pair})
	suite.Require().NoError(err)
	suite.Equal(bn.Float(0.97).String(), points[pair].Value.(value.Tick).Price.Float().String())
	suite.Greater(points[pair].Time.Unix(), int64(0))

	// Inverted pair
	pair = value.Pair{Base: "WETH", Quote: "STETH"}
	points, err = suite.origin.FetchDataPoints(context.Background(), []any{pair})
	suite.Require().NoError(err)
	suite.Equal(bn.Float(1/0.97).String(), points[pair].Value.(value.Tick).Price.Float().String())
	suite.Greater(points[pair].Time.Unix(), int64(0))
}

func (suite *BalancerV2Suite) TestSuccessResponseWithRef() {
	resp := [][]byte{
		types.Bytes(big.NewInt(0.94 * ether).Bytes()).PadLeft(32),
		types.Bytes(big.NewInt(0.98 * ether).Bytes()).PadLeft(32),
		types.Bytes(big.NewInt(0.99 * ether).Bytes()).PadLeft(32),
	}

	resp2 := [][]byte{
		types.Bytes(big.NewInt(0.2 * ether).Bytes()).PadLeft(32),
		types.Bytes(big.NewInt(0.6 * ether).Bytes()).PadLeft(32),
		types.Bytes(big.NewInt(0.7 * ether).Bytes()).PadLeft(32),
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

	// Mock for `getLatest`, `getPriceRateCache`
	respEncoded, _ := abi.EncodeValues(tuple, blockNumber.Uint64(), []any{resp[0], resp2[0]})
	suite.client.On(
		"Call",
		ctx,
		types.Call{
			To:    &contract,
			Input: hexutil.MustHexToBytes("0x252dba4200000000000000000000000000000000000000000000000000000000000000200000000000000000000000000000000000000000000000000000000000000002000000000000000000000000000000000000000000000000000000000000004000000000000000000000000000000000000000000000000000000000000000e00000000000000000000000001e19cf2d73a72ef1332c882f20534b6519be027600000000000000000000000000000000000000000000000000000000000000400000000000000000000000000000000000000000000000000000000000000024b10be7390000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000001e19cf2d73a72ef1332c882f20534b6519be027600000000000000000000000000000000000000000000000000000000000000400000000000000000000000000000000000000000000000000000000000000024b867ee5a000000000000000000000000ae78736cd615f374d3085123a210448e74fc639300000000000000000000000000000000000000000000000000000000"),
		},
		types.BlockNumberFromUint64(uint64(100)),
	).Return(respEncoded, &types.Call{}, nil).Twice()

	// Mock for `getLatest`, `getPriceRateCache`
	respEncoded, _ = abi.EncodeValues(tuple, blockNumber.Uint64(), []any{resp[1], resp2[1]})
	suite.client.On(
		"Call",
		ctx,
		types.Call{
			To:    &contract,
			Input: hexutil.MustHexToBytes("0x252dba4200000000000000000000000000000000000000000000000000000000000000200000000000000000000000000000000000000000000000000000000000000002000000000000000000000000000000000000000000000000000000000000004000000000000000000000000000000000000000000000000000000000000000e00000000000000000000000001e19cf2d73a72ef1332c882f20534b6519be027600000000000000000000000000000000000000000000000000000000000000400000000000000000000000000000000000000000000000000000000000000024b10be7390000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000001e19cf2d73a72ef1332c882f20534b6519be027600000000000000000000000000000000000000000000000000000000000000400000000000000000000000000000000000000000000000000000000000000024b867ee5a000000000000000000000000ae78736cd615f374d3085123a210448e74fc639300000000000000000000000000000000000000000000000000000000"),
		},
		types.BlockNumberFromUint64(uint64(90)),
	).Return(respEncoded, &types.Call{}, nil).Twice()

	// Mock for `getLatest`, `getPriceRateCache`
	respEncoded, _ = abi.EncodeValues(tuple, blockNumber.Uint64(), []any{resp[2], resp2[2]})
	suite.client.On(
		"Call",
		ctx,
		types.Call{
			To:    &contract,
			Input: hexutil.MustHexToBytes("0x252dba4200000000000000000000000000000000000000000000000000000000000000200000000000000000000000000000000000000000000000000000000000000002000000000000000000000000000000000000000000000000000000000000004000000000000000000000000000000000000000000000000000000000000000e00000000000000000000000001e19cf2d73a72ef1332c882f20534b6519be027600000000000000000000000000000000000000000000000000000000000000400000000000000000000000000000000000000000000000000000000000000024b10be7390000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000001e19cf2d73a72ef1332c882f20534b6519be027600000000000000000000000000000000000000000000000000000000000000400000000000000000000000000000000000000000000000000000000000000024b867ee5a000000000000000000000000ae78736cd615f374d3085123a210448e74fc639300000000000000000000000000000000000000000000000000000000"),
		},
		types.BlockNumberFromUint64(uint64(80)),
	).Return(respEncoded, &types.Call{}, nil).Twice()

	pair := value.Pair{Base: "RETH", Quote: "WETH"}
	points, err := suite.origin.FetchDataPoints(ctx, []any{pair})
	suite.Require().NoError(err)
	suite.Equal(bn.Float(0.485).String(), points[pair].Value.(value.Tick).Price.Float().String())
	suite.Greater(points[pair].Time.Unix(), int64(0))

	// Inverted pair
	pair = value.Pair{Base: "WETH", Quote: "RETH"}
	points, err = suite.origin.FetchDataPoints(context.Background(), []any{pair})
	suite.Require().NoError(err)
	suite.Equal(bn.Float(1/0.485).String(), points[pair].Value.(value.Tick).Price.Float().String())
	suite.Greater(points[pair].Time.Unix(), int64(0))
}

func (suite *BalancerV2Suite) TestFailOnWrongPair() {
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
