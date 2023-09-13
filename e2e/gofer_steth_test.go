package e2e

import (
	"context"
	"math/big"
	"testing"
	"time"

	"github.com/defiweb/go-eth/types"

	"github.com/chronicleprotocol/infestor"
	"github.com/chronicleprotocol/infestor/origin"
	"github.com/chronicleprotocol/infestor/smocker"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func Test_Gofer_STETH(t *testing.T) {
	ctx, ctxCancel := context.WithTimeout(context.Background(), 2*time.Minute)
	defer ctxCancel()

	s := smocker.NewAPI("http://127.0.0.1:8081")
	require.NoError(t, s.Reset(ctx))

	const blockNumber int = 100
	getLatest := 0.94
	getPriceRateCache := 0.2
	getDy := 2
	STETH2ETH := float64(getLatest*getPriceRateCache+1/float64(getDy)) / 2.0
	err := infestor.NewMocksBuilder().
		Add(origin.NewExchange("balancerV2").
			WithSymbol("WSTETH/WETH").
			WithFunctionData("getLatest", []origin.FunctionData{
				{
					Address: types.MustAddressFromHex("0x32296969ef14eb0c6d29669c550d4a0449130230"), // WSTETH/WETH
					Args:    []any{byte(0)},
					Return:  []any{big.NewInt(int64(getLatest * 1e18))},
				},
			}).
			WithFunctionData("getPriceRateCache", []origin.FunctionData{
				{
					Address: types.MustAddressFromHex("0x32296969ef14eb0c6d29669c550d4a0449130230"), // WSTETH/WETH
					Args:    []any{types.MustAddressFromHex("0x7f39C581F595B53c5cb19bD0b3f8dA6c935E2Ca0")},
					Return:  []any{big.NewInt(int64(getPriceRateCache * 1e18)), big.NewInt(10), big.NewInt(time.Now().Unix())},
				},
			}).
			WithCustom("blockNumber", blockNumber)).
		Add(origin.NewExchange("curve").
			WithSymbol("ETH/STETH").
			WithCustom("ETH/STETH", types.MustAddressFromHex("0xDC24316b9AE028F1497c275EB9192a3Ea0f67022")).
			WithFunctionData("coins", []origin.FunctionData{
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
			WithCustom("tokens", []types.Address{
				types.MustAddressFromHex("0xae7ab96520de3a18e5e111b5eaab095312d7fe84"),
				// types.MustAddressFromHex("0xeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeee"), // do not call for ETH
			}).
			WithFunctionData("symbols", []origin.FunctionData{
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
					Address: types.MustAddressFromHex("0xDC24316b9AE028F1497c275EB9192a3Ea0f67022"), // ETH/STETH
					Args:    []any{0, 1, big.NewInt(1e18)},
					Return:  []any{big.NewInt(int64(getDy * 1e18))},
				},
			}).
			WithCustom("blockNumber", blockNumber)).
		Deploy(*s)
	require.NoError(t, err)

	out, err := execCommand(ctx, "..", nil, "./gofer", "-c", "./e2e/testdata/config/gofer.hcl", "-v", "debug", "price", "STETH/ETH", "--format", "json")
	require.NoError(t, err)

	priceMap, err := parseGoferPrice(out)
	require.NoError(t, err)
	p := priceMap["STETH/ETH"]

	assert.Equal(t, "reference", p.Meta["type"])
	assert.Equal(t, STETH2ETH, p.Value.Price)
	assert.Greater(t, len(p.SubPoints), 0)
}
