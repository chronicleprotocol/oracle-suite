package origin

import (
	"fmt"
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
	var BoneFloat = bn.DecFloatPoint("1000000000000000000")
	swapFee := bn.DecFloatPoint(big.NewFloat(0.000001)).Mul(BoneFloat)
	config := ComposableStablePoolFullConfig{
		Pair: value.Pair{
			Base:  "A",
			Quote: "B",
		},
		ContractAddress: types.MustAddressFromHex("0x9001cbbd96f54a658ff4e6e65ab564ded76a5431"),
		PoolID:          types.Bytes("0x9001cbbd96f54a658ff4e6e65ab564ded76a543100000000000000000000050a"),
		Vault:           types.MustAddressFromHex("0xba12222222228d8ba445958a75a0704d566bf2c8"),
		Tokens: []types.Address{
			types.MustAddressFromHex("0x60d604890feaa0b5460b28a424407c24fe89374a"), // A
			types.MustAddressFromHex("0x9001cbbd96f54a658ff4e6e65ab564ded76a5431"), // B
			types.MustAddressFromHex("0xbe9895146f7af43049ca1c1ae358b0541ea49704"), // C
		},
		BptIndex: 1,
		RateProviders: []types.Address{
			types.MustAddressFromHex("0x60d604890feaa0b5460b28a424407c24fe89374a"),
			types.MustAddressFromHex("0x0000000000000000000000000000000000000000"),
			types.MustAddressFromHex("0x7311e4bb8a72e7b300c5b8bde4de6cdaa822a5b1"),
		},
		Balances: []*bn.DecFloatPointNumber{
			bn.DecFloatPoint("2518960237189623226641"),
			bn.DecFloatPoint("2596148429266323438822175768385755"),
			bn.DecFloatPoint("3457262534881651304610"),
		},
		TotalSupply:       bn.DecFloatPoint("2596148429272429220684965023562161"),
		SwapFeePercentage: swapFee,
		Extra: Extra{
			AmplificationParameter: AmplificationParameter{
				Value:      bn.DecFloatPoint(700000),
				IsUpdating: false,
				Precision:  bn.DecFloatPoint(1000),
			},
			ScalingFactors: []*bn.DecFloatPointNumber{
				bn.DecFloatPoint("1003649423771917631"),
				bn.DecFloatPoint("1000000000000000000"),
				bn.DecFloatPoint("1043680240732074966"),
			},
			LastJoinExit: LastJoinExitData{
				LastJoinExitAmplification: bn.DecFloatPoint("700000"),
				LastPostJoinExitInvariant: bn.DecFloatPoint("6135006746648647084879"),
			},
			TokensExemptFromYieldProtocolFee: []bool{
				false, false, false,
			},
			TokenRateCaches: []TokenRateCache{
				{
					Rate:     bn.DecFloatPoint("1003649423771917631"),
					OldRate:  bn.DecFloatPoint("1003554274984131981"),
					Duration: bn.DecFloatPoint("21600"),
					Expires:  bn.DecFloatPoint("1689845039"),
				},
				{
					Rate:     nil,
					OldRate:  nil,
					Duration: nil,
					Expires:  nil,
				},
				{
					Rate:     bn.DecFloatPoint("1043680240732074966"),
					OldRate:  bn.DecFloatPoint("1043375386816533719"),
					Duration: bn.DecFloatPoint("21600"),
					Expires:  bn.DecFloatPoint("1689845039"),
				},
			},
			ProtocolFeePercentageCacheSwapType:  bn.DecFloatPoint(0),
			ProtocolFeePercentageCacheYieldType: bn.DecFloatPoint(0),
		},
	}

	p, _ := NewComposableStablePoolFull(config)

	testCases := []struct {
		tokenIn   ERC20Details
		amountIn  *bn.DecFloatPointNumber
		tokenOut  ERC20Details
		amountOut *bn.DecFloatPointNumber
	}{
		{
			tokenIn: ERC20Details{
				address:  types.MustAddressFromHex("0x60d604890feaa0b5460b28a424407c24fe89374a"),
				symbol:   "A",
				decimals: 18,
			},
			amountIn: bn.DecFloatPoint("12000000000000000000"),
			tokenOut: ERC20Details{
				address:  types.MustAddressFromHex("0xbe9895146f7af43049ca1c1ae358b0541ea49704"),
				symbol:   "C",
				decimals: 18,
			},
			amountOut: bn.DecFloatPoint("11545818036500155269"),
		},
		{
			tokenIn: ERC20Details{
				address:  types.MustAddressFromHex("0x60d604890feaa0b5460b28a424407c24fe89374a"),
				symbol:   "A",
				decimals: 18,
			},
			amountIn: bn.DecFloatPoint("1000000000000000000"),
			tokenOut: ERC20Details{
				address:  types.MustAddressFromHex("0xbe9895146f7af43049ca1c1ae358b0541ea49704"),
				symbol:   "C",
				decimals: 18,
			},
			amountOut: bn.DecFloatPoint("962157416748443460"),
		},
		{
			tokenIn: ERC20Details{
				address:  types.MustAddressFromHex("0x9001cbbd96f54a658ff4e6e65ab564ded76a5431"),
				symbol:   "B",
				decimals: 18,
			},
			amountIn: bn.DecFloatPoint("1000000000000000000"),
			tokenOut: ERC20Details{
				address:  types.MustAddressFromHex("0xbe9895146f7af43049ca1c1ae358b0541ea49704"),
				symbol:   "C",
				decimals: 18,
			},
			amountOut: bn.DecFloatPoint("963168955346971740"),
		},
		{
			tokenIn: ERC20Details{
				address:  types.MustAddressFromHex("0xbe9895146f7af43049ca1c1ae358b0541ea49704"),
				symbol:   "C",
				decimals: 18,
			},
			amountIn: bn.DecFloatPoint("1000000000000000000"),
			tokenOut: ERC20Details{
				address:  types.MustAddressFromHex("0x9001cbbd96f54a658ff4e6e65ab564ded76a5431"),
				symbol:   "B",
				decimals: 18,
			},
			amountOut: bn.DecFloatPoint("1038238386186088886"),
		},
	}

	for i, testcase := range testCases {
		t.Run(fmt.Sprintf("testcase %d, tokenIn %s amountIn %s tokenOut %s amountOut %s", i, testcase.tokenIn.symbol, testcase.amountIn.String(), testcase.tokenOut.symbol, testcase.amountOut.String()), func(t *testing.T) {
			amountOut, _, _ := p.calcAmountOut(testcase.tokenIn, testcase.tokenOut, testcase.amountIn)
			assert.Equal(t, testcase.amountOut, amountOut)
		})
	}
}

func TestCalculateInvariant(t *testing.T) {
	a := bn.DecFloatPoint(60000)
	b1 := string2DecFloatPointNumber("50310513788381313281")
	b2 := string2DecFloatPointNumber("19360701460293571158")
	b3 := string2DecFloatPointNumber("58687814461000000000000")

	fmt.Println(b1.String())
	fmt.Println(b2.String())
	fmt.Println(b3.String())

	balances := []*bn.DecFloatPointNumber{
		b1, b2, b3,
	}
	_, err := _calculateInvariant(a, balances)
	assert.Equal(t, err, fmt.Errorf("STABLE_INVARIANT_DIDNT_CONVERGE"))
}
