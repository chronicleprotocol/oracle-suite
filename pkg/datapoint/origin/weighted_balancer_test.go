package origin

import (
	"fmt"
	"testing"

	"github.com/defiweb/go-eth/types"
	"github.com/stretchr/testify/assert"

	"github.com/chronicleprotocol/oracle-suite/pkg/datapoint/value"
	"github.com/chronicleprotocol/oracle-suite/pkg/util/bn"
)

func TestWeightedPool_Swap(t *testing.T) {
	testCases := []struct {
		pool      *WeightedPool
		tokenIn   ERC20Details
		amountIn  *bn.DecFixedPointNumber
		tokenOut  ERC20Details
		amountOut *bn.DecFixedPointNumber
	}{
		{
			// txhash: 0x74dac9957a9b4f3892ebbcf6deb7ca4d98ed5e0b0769c28ae1c81f5819125955
			pool: &WeightedPool{
				pair: value.Pair{
					Base:  "RDNT",
					Quote: "WETH",
				},
				address: types.MustAddressFromHex("0xcF7b51ce5755513d4bE016b0e28D6EDEffa1d52a"),
				tokens: []types.Address{
					types.MustAddressFromHex("0x137dDB47Ee24EaA998a535Ab00378d6BFa84F893"), // RDNT
					types.MustAddressFromHex("0xC02aaA39b223FE8D0A0e5C4F27eAD9083C756Cc2"), // WETH
				},
				balances: []*bn.DecFixedPointNumber{
					string2DecFixedPointNumber("34043497190382699990148821"), // RDNT
					string2DecFixedPointNumber("1060514722983166251296"),     // WETH
				},
				swapFeePercentage: string2DecFixedPointNumber("5000000000000000"),
				scalingFactors: []*bn.DecFixedPointNumber{
					string2DecFixedPointNumber("1000000000000000000"),
					string2DecFixedPointNumber("1000000000000000000"),
				},
				normalizedWeights: []*bn.DecFixedPointNumber{
					string2DecFixedPointNumber("800000000000000000"),
					string2DecFixedPointNumber("200000000000000000"),
				},
			},
			tokenIn: ERC20Details{
				address:  types.MustAddressFromHex("0x137dDB47Ee24EaA998a535Ab00378d6BFa84F893"),
				symbol:   "RDNT",
				decimals: 18,
			},
			amountIn: string2DecFixedPointNumber("40000000000000000000000"),
			tokenOut: ERC20Details{
				address:  types.MustAddressFromHex("0xC02aaA39b223FE8D0A0e5C4F27eAD9083C756Cc2"),
				symbol:   "WETH",
				decimals: 18,
			},
			amountOut: string2DecFixedPointNumber("4944898525417925727"),
		},
		{
			pool: &WeightedPool{
				pair: value.Pair{
					Base:  "WUSDM",
					Quote: "WSTETH",
				},
				address: types.MustAddressFromHex("0x54ca50EE86616379420Cc56718E12566aa75Abbe"),
				tokens: []types.Address{
					types.MustAddressFromHex("0x57F5E098CaD7A3D1Eed53991D4d66C45C9AF7812"), // WUSDM
					types.MustAddressFromHex("0x7f39C581F595B53c5cb19bD0b3f8dA6c935E2Ca0"), // WSTETH
				},
				balances: []*bn.DecFixedPointNumber{
					string2DecFixedPointNumber("60655883048463530117866"), // WUSDM
					string2DecFixedPointNumber("25630194454768640289"),    // WSTETH
				},
				swapFeePercentage: string2DecFixedPointNumber("3000000000000000"),
				scalingFactors: []*bn.DecFixedPointNumber{
					string2DecFixedPointNumber("1000000000000000000"),
					string2DecFixedPointNumber("1000000000000000000"),
				},
				normalizedWeights: []*bn.DecFixedPointNumber{
					string2DecFixedPointNumber("500000000000000000"),
					string2DecFixedPointNumber("500000000000000000"),
				},
			},
			tokenIn: ERC20Details{
				address:  types.MustAddressFromHex("0x57F5E098CaD7A3D1Eed53991D4d66C45C9AF7812"),
				symbol:   "WUSDM",
				decimals: 18,
			},
			amountIn: string2DecFixedPointNumber("1000000000000000000000"), // 1000 WUSDM
			tokenOut: ERC20Details{
				address:  types.MustAddressFromHex("0x7f39C581F595B53c5cb19bD0b3f8dA6c935E2Ca0"),
				symbol:   "WSTETH",
				decimals: 18,
			},
			amountOut: string2DecFixedPointNumber("414470542299175666"),
		},
	}

	for i, testcase := range testCases {
		t.Run(fmt.Sprintf("testcase %d, tokenIn %s amountIn %s tokenOut %s amountOut %s", i, testcase.tokenIn.symbol, testcase.amountIn.String(), testcase.tokenOut.symbol, testcase.amountOut.String()), func(t *testing.T) {
			amountOut, _, err := testcase.pool.CalcAmountOut(testcase.tokenIn.address, testcase.tokenOut.address, testcase.amountIn)
			assert.NoError(t, err)
			assert.Equal(t, testcase.amountOut, amountOut)
		})
	}
}
