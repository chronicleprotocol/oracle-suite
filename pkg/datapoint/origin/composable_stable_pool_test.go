package origin

import (
	"fmt"
	"github.com/stretchr/testify/require"
	"math/big"
	"testing"

	"github.com/defiweb/go-eth/types"
	"github.com/stretchr/testify/assert"

	"github.com/chronicleprotocol/oracle-suite/pkg/datapoint/value"
	"github.com/chronicleprotocol/oracle-suite/pkg/util/bn"
)

func string2DecFloatPointNumber(s string) *bn.DecFloatPointNumber {
	b, _ := new(big.Int).SetString(s, 10)
	return bn.DecFloatPoint(b)
}

func TestComposableStablePool_Swap(t *testing.T) {
	// https://etherscan.io/tx/0xd6b6c4b43551b658dad1032c832f947c8e2cbb6ee61a69dab0558902579b0548
	pool := ComposableStablePool{
		pair: value.Pair{
			Base:  "GHO",
			Quote: "USDC",
		},
		address: types.MustAddressFromHex("0x8353157092ED8Be69a9DF8F95af097bbF33Cb2aF"),
		tokens: []types.Address{
			types.MustAddressFromHex("0x40D16FC0246aD3160Ccc09B8D0D3A2cD28aE6C2f"), // GHO
			types.MustAddressFromHex("0x8353157092ED8Be69a9DF8F95af097bbF33Cb2aF"), // GHO/USDT/USDC
			types.MustAddressFromHex("0xA0b86991c6218b36c1d19D4a2e9Eb0cE3606eB48"), // USDC
			types.MustAddressFromHex("0xdAC17F958D2ee523a2206206994597C13D831ec7"), // USDT
		},
		balances: []*bn.DecFloatPointNumber{
			string2DecFloatPointNumber("6448444062456011477376368"),
			string2DecFloatPointNumber("2596148429302257816743021881556180"),
			string2DecFloatPointNumber("1513827018794"),
			string2DecFloatPointNumber("1538170251459"),
		},
		bptIndex: 1,
		rateProviders: []types.Address{
			types.MustAddressFromHex("0x0000000000000000000000000000000000000000"),
			types.MustAddressFromHex("0x0000000000000000000000000000000000000000"),
			types.MustAddressFromHex("0x0000000000000000000000000000000000000000"),
			types.MustAddressFromHex("0x0000000000000000000000000000000000000000"),
		},
		totalSupply:       string2DecFloatPointNumber("2596148438770953798709961309149655"),
		swapFeePercentage: string2DecFloatPointNumber("500000000000000"),
		extra: Extra{
			amplificationParameter: AmplificationParameter{
				value:      bn.DecFloatPoint(200000),
				isUpdating: false,
				precision:  bn.DecFloatPoint(1000),
			},
			scalingFactors: []*bn.DecFloatPointNumber{
				string2DecFloatPointNumber("1000000000000000000"),
				string2DecFloatPointNumber("1000000000000000000"),
				string2DecFloatPointNumber("1000000000000000000000000000000"),
				string2DecFloatPointNumber("1000000000000000000000000000000"),
			},
			lastJoinExit: LastJoinExitData{
				lastJoinExitAmplification: string2DecFloatPointNumber("200000"),
				lastPostJoinExitInvariant: string2DecFloatPointNumber("9482927260967981674261420"),
			},
			tokensExemptFromYieldProtocolFee: []bool{
				false, false, false, false,
			},
			tokenRateCaches: []TokenRateCache{
				{rate: nil, oldRate: nil, duration: nil, expires: nil},
				{rate: nil, oldRate: nil, duration: nil, expires: nil},
				{rate: nil, oldRate: nil, duration: nil, expires: nil},
				{rate: nil, oldRate: nil, duration: nil, expires: nil},
			},
			protocolFeePercentageCacheSwapType:  string2DecFloatPointNumber("500000000000000000"),
			protocolFeePercentageCacheYieldType: string2DecFloatPointNumber("500000000000000000"),
		},
	}

	testCases := []struct {
		tokenIn   types.Address
		amountIn  *bn.DecFloatPointNumber
		tokenOut  types.Address
		amountOut *bn.DecFloatPointNumber
	}{
		{
			tokenIn:   types.MustAddressFromHex("0x40D16FC0246aD3160Ccc09B8D0D3A2cD28aE6C2f"), // GHO
			amountIn:  string2DecFloatPointNumber("10551510000000000000000"),
			tokenOut:  types.MustAddressFromHex("0xA0b86991c6218b36c1d19D4a2e9Eb0cE3606eB48"), // USDC
			amountOut: string2DecFloatPointNumber("10371548845"),
		},
		{
			tokenIn:   types.MustAddressFromHex("0x40D16FC0246aD3160Ccc09B8D0D3A2cD28aE6C2f"), // GHO
			amountIn:  string2DecFloatPointNumber("1000000000000000000"),
			tokenOut:  types.MustAddressFromHex("0xA0b86991c6218b36c1d19D4a2e9Eb0cE3606eB48"), // USDC
			amountOut: string2DecFloatPointNumber("983063"),
		},
		//{
		//	tokenIn:   types.MustAddressFromHex("0x8353157092ED8Be69a9DF8F95af097bbF33Cb2aF"), // GHO/USDT/USDC
		//	amountIn:  string2DecFloatPointNumber("1000000000000000000"),
		//	tokenOut:  types.MustAddressFromHex("0xA0b86991c6218b36c1d19D4a2e9Eb0cE3606eB48"), // USDC
		//	amountOut: string2DecFloatPointNumber("991677"),
		//},
		//{
		//	tokenIn:   types.MustAddressFromHex("0xA0b86991c6218b36c1d19D4a2e9Eb0cE3606eB48"), // USDC
		//	amountIn:  string2DecFloatPointNumber("1000000000000000000"),
		//	tokenOut:  types.MustAddressFromHex("0x8353157092ED8Be69a9DF8F95af097bbF33Cb2aF"), // GHO/USDT/USDC
		//	amountOut: string2DecFloatPointNumber("19877475578824849899330863774"),
		//},
		{
			tokenIn:   types.MustAddressFromHex("0xdAC17F958D2ee523a2206206994597C13D831ec7"), // USDT
			amountIn:  string2DecFloatPointNumber("1000000000000000000"),
			tokenOut:  types.MustAddressFromHex("0xA0b86991c6218b36c1d19D4a2e9Eb0cE3606eB48"), // USDC
			amountOut: string2DecFloatPointNumber("1513827018793"),
		},
	}

	for i, testcase := range testCases {
		t.Run(fmt.Sprintf("testcase %d, tokenIn %s amountIn %s tokenOut %s amountOut %s", i, testcase.tokenIn, testcase.amountIn.String(), testcase.tokenOut, testcase.amountOut.String()), func(t *testing.T) {
			amountOut, _, _ := pool.CalcAmountOut(testcase.tokenIn, testcase.tokenOut, testcase.amountIn)
			assert.Equal(t, testcase.amountOut, amountOut)
		})
	}
}

