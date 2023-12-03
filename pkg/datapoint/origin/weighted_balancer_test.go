package origin

import (
	"fmt"
	"github.com/chronicleprotocol/oracle-suite/pkg/datapoint/value"
	"github.com/chronicleprotocol/oracle-suite/pkg/util/bn"
	"github.com/defiweb/go-eth/types"
	"github.com/stretchr/testify/assert"
	"math/big"
	"testing"
)

func string2DecFloatPointNumber(s string) *bn.DecFloatPointNumber {
	b, _ := new(big.Int).SetString(s, 10)
	return bn.DecFloatPoint(b)
}

func TestWeightedPool_Swap(t *testing.T) {
	testCases := []struct {
		pool      *WeightedPool
		tokenIn   ERC20Details
		amountIn  *bn.DecFloatPointNumber
		tokenOut  ERC20Details
		amountOut *bn.DecFloatPointNumber
	}{
		{
			// txhash: 0x74dac9957a9b4f3892ebbcf6deb7ca4d98ed5e0b0769c28ae1c81f5819125955
			pool: &WeightedPool{
				pair: value.Pair{
					Base:  "RDNT",
					Quote: "WETH",
				},
				address: types.MustAddressFromHex("0x54ca50ee86616379420cc56718e12566aa75abbe"),
				tokens: []types.Address{
					types.MustAddressFromHex("0x137dDB47Ee24EaA998a535Ab00378d6BFa84F893"), // RDNT
					types.MustAddressFromHex("0xC02aaA39b223FE8D0A0e5C4F27eAD9083C756Cc2"), // WETH
				},
				balances: []*bn.DecFloatPointNumber{
					string2DecFloatPointNumber("34043497190382699990148821"), // RDNT
					string2DecFloatPointNumber("1060514722983166251296"),     // WETH
				},
				swapFeePercentage: bn.DecFloatPoint("5000000000000000"),
				scalingFactors: []*bn.DecFloatPointNumber{
					string2DecFloatPointNumber("1000000000000000000"),
					string2DecFloatPointNumber("1000000000000000000"),
				},
				normalizedWeights: []*bn.DecFloatPointNumber{
					string2DecFloatPointNumber("800000000000000000"),
					string2DecFloatPointNumber("200000000000000000"),
				},
			},
			tokenIn: ERC20Details{
				address:  types.MustAddressFromHex("0x137dDB47Ee24EaA998a535Ab00378d6BFa84F893"),
				symbol:   "RDNT",
				decimals: 18,
			},
			amountIn: string2DecFloatPointNumber("40000000000000000000000"),
			tokenOut: ERC20Details{
				address:  types.MustAddressFromHex("0xC02aaA39b223FE8D0A0e5C4F27eAD9083C756Cc2"),
				symbol:   "WETH",
				decimals: 18,
			},
			amountOut: string2DecFloatPointNumber("4944898525417925727"),
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
