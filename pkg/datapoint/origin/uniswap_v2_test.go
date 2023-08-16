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

type UniswapV2Suite struct {
	suite.Suite
	client *ethereumMocks.RPC
	origin *UniswapV2
}

func (suite *UniswapV2Suite) SetupTest() {
	suite.client = &ethereumMocks.RPC{}
	o, err := NewUniswapV2(UniswapV2Config{
		Client: suite.client,
		ContractAddresses: ContractAddresses{
			AssetPair{"STETH", "WETH"}: types.MustAddressFromHex("0x4028DAAC072e492d34a3Afdbef0ba7e35D8b55C4"),
		},
		Blocks: []int64{0, 10, 20},
		Logger: nil,
	})
	suite.NoError(err)
	suite.origin = o
}

func (suite *UniswapV2Suite) TearDownTest() {
	suite.origin = nil
	suite.client = nil
}

func (suite *UniswapV2Suite) Origin() *UniswapV2 {
	return suite.origin
}

func TestUniswapV2Suite(t *testing.T) {
	suite.Run(t, new(UniswapV2Suite))
}

func (suite *UniswapV2Suite) TestSuccessResponse() {
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

	// token0(), token1()
	tokens := [][]byte{
		types.Bytes(types.MustAddressFromHex("0xae7ab96520DE3A18E5e111B5EaAb095312D7fE84").Bytes()).PadLeft(32), // stETH
		types.Bytes(types.MustAddressFromHex("0xC02aaA39b223FE8D0A0e5C4F27eAD9083C756Cc2").Bytes()).PadLeft(32), // WETH
	}
	respEncoded := abi.MustEncodeValues(tuple, blockNumber.Uint64(), []any{tokens[0], tokens[1]})
	suite.client.On(
		"Call",
		ctx,
		types.Call{
			To:    &contract,
			Input: hexutil.MustHexToBytes("252dba4200000000000000000000000000000000000000000000000000000000000000200000000000000000000000000000000000000000000000000000000000000002000000000000000000000000000000000000000000000000000000000000004000000000000000000000000000000000000000000000000000000000000000c00000000000000000000000004028daac072e492d34a3afdbef0ba7e35d8b55c4000000000000000000000000000000000000000000000000000000000000004000000000000000000000000000000000000000000000000000000000000000040dfe1681000000000000000000000000000000000000000000000000000000000000000000000000000000004028daac072e492d34a3afdbef0ba7e35d8b55c400000000000000000000000000000000000000000000000000000000000000400000000000000000000000000000000000000000000000000000000000000004d21220a700000000000000000000000000000000000000000000000000000000"),
		},
		mock.Anything,
	).Return(respEncoded, nil).Twice()

	// stETH.symbol(), stETH.decimals(), WETH.symbol(), WETH.decimals()
	symbolAbi := abi.MustParseType("(string memory)")
	symbolMap := make(map[string]string)
	symbolMap["arg0"] = "stETH"
	stETHBytes := abi.MustEncodeValue(symbolAbi, symbolMap)
	symbolMap["arg0"] = "WETH"
	WETHBytes := abi.MustEncodeValue(symbolAbi, symbolMap)
	respEncoded = abi.MustEncodeValues(tuple, blockNumber.Uint64(), []any{
		stETHBytes,
		types.Bytes(big.NewInt(18).Bytes()).PadLeft(32),
		WETHBytes,
		types.Bytes(big.NewInt(18).Bytes()).PadLeft(32),
	})
	suite.client.On(
		"Call",
		ctx,
		types.Call{
			To:    &contract,
			Input: hexutil.MustHexToBytes("252dba42000000000000000000000000000000000000000000000000000000000000002000000000000000000000000000000000000000000000000000000000000000040000000000000000000000000000000000000000000000000000000000000080000000000000000000000000000000000000000000000000000000000000010000000000000000000000000000000000000000000000000000000000000001800000000000000000000000000000000000000000000000000000000000000200000000000000000000000000ae7ab96520de3a18e5e111b5eaab095312d7fe840000000000000000000000000000000000000000000000000000000000000040000000000000000000000000000000000000000000000000000000000000000495d89b4100000000000000000000000000000000000000000000000000000000000000000000000000000000ae7ab96520de3a18e5e111b5eaab095312d7fe8400000000000000000000000000000000000000000000000000000000000000400000000000000000000000000000000000000000000000000000000000000004313ce56700000000000000000000000000000000000000000000000000000000000000000000000000000000c02aaa39b223fe8d0a0e5c4f27ead9083c756cc20000000000000000000000000000000000000000000000000000000000000040000000000000000000000000000000000000000000000000000000000000000495d89b4100000000000000000000000000000000000000000000000000000000000000000000000000000000c02aaa39b223fe8d0a0e5c4f27ead9083c756cc200000000000000000000000000000000000000000000000000000000000000400000000000000000000000000000000000000000000000000000000000000004313ce56700000000000000000000000000000000000000000000000000000000"),
		},
		mock.Anything,
	).Return(respEncoded, nil).Twice()

	// getReserves()
	reservesMap := make(map[string]*big.Int)
	reservesMap["_reserve0"] = new(big.Int).Mul(big.NewInt(100), big.NewInt(int64(ether)))
	reservesMap["_reserve1"] = new(big.Int).Mul(big.NewInt(200), big.NewInt(int64(ether)))
	reservesMap["_blockTimestampLast"] = big.NewInt(1692188531)
	reserves100Bytes := abi.MustEncodeValue(getReserves.Outputs(), reservesMap)
	respEncoded = abi.MustEncodeValues(tuple, blockNumber.Uint64(), []any{reserves100Bytes})
	suite.client.On(
		"Call",
		ctx,
		types.Call{
			To:    &contract,
			Input: hexutil.MustHexToBytes("252dba420000000000000000000000000000000000000000000000000000000000000020000000000000000000000000000000000000000000000000000000000000000100000000000000000000000000000000000000000000000000000000000000200000000000000000000000004028daac072e492d34a3afdbef0ba7e35d8b55c4000000000000000000000000000000000000000000000000000000000000004000000000000000000000000000000000000000000000000000000000000000040902f1ac00000000000000000000000000000000000000000000000000000000"),
		},
		types.BlockNumberFromUint64(uint64(100)),
	).Return(respEncoded, nil).Twice()

	reservesMap["_reserve0"] = new(big.Int).Mul(big.NewInt(90), big.NewInt(int64(ether)))
	reservesMap["_reserve1"] = new(big.Int).Mul(big.NewInt(210), big.NewInt(int64(ether)))
	reserves90Bytes := abi.MustEncodeValue(getReserves.Outputs(), reservesMap)
	respEncoded = abi.MustEncodeValues(tuple, blockNumber.Uint64(), []any{reserves90Bytes})
	suite.client.On(
		"Call",
		ctx,
		types.Call{
			To:    &contract,
			Input: hexutil.MustHexToBytes("252dba420000000000000000000000000000000000000000000000000000000000000020000000000000000000000000000000000000000000000000000000000000000100000000000000000000000000000000000000000000000000000000000000200000000000000000000000004028daac072e492d34a3afdbef0ba7e35d8b55c4000000000000000000000000000000000000000000000000000000000000004000000000000000000000000000000000000000000000000000000000000000040902f1ac00000000000000000000000000000000000000000000000000000000"),
		},
		types.BlockNumberFromUint64(uint64(90)),
	).Return(respEncoded, nil).Twice()

	reservesMap["_reserve0"] = new(big.Int).Mul(big.NewInt(80), big.NewInt(int64(ether)))
	reservesMap["_reserve1"] = new(big.Int).Mul(big.NewInt(220), big.NewInt(int64(ether)))
	reserves80Bytes := abi.MustEncodeValue(getReserves.Outputs(), reservesMap)
	respEncoded = abi.MustEncodeValues(tuple, blockNumber.Uint64(), []any{reserves80Bytes})
	suite.client.On(
		"Call",
		ctx,
		types.Call{
			To:    &contract,
			Input: hexutil.MustHexToBytes("252dba420000000000000000000000000000000000000000000000000000000000000020000000000000000000000000000000000000000000000000000000000000000100000000000000000000000000000000000000000000000000000000000000200000000000000000000000004028daac072e492d34a3afdbef0ba7e35d8b55c4000000000000000000000000000000000000000000000000000000000000004000000000000000000000000000000000000000000000000000000000000000040902f1ac00000000000000000000000000000000000000000000000000000000"),
		},
		types.BlockNumberFromUint64(uint64(80)),
	).Return(respEncoded, nil).Twice()

	pair := value.Pair{Base: "STETH", Quote: "WETH"}
	points, err := suite.origin.FetchDataPoints(context.Background(), []any{pair})
	suite.Require().NoError(err)
	const expectedPrice = (200/100.0 + 210/90.0 + 220/80.0) / 3
	suite.Equal(expectedPrice, points[pair].Value.(value.Tick).Price.Float64())
	suite.Greater(points[pair].Time.Unix(), int64(0))

	pair = value.Pair{Base: "WETH", Quote: "STETH"}
	points, err = suite.origin.FetchDataPoints(context.Background(), []any{pair})
	suite.Require().NoError(err)
	const expectedInvertPrice = (100/200.0 + 90/210.0 + 80/220.0) / 3
	suite.Equal(expectedInvertPrice, points[pair].Value.(value.Tick).Price.Float64())
	suite.Greater(points[pair].Time.Unix(), int64(0))
}

func (suite *UniswapV2Suite) TestFailOnWrongPair() {
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
