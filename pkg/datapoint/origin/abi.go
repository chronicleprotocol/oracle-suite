package origin

import "github.com/defiweb/go-eth/abi"

// [Balancer V2]
var getLatest = abi.MustParseMethod("getLatest(uint8)(uint256)")
var getPriceRateCache = abi.MustParseMethod("getPriceRateCache(address)(uint256,uint256,uint256)")
var getPoolID = abi.MustParseMethod("getPoolId()(bytes32)")
var getVault = abi.MustParseMethod("getVault()(address)")
var getPoolTokens = abi.MustParseMethod("getPoolTokens(bytes32)(address[] memory tokens,uint256[] memory balances,uint256 lastChangeBlock)")
var getSwapFeePercentage = abi.MustParseMethod("getSwapFeePercentage()(uint256)")
var getScalingFactors = abi.MustParseMethod("getScalingFactors()(uint256[] memory)")
var getNormalizedWeights = abi.MustParseMethod("getNormalizedWeights()(uint256[] memory)")
var getBptIndex = abi.MustParseMethod("getBptIndex()(uint256)")

// [Balancer V2 - ComposableStablePool]
var getAmplificationParameter = abi.MustParseMethod("getAmplificationParameter()(uint256 value,bool isUpdating,uint256 precision)")
var getLastJoinExitData = abi.MustParseMethod("getLastJoinExitData()(uint256 lastJoinExitAmplification,uint256 lastPostJoinExitInvariant)")
var getTotalSupply = abi.MustParseMethod("totalSupply()(uint256)")
var getRateProviders = abi.MustParseMethod("getRateProviders()(address[] memory)")
var getProtocolFeePercentageCache = abi.MustParseMethod("getProtocolFeePercentageCache(uint256)(uint256)")
var isTokenExemptFromYieldProtocolFee = abi.MustParseMethod("isTokenExemptFromYieldProtocolFee(address)(bool)")
var getTokenRateCache = abi.MustParseMethod("getTokenRateCache(address)(uint256 rate,uint256 oldRate,uint256 duration,uint256 expires)")

// [Curve]
// Since curve has `stableswap` pool and `cryptoswap` pool, and their smart contracts have pretty similar interface
// `stableswap` pool is using `int128` in `get_dy`, `get_dx` ...,
// while `cryptoswap` pool is using `uint256` in `get_dy`, `get_dx`, ...
var getDy1 = abi.MustParseMethod("get_dy(int128,int128,uint256)(uint256)")
var getDy2 = abi.MustParseMethod("get_dy(uint256,uint256,uint256)(uint256)")
var coins = abi.MustParseMethod("coins(uint256)(address)")

// [dsr]
var dsr = abi.MustParseMethod("dsr()(uint256)")

// [RocketPool]
var getExchangeRate = abi.MustParseMethod("getExchangeRate()(uint256)")

// [sDAI]
var previewRedeem = abi.MustParseMethod("previewRedeem(uint256)(uint256)")

// [Sushiswap]
var getReserves = abi.MustParseMethod("getReserves()(uint112 _reserve0,uint112 _reserve1,uint32 _blockTimestampLast)")
var token0Abi = abi.MustParseMethod("token0()(address)")
var token1Abi = abi.MustParseMethod("token1()(address)")

// [Uniswap v3]
var slot0 = abi.MustParseMethod("slot0()(uint160,int24,uint16,uint16,uint16,uint8,bool)")

// var token0Abi = abi.MustParseMethod("token0()(address)")
// var token1Abi = abi.MustParseMethod("token1()(address)")

// [Uniswap v2]
// var getReserves =
//     abi.MustParseMethod("getReserves()(uint112 _reserve0,uint112 _reserve1,uint32 _blockTimestampLast)")
// var token0Abi = abi.MustParseMethod("token0()(address)")
// var token1Abi = abi.MustParseMethod("token1()(address)")

// [wstETH]
var stEthPerToken = abi.MustParseMethod("stEthPerToken()(uint256)")
