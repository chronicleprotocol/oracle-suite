package origin

import "github.com/defiweb/go-eth/abi"

// [Balancer V2]
var getLatest = abi.MustParseMethod("getLatest(uint8)(uint256)")
var getPriceRateCache = abi.MustParseMethod("getPriceRateCache(address)(uint256,uint256,uint256)")

// [Curve]
// Since curve has `stableswap` pool and `cryptoswap` pool, and their smart contracts have pretty similar interface
// `stableswap` pool is using `int128` in `get_dy`, `get_dx` ...,
// while `cryptoswap` pool is using `uint256` in `get_dy`, `get_dx`, ...
var getDy1 = abi.MustParseMethod("get_dy(int128,int128,uint256)(uint256)")
var getDy2 = abi.MustParseMethod("get_dy(uint256,uint256,uint256)(uint256)")
var coins = abi.MustParseMethod("coins(uint256)(address)")

// [RocketPool]
var getExchangeRate = abi.MustParseMethod("getExchangeRate()(uint256)")

// [sDAI]
var previewRedeem = abi.MustParseMethod("previewRedeem(uint256)(uint256)")

// [Sushiswap]
var getReserves = abi.MustParseMethod("getReserves()(uint112,uint112,uint32)")
var token0Abi = abi.MustParseMethod("token0()(address)")
var token1Abi = abi.MustParseMethod("token1()(address)")

// [Uniswap]
var slot0 = abi.MustParseMethod("slot0()(uint160,int24,uint16,uint16,uint16,uint8,bool)")

// var token0Abi = abi.MustParseMethod("token0()(address)")
// var token1Abi = abi.MustParseMethod("token1()(address)")

// [wstETH]
var stEthPerToken = abi.MustParseMethod("stEthPerToken()(uint256)")
