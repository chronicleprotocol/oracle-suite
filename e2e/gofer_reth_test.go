package e2e

import (
	"context"
	"math/big"
	"sort"
	"testing"
	"time"

	"github.com/defiweb/go-eth/types"

	"github.com/chronicleprotocol/infestor"
	"github.com/chronicleprotocol/infestor/origin"
	"github.com/chronicleprotocol/infestor/smocker"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func Test_Gofer_Balancer_RETH2WETH(t *testing.T) {
	ctx, ctxCancel := context.WithTimeout(context.Background(), 2*time.Minute)
	defer ctxCancel()

	s := smocker.NewAPI("http://127.0.0.1:8081")
	require.NoError(t, s.Reset(ctx))

	const blockNumber int = 100
	getLatest_RETH2ETH := 0.98
	getPriceRateCache_RETH2ETH := 0.4

	balancer_RETH2WETH := getLatest_RETH2ETH * getPriceRateCache_RETH2ETH
	err := infestor.NewMocksBuilder().
		//origin "balancerV2" { query = "RETH/WETH" }
		Add(origin.NewExchange("balancerV2").
			WithSymbol("RETH/WETH").
			WithFunctionData("getLatest", []origin.FunctionData{
				{
					Address: types.MustAddressFromHex("0x1E19CF2D73a72Ef1332C882F20534B6519Be0276"), // RETH/WETH
					Args:    []any{byte(0)},
					Return:  []any{big.NewInt(int64(getLatest_RETH2ETH * 1e18))},
				},
			}).
			WithFunctionData("getPriceRateCache", []origin.FunctionData{
				{
					Address: types.MustAddressFromHex("0x1E19CF2D73a72Ef1332C882F20534B6519Be0276"), // RETH/WETH
					Args:    []any{types.MustAddressFromHex("0xae78736Cd615f374D3085123A210448E74Fc6393")},
					Return:  []any{big.NewInt(int64(getPriceRateCache_RETH2ETH * 1e18)), big.NewInt(10), big.NewInt(time.Now().Unix())},
				},
			}).
			WithCustom("blockNumber", blockNumber)).
		Deploy(*s)
	require.NoError(t, err)

	mocks := []*smocker.Mock{
		smocker.NewMockBuilder().
			AddResponseHeader("Content-Type", "application/json").
			SetRequestBodyString(smocker.ShouldContainSubstring("eth_chainId")).
			SetResponseBody(mustReadFile("./testdata/mock/eth_chainId.json")).
			Mock(),
	}
	require.NoError(t, s.AddMocks(ctx, mocks))

	out, err := execCommand(ctx, "..", nil, "./gofer", "-c", "./e2e/testdata/config/gofer.hcl", "-v", "debug", "price", "RETH/WETH", "--format", "json")
	require.NoError(t, err)

	priceMap, err := parseGoferPrice(out)
	require.NoError(t, err)
	p := priceMap["RETH/WETH"]

	assert.Equal(t, "reference", p.Meta["type"])
	assert.Equal(t, balancer_RETH2WETH, p.Value.Price)
	assert.Greater(t, len(p.SubPoints), 0)
}

func Test_Gofer_Curve_RETH2WSTETH(t *testing.T) {
	ctx, ctxCancel := context.WithTimeout(context.Background(), 2*time.Minute)
	defer ctxCancel()

	s := smocker.NewAPI("http://127.0.0.1:8081")
	require.NoError(t, s.Reset(ctx))

	const blockNumber int = 100
	getDy_RETH2WSTETH := float64(3)
	err := infestor.NewMocksBuilder().
		// origin "curve" { query = "RETH/WSTETH" }
		Add(origin.NewExchange("curve").
			WithSymbol("RETH/WSTETH").
			WithCustom("RETH/WSTETH", types.MustAddressFromHex("0x447Ddd4960d9fdBF6af9a790560d0AF76795CB08")).
			WithFunctionData("coins", []origin.FunctionData{
				{
					Address: types.MustAddressFromHex("0x447Ddd4960d9fdBF6af9a790560d0AF76795CB08"), // RETH/WSTETH
					Args:    []any{0},
					Return:  []any{types.MustAddressFromHex("0xae78736Cd615f374D3085123A210448E74Fc6393")},
				},
				{
					Address: types.MustAddressFromHex("0x447Ddd4960d9fdBF6af9a790560d0AF76795CB08"), // RETH/WSTETH
					Args:    []any{1},
					Return:  []any{types.MustAddressFromHex("0x7f39C581F595B53c5cb19bD0b3f8dA6c935E2Ca0")},
				},
			}).
			WithFunctionData("symbols", []origin.FunctionData{
				{
					Address: types.MustAddressFromHex("0xae78736Cd615f374D3085123A210448E74Fc6393"),
					Args:    []any{},
					Return:  []any{"rETH"},
				},
				{
					Address: types.MustAddressFromHex("0x7f39C581F595B53c5cb19bD0b3f8dA6c935E2Ca0"),
					Args:    []any{},
					Return:  []any{"wstETH"},
				},
			}).
			WithFunctionData("decimals", []origin.FunctionData{
				{
					Address: types.MustAddressFromHex("0xae78736Cd615f374D3085123A210448E74Fc6393"),
					Args:    []any{},
					Return:  []any{big.NewInt(18)},
				},
				{
					Address: types.MustAddressFromHex("0x7f39C581F595B53c5cb19bD0b3f8dA6c935E2Ca0"),
					Args:    []any{},
					Return:  []any{big.NewInt(18)},
				},
			}).
			WithFunctionData("get_dy1", []origin.FunctionData{
				{
					Address: types.MustAddressFromHex("0x447Ddd4960d9fdBF6af9a790560d0AF76795CB08"), // RETH/WSTETH
					Args:    []any{0, 1, big.NewInt(1e18)},
					Return:  []any{big.NewInt(int64(getDy_RETH2WSTETH * 1e18))},
				},
			}).
			WithCustom("blockNumber", blockNumber)).
		Deploy(*s)
	require.NoError(t, err)

	mocks := []*smocker.Mock{
		smocker.NewMockBuilder().
			AddResponseHeader("Content-Type", "application/json").
			SetRequestBodyString(smocker.ShouldContainSubstring("eth_chainId")).
			SetResponseBody(mustReadFile("./testdata/mock/eth_chainId.json")).
			Mock(),
	}
	require.NoError(t, s.AddMocks(ctx, mocks))

	out, err := execCommand(ctx, "..", nil, "./gofer", "-c", "./e2e/testdata/config/gofer.hcl", "-v", "debug", "price", "RETH/WSTETH", "--format", "json")
	require.NoError(t, err)

	priceMap, err := parseGoferPrice(out)
	require.NoError(t, err)
	p := priceMap["RETH/WSTETH"]

	assert.Equal(t, "reference", p.Meta["type"])
	assert.Equal(t, getDy_RETH2WSTETH, p.Value.Price)
	assert.Greater(t, len(p.SubPoints), 0)
}

func Test_Gofer_RETH(t *testing.T) {
	ctx, ctxCancel := context.WithTimeout(context.Background(), 2*time.Minute)
	defer ctxCancel()

	s := smocker.NewAPI("http://127.0.0.1:8081")
	require.NoError(t, s.Reset(ctx))

	const blockNumber int = 100
	getLatest_RETH2ETH := 0.98
	getPriceRateCache_RETH2ETH := 0.4
	getDy_RETH2WSTETH := 3
	getExchangeRate_RETH2ETH := 0.95
	getLatest_WSTETH2WETH := 0.94
	getPriceRateCache_WSTETH2WETH := 0.2
	getDy_STETH2ETH := 2
	stEthPerToken_WSTETH2STETH := 0.9
	sqrtPriceX96_WSTETH2ETH, _ := new(big.Int).SetString("84554395222218770838379633172", 10)

	wsteth_WSTETH2STETH := stEthPerToken_WSTETH2STETH
	balancer_WSTETH2WETH := getLatest_WSTETH2WETH * getPriceRateCache_WSTETH2WETH
	uniswapV3_WSTETH2WETH := 1.1389724195627926
	curve_STETH2ETH := 1 / float64(getDy_STETH2ETH)
	STETH2ETH := float64(balancer_WSTETH2WETH+curve_STETH2ETH) / 2.0
	WSTETH2ETH := (uniswapV3_WSTETH2WETH + wsteth_WSTETH2STETH*STETH2ETH) / 2.0

	balancer_RETH2WETH := getLatest_RETH2ETH * getPriceRateCache_RETH2ETH
	curve_RETH2WSTETH := float64(getDy_RETH2WSTETH)
	rocketpool_RETH2ETH := getExchangeRate_RETH2ETH
	RETH2ETHs := []float64{
		balancer_RETH2WETH,
		curve_RETH2WSTETH * WSTETH2ETH,
		rocketpool_RETH2ETH,
	}
	sort.Slice(RETH2ETHs, func(i, j int) bool {
		return RETH2ETHs[i] < RETH2ETHs[j]
	})
	RETH2ETH := RETH2ETHs[1]

	//fmt.Println("balancer, WSTETH/WETH", balancer_WSTETH2WETH)
	//fmt.Println("curve, STETH/ETH", curve_STETH2ETH)
	//fmt.Println(">> STETH/ETH", STETH2ETH)
	//fmt.Println("uniswapV3, WSTETH/WETH", uniswapV3_WSTETH2WETH)
	//fmt.Println("wsteth, WSTETH/STETH", wsteth_WSTETH2STETH)
	//fmt.Println(">> WSTETH2ETH", WSTETH2ETH)
	//fmt.Println("balancer, RETH/WETH", balancer_RETH2WETH)
	//fmt.Println("curve, RETH/WSTETH", curve_RETH2WSTETH)
	//fmt.Println("rocketpool, RETH/ETH", rocketpool_RETH2ETH)
	//fmt.Println(">> RETH2ETH", RETH2ETH)

	//data_model "RETH/ETH" {
	//	median {
	//		min_values = 3
	//		alias "RETH/ETH" {
	//		origin "balancerV2" { query = "RETH/WETH" }
	//	}
	//		indirect {
	//		origin "curve" { query = "RETH/WSTETH" }
	//		reference { data_model = "WSTETH/ETH" }
	//	}
	//		origin "rocketpool" { query = "RETH/ETH" }
	//	}
	//}
	//data_model "WSTETH/ETH" {
	//	median {
	//		min_values = 2
	//		alias "WSTETH/ETH" {
	//		origin "uniswapV3" { query = "WSTETH/WETH" }
	//	}
	//		indirect {
	//		origin "wsteth" { query = "WSTETH/STETH" }
	//		reference { data_model = "STETH/ETH" }
	//	}
	//	}
	//}
	//data_model "STETH/ETH" {
	//	median {
	//		min_values = 2
	//		alias "STETH/ETH" {
	//		origin "balancerV2" { query = "WSTETH/WETH" }
	//	}
	//		origin "curve" { query = "STETH/ETH" }
	//	}
	//}
	err := infestor.NewMocksBuilder().
		//origin "balancerV2" { query = "RETH/WETH" }
		//origin "balancerV2" { query = "WSTETH/WETH" }
		Add(origin.NewExchange("balancerV2").
			WithSymbol("RETH/WETH").
			WithFunctionData("getLatest", []origin.FunctionData{
				{
					Address: types.MustAddressFromHex("0x1E19CF2D73a72Ef1332C882F20534B6519Be0276"), // RETH/WETH
					Args:    []any{byte(0)},
					Return:  []any{big.NewInt(int64(getLatest_RETH2ETH * 1e18))},
				},
				{
					Address: types.MustAddressFromHex("0x32296969ef14eb0c6d29669c550d4a0449130230"), // WSTETH/WETH
					Args:    []any{byte(0)},
					Return:  []any{big.NewInt(int64(getLatest_WSTETH2WETH * 1e18))},
				},
			}).
			WithFunctionData("getPriceRateCache", []origin.FunctionData{
				{
					Address: types.MustAddressFromHex("0x1E19CF2D73a72Ef1332C882F20534B6519Be0276"), // RETH/WETH
					Args:    []any{types.MustAddressFromHex("0xae78736Cd615f374D3085123A210448E74Fc6393")},
					Return:  []any{big.NewInt(int64(getPriceRateCache_RETH2ETH * 1e18)), big.NewInt(10), big.NewInt(time.Now().Unix())},
				},
				{
					Address: types.MustAddressFromHex("0x32296969ef14eb0c6d29669c550d4a0449130230"), // WSTETH/WETH
					Args:    []any{types.MustAddressFromHex("0x7f39C581F595B53c5cb19bD0b3f8dA6c935E2Ca0")},
					Return:  []any{big.NewInt(int64(getPriceRateCache_WSTETH2WETH * 1e18)), big.NewInt(10), big.NewInt(time.Now().Unix())},
				},
			}).
			WithCustom("blockNumber", blockNumber)).
		//origin "curve" { query = "RETH/WSTETH" }
		//origin "curve" { query = "STETH/ETH" }
		Add(origin.NewExchange("curve").
			WithSymbol("RETH/WSTETH").
			WithFunctionData("coins", []origin.FunctionData{
				{
					Address: types.MustAddressFromHex("0x447Ddd4960d9fdBF6af9a790560d0AF76795CB08"), // RETH/WSTETH
					Args:    []any{0},
					Return:  []any{types.MustAddressFromHex("0xae78736Cd615f374D3085123A210448E74Fc6393")},
				},
				{
					Address: types.MustAddressFromHex("0x447Ddd4960d9fdBF6af9a790560d0AF76795CB08"), // RETH/WSTETH
					Args:    []any{1},
					Return:  []any{types.MustAddressFromHex("0x7f39C581F595B53c5cb19bD0b3f8dA6c935E2Ca0")},
				},
				{
					Address: types.MustAddressFromHex("0xDC24316b9AE028F1497c275EB9192a3Ea0f67022"), // ETH/STETH
					Args:    []any{1},
					Return:  []any{types.MustAddressFromHex("0xae7ab96520de3a18e5e111b5eaab095312d7fe84")},
				},
				{
					Address: types.MustAddressFromHex("0xDC24316b9AE028F1497c275EB9192a3Ea0f67022"), // ETH/STETH
					Args:    []any{0},
					Return:  []any{types.MustAddressFromHex("0xeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeee")},
				},
			}).
			WithFunctionData("symbols", []origin.FunctionData{
				{
					Address: types.MustAddressFromHex("0xae78736Cd615f374D3085123A210448E74Fc6393"),
					Args:    []any{},
					Return:  []any{"rETH"},
				},
				{
					Address: types.MustAddressFromHex("0x7f39C581F595B53c5cb19bD0b3f8dA6c935E2Ca0"),
					Args:    []any{},
					Return:  []any{"wstETH"},
				},
				{
					Address: types.MustAddressFromHex("0xae7ab96520de3a18e5e111b5eaab095312d7fe84"),
					Args:    []any{},
					Return:  []any{"stETH"},
				},
				//{ // do not call for ETH
				//	Address: types.MustAddressFromHex("0xeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeee"),
				//	Args:   []any{},
				//	Return: []any{"ETH"},
				//},
			}).
			WithFunctionData("decimals", []origin.FunctionData{
				{
					Address: types.MustAddressFromHex("0xae78736Cd615f374D3085123A210448E74Fc6393"),
					Args:    []any{},
					Return:  []any{big.NewInt(18)},
				},
				{
					Address: types.MustAddressFromHex("0x7f39C581F595B53c5cb19bD0b3f8dA6c935E2Ca0"),
					Args:    []any{},
					Return:  []any{big.NewInt(18)},
				},
				{
					Address: types.MustAddressFromHex("0xae7ab96520de3a18e5e111b5eaab095312d7fe84"),
					Args:    []any{},
					Return:  []any{big.NewInt(18)},
				},
				// { // do not call for ETH
				//	Address: types.MustAddressFromHex("0xeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeee"),
				//	Args:   []any{},
				//	Return: []any{big.NewInt(18)},
				// },
			}).
			WithFunctionData("get_dy1", []origin.FunctionData{
				{
					Address: types.MustAddressFromHex("0x447Ddd4960d9fdBF6af9a790560d0AF76795CB08"), // RETH/WSTETH
					Args:    []any{0, 1, big.NewInt(1e18)},
					Return:  []any{big.NewInt(int64(getDy_RETH2WSTETH * 1e18))},
				},
				{
					Address: types.MustAddressFromHex("0xDC24316b9AE028F1497c275EB9192a3Ea0f67022"), // ETH/STETH
					Args:    []any{0, 1, big.NewInt(1e18)},
					Return:  []any{big.NewInt(int64(getDy_STETH2ETH * 1e18))},
				},
			}).
			WithCustom("blockNumber", blockNumber)).
		//origin "rocketpool" { query = "RETH/ETH" }
		Add(origin.NewExchange("rocketpool").
			WithSymbol("RETH/ETH").
			WithFunctionData("getExchangeRate", []origin.FunctionData{
				{
					Address: types.MustAddressFromHex("0xae78736Cd615f374D3085123A210448E74Fc6393"), // RETH/ETH
					Args:    []any{},
					Return:  []any{big.NewInt(int64(getExchangeRate_RETH2ETH * 1e18))},
				},
			}).
			WithCustom("blockNumber", blockNumber)).
		//origin "wsteth" { query = "WSTETH/STETH" }
		Add(origin.NewExchange("wsteth").
			WithSymbol("WSTETH/STETH").
			WithFunctionData("stEthPerToken", []origin.FunctionData{
				{
					Address: types.MustAddressFromHex("0x7f39C581F595B53c5cb19bD0b3f8dA6c935E2Ca0"), // WSTETH/STETH
					Args:    []any{},
					Return:  []any{big.NewInt(int64(stEthPerToken_WSTETH2STETH * 1e18))},
				},
			}).
			WithCustom("blockNumber", blockNumber)).
		//origin "uniswapV3" { query = "WSTETH/WETH" }
		Add(origin.NewExchange("uniswapV3").
			WithSymbol("WSTETH/WETH").
			WithFunctionData("slot0", []origin.FunctionData{
				{
					Address: types.MustAddressFromHex("0x109830a1AAaD605BbF02a9dFA7B0B92EC2FB7dAa"), // WSTETH/WETH
					Args:    []any{},
					Return: []any{
						sqrtPriceX96_WSTETH2ETH, // sqrtPriceX96
						big.NewInt(1301),        // tick
						big.NewInt(23),          // observationIndex
						big.NewInt(150),         // observationCardinality
						big.NewInt(150),         // observationCardinalityNext
						0,                       // feeProtocol
						false,                   // unlocked
					},
				},
			}).
			WithFunctionData("token0", []origin.FunctionData{
				{
					Address: types.MustAddressFromHex("0x109830a1AAaD605BbF02a9dFA7B0B92EC2FB7dAa"), // WSTETH/WETH
					Args:    []any{},
					Return:  []any{types.MustAddressFromHex("0x7f39C581F595B53c5cb19bD0b3f8dA6c935E2Ca0")},
				},
			}).
			WithFunctionData("token1", []origin.FunctionData{
				{
					Address: types.MustAddressFromHex("0x109830a1AAaD605BbF02a9dFA7B0B92EC2FB7dAa"), // WSTETH/WETH
					Args:    []any{},
					Return:  []any{types.MustAddressFromHex("0xC02aaA39b223FE8D0A0e5C4F27eAD9083C756Cc2")},
				},
			}).
			WithFunctionData("symbols", []origin.FunctionData{
				{
					Address: types.MustAddressFromHex("0x7f39C581F595B53c5cb19bD0b3f8dA6c935E2Ca0"),
					Args:    []any{},
					Return:  []any{"wstETH"},
				},
				{
					Address: types.MustAddressFromHex("0xC02aaA39b223FE8D0A0e5C4F27eAD9083C756Cc2"),
					Args:    []any{},
					Return:  []any{"WETH"},
				},
			}).
			WithFunctionData("decimals", []origin.FunctionData{
				{
					Address: types.MustAddressFromHex("0x7f39C581F595B53c5cb19bD0b3f8dA6c935E2Ca0"),
					Args:    []any{},
					Return:  []any{big.NewInt(18)},
				},
				{
					Address: types.MustAddressFromHex("0xC02aaA39b223FE8D0A0e5C4F27eAD9083C756Cc2"),
					Args:    []any{},
					Return:  []any{big.NewInt(18)},
				},
			}).
			WithCustom("blockNumber", blockNumber)).
		Deploy(*s)
	require.NoError(t, err)

	mocks := []*smocker.Mock{
		smocker.NewMockBuilder().
			AddResponseHeader("Content-Type", "application/json").
			SetRequestBodyString(smocker.ShouldContainSubstring("eth_chainId")).
			SetResponseBody(mustReadFile("./testdata/mock/eth_chainId.json")).
			Mock(),
	}
	require.NoError(t, s.AddMocks(ctx, mocks))

	out, err := execCommand(ctx, "..", nil, "./gofer", "-c", "./e2e/testdata/config/gofer.hcl", "-v", "debug", "price", "RETH/ETH", "--format", "json")
	require.NoError(t, err)

	priceMap, err := parseGoferPrice(out)
	require.NoError(t, err)
	p := priceMap["RETH/ETH"]

	assert.Equal(t, "reference", p.Meta["type"])
	assert.Equal(t, RETH2ETH, p.Value.Price)
	assert.Greater(t, len(p.SubPoints), 0)
}