func TestCalculateInvariant(t *testing.T) {
	testCases := []struct {
		name      string
		amp       *bn.DecFloatPointNumber
		balances  []*bn.DecFloatPointNumber
		invariant *bn.DecFloatPointNumber
		error     error
	}{
		{
			name: "success",
			amp:  bn.DecFloatPoint(60000),
			balances: []*bn.DecFloatPointNumber{
				string2DecFloatPointNumber("50310513788381313281"),
				string2DecFloatPointNumber("19360701460293571158"),
				string2DecFloatPointNumber("58687814461000000000000"),
			},
			invariant: string2DecFloatPointNumber("10749877394384654056023"),
		},
		{
			name: "revert",
			amp:  bn.DecFloatPoint(60000),
			balances: []*bn.DecFloatPointNumber{
				string2DecFloatPointNumber("50310513788381313281"),
				string2DecFloatPointNumber("19360701460293571158"),
				string2DecFloatPointNumber("10000"),
			},
			error: STABLE_INVARIANT_DIDNT_CONVERGE,
		},
	}

	for _, testcase := range testCases {
		t.Run(testcase.name, func(t *testing.T) {
			invariant, err := _calculateInvariant(testcase.amp, testcase.balances)
			if testcase.error != nil {
				assert.Equal(t, err, testcase.error)
			} else {
				require.NoError(t, err)
				fmt.Println(invariant.String())
				fmt.Println(testcase.invariant.String())
				assert.Equal(t, invariant, testcase.invariant)
			}
		})
	}
}
