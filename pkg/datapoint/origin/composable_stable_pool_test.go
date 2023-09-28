package origin

import (
	"fmt"
	"math/big"
	"testing"

	"github.com/defiweb/go-eth/types"

	"github.com/chronicleprotocol/oracle-suite/pkg/datapoint/value"

	"github.com/stretchr/testify/assert"
)

func string2BigInt(s string) *big.Int {
	b, _ := new(big.Int).SetString(s, 10)
	return b
}

func TestComposableStablePool_Swap(t *testing.T) {
	var BoneFloat, _ = new(big.Float).SetString("1000000000000000000")
	swapFee, _ := new(big.Float).Mul(big.NewFloat(0.000001), BoneFloat).Int(nil)
	config := ComposableStablePoolFullConfig{
		Pair: value.Pair{
			Base:  "A",
			Quote: "B",
		},
		ContractAddress: types.MustAddressFromHex("0x9001cbbd96f54a658ff4e6e65ab564ded76a5431"),
		PoolID:          MustBytes32FromHex("0x9001cbbd96f54a658ff4e6e65ab564ded76a543100000000000000000000050a"),
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
		Balances: []*big.Int{
			string2BigInt("2518960237189623226641"),
			string2BigInt("2596148429266323438822175768385755"),
			string2BigInt("3457262534881651304610"),
		},
		TotalSupply:       string2BigInt("2596148429272429220684965023562161"),
		SwapFeePercentage: swapFee,
		Extra: Extra{
			AmplificationParameter: AmplificationParameter{
				Value:      big.NewInt(700000),
				IsUpdating: false,
				Precision:  big.NewInt(1000),
			},
			ScalingFactors: []*big.Int{
				string2BigInt("1003649423771917631"),
				string2BigInt("1000000000000000000"),
				string2BigInt("1043680240732074966"),
			},
			LastJoinExit: LastJoinExitData{
				LastJoinExitAmplification: string2BigInt("700000"),
				LastPostJoinExitInvariant: string2BigInt("6135006746648647084879"),
			},
			TokensExemptFromYieldProtocolFee: []bool{
				false, false, false,
			},
			TokenRateCaches: []TokenRateCache{
				{
					Rate:     string2BigInt("1003649423771917631"),
					OldRate:  string2BigInt("1003554274984131981"),
					Duration: string2BigInt("21600"),
					Expires:  string2BigInt("1689845039"),
				},
				{
					Rate:     nil,
					OldRate:  nil,
					Duration: nil,
					Expires:  nil,
				},
				{
					Rate:     string2BigInt("1043680240732074966"),
					OldRate:  string2BigInt("1043375386816533719"),
					Duration: string2BigInt("21600"),
					Expires:  string2BigInt("1689845039"),
				},
			},
			ProtocolFeePercentageCacheSwapType:  big.NewInt(0),
			ProtocolFeePercentageCacheYieldType: big.NewInt(0),
		},
	}

	p, _ := NewComposableStablePoolFull(config)

	testCases := []struct {
		tokenIn   ERC20Details
		amountIn  *big.Int
		tokenOut  ERC20Details
		amountOut *big.Int
	}{
		{
			tokenIn: ERC20Details{
				address:  types.MustAddressFromHex("0x60d604890feaa0b5460b28a424407c24fe89374a"),
				symbol:   "A",
				decimals: 18,
			},
			amountIn: string2BigInt("12000000000000000000"),
			tokenOut: ERC20Details{
				address:  types.MustAddressFromHex("0xbe9895146f7af43049ca1c1ae358b0541ea49704"),
				symbol:   "C",
				decimals: 18,
			},
			amountOut: string2BigInt("11545818036500154428"),
		},
		{
			tokenIn: ERC20Details{
				address:  types.MustAddressFromHex("0x60d604890feaa0b5460b28a424407c24fe89374a"),
				symbol:   "A",
				decimals: 18,
			},
			amountIn: string2BigInt("1000000000000000000"),
			tokenOut: ERC20Details{
				address:  types.MustAddressFromHex("0xbe9895146f7af43049ca1c1ae358b0541ea49704"),
				symbol:   "C",
				decimals: 18,
			},
			amountOut: string2BigInt("962157416748442610"),
		},
		{
			tokenIn: ERC20Details{
				address:  types.MustAddressFromHex("0x9001cbbd96f54a658ff4e6e65ab564ded76a5431"),
				symbol:   "B",
				decimals: 18,
			},
			amountIn: string2BigInt("1000000000000000000"),
			tokenOut: ERC20Details{
				address:  types.MustAddressFromHex("0xbe9895146f7af43049ca1c1ae358b0541ea49704"),
				symbol:   "C",
				decimals: 18,
			},
			amountOut: string2BigInt("963168966727011371"),
		},
		{
			tokenIn: ERC20Details{
				address:  types.MustAddressFromHex("0xbe9895146f7af43049ca1c1ae358b0541ea49704"),
				symbol:   "C",
				decimals: 18,
			},
			amountIn: string2BigInt("1000000000000000000"),
			tokenOut: ERC20Details{
				address:  types.MustAddressFromHex("0x9001cbbd96f54a658ff4e6e65ab564ded76a5431"),
				symbol:   "B",
				decimals: 18,
			},
			amountOut: string2BigInt("1038238373919086616"),
		},
	}

	for i, testcase := range testCases {
		t.Run(fmt.Sprintf("testcase %d, tokenIn %s amountIn %s tokenOut %s amountOut %s", i, testcase.tokenIn.symbol, testcase.amountIn.String(), testcase.tokenOut.symbol, testcase.amountOut.String()), func(t *testing.T) {
			amountOut, _, _ := p.CalcAmountOut(testcase.tokenIn, testcase.tokenOut, testcase.amountIn)
			assert.Equal(t, testcase.amountOut, amountOut)
		})
	}
}

func TestCalculateInvariant(t *testing.T) {
	a := big.NewInt(60000)
	b1 := string2BigInt("50310513788381313281")
	b2 := string2BigInt("19360701460293571158")
	b3 := string2BigInt("58687814461000000000000")

	balances := []*big.Int{
		b1, b2, b3,
	}
	_, err := CalculateInvariant(a, balances, false)
	assert.Equal(t, err, fmt.Errorf("STABLE_INVARIANT_DIDNT_CONVERGE"))
}
