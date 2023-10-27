package origin

import (
	"fmt"
	"github.com/chronicleprotocol/oracle-suite/pkg/util/bn"
)

var bigIntZero = bn.DecFloatPoint(0)
var bigIntOne = bn.DecFloatPoint(1)
var bigIntTwo = bn.DecFloatPoint(2)
var bigIntEther = bn.DecFloatPoint(ether)

const AmpPrecision = 1e3
const ComposableStablePrecision = 18

var ampPrecision = bn.DecFloatPoint(AmpPrecision)

func _mulDownFixed(a *bn.DecFloatPointNumber, b *bn.DecFloatPointNumber) *bn.DecFloatPointNumber {
	var ret = a.Mul(b)
	return ret.Deflate(ComposableStablePrecision)
}

func _mulUpFixed(a *bn.DecFloatPointNumber, b *bn.DecFloatPointNumber) *bn.DecFloatPointNumber {
	var ret = a.Mul(b)
	if ret.Cmp(bigIntZero) == 0 {
		return ret
	}
	return ret.Sub(bigIntOne).Deflate(ComposableStablePrecision).Add(bigIntOne)
}

func _divRounding(a *bn.DecFloatPointNumber, b *bn.DecFloatPointNumber, roundUp bool) *bn.DecFloatPointNumber {
	if roundUp {
		return _divUp(a, b)
	}
	return _divDown(a, b)
}

func _divDown(a *bn.DecFloatPointNumber, b *bn.DecFloatPointNumber) *bn.DecFloatPointNumber {
	return a.Div(b)
}

func _divUp(a *bn.DecFloatPointNumber, b *bn.DecFloatPointNumber) *bn.DecFloatPointNumber {
	if a.Cmp(bigIntZero) == 0 {
		return bigIntZero
	}
	return a.Sub(bigIntOne).Div(b).Add(bigIntOne)
}

func _divUpFixed(a *bn.DecFloatPointNumber, b *bn.DecFloatPointNumber) *bn.DecFloatPointNumber {
	if a.Cmp(bigIntZero) == 0 {
		return bigIntZero
	}
	aInflated := a.Inflate(ComposableStablePrecision)
	return aInflated.Sub(bigIntOne).Div(b).Add(bigIntOne)
}

func _divDownFixed(a *bn.DecFloatPointNumber, b *bn.DecFloatPointNumber) *bn.DecFloatPointNumber {
	if a.Cmp(bigIntZero) == 0 {
		return bigIntZero
	}
	var ret = a.Inflate(ComposableStablePrecision)
	return ret.Div(b)
}

func _complementFixed(x *bn.DecFloatPointNumber) *bn.DecFloatPointNumber {
	if x.Cmp(bigIntEther) < 0 {
		return bigIntEther.Sub(x)
	}
	return bn.DecFloatPoint(0)
}

// https://github.com/balancer/balancer-v2-monorepo/blob/master/pkg/pool-stable/contracts/StableMath.sol#L57
func _calculateInvariant(amp *bn.DecFloatPointNumber, balances []*bn.DecFloatPointNumber, roundUp bool) ( //nolint:unparam
	*bn.DecFloatPointNumber,
	error,
) {

	var sum = bigIntZero
	var numTokens = len(balances)
	var numTokensBi = bn.DecFloatPoint(numTokens)
	for i := 0; i < numTokens; i++ {
		sum = sum.Add(balances[i])
	}
	if sum.Cmp(bigIntZero) == 0 {
		return bigIntZero, nil
	}
	var prevInvariant *bn.DecFloatPointNumber
	var invariant = sum
	var ampTotal = amp.Mul(numTokensBi)
	for i := 0; i < 255; i++ {
		var PD = balances[0].Mul(numTokensBi) // P_D
		for j := 1; j < numTokens; j++ {
			PD = _divRounding(PD.Mul(balances[j]).Mul(numTokensBi), invariant, roundUp)
		}
		prevInvariant = invariant
		invariant = _divRounding(
			numTokensBi.Mul(invariant).Mul(invariant).Add(
				_divRounding(ampTotal.Mul(sum).Mul(PD), ampPrecision, roundUp)),
			numTokensBi.Add(bigIntOne).Mul(invariant).Add(
				_divRounding(ampTotal.Sub(ampPrecision).Mul(PD), ampPrecision, !roundUp)),
			roundUp,
		)
		if invariant.Cmp(prevInvariant) > 0 {
			if invariant.Sub(prevInvariant).Cmp(bigIntOne) <= 0 {
				return invariant, nil
			}
		} else if prevInvariant.Sub(invariant).Cmp(bigIntOne) <= 0 {
			return invariant, nil
		}
	}
	return nil, fmt.Errorf("STABLE_INVARIANT_DIDNT_CONVERGE")
}

// _calcBptOutGivenExactTokensIn implements same functionality with the following url:
// https://github.com/balancer/balancer-v2-monorepo/blob/master/pkg/pool-stable/contracts/StableMath.sol#L201
func _calcBptOutGivenExactTokensIn(
	amp *bn.DecFloatPointNumber,
	balances []*bn.DecFloatPointNumber,
	amountsIn []*bn.DecFloatPointNumber,
	bptTotalSupply, invariant, swapFeePercentage *bn.DecFloatPointNumber,
) (*bn.DecFloatPointNumber, *bn.DecFloatPointNumber, error) {

	feeAmountIn := bn.DecFloatPoint(0)
	sumBalances := bn.DecFloatPoint(0)
	for _, balance := range balances {
		sumBalances = sumBalances.Add(balance)
	}

	balanceRatiosWithFee := make([]*bn.DecFloatPointNumber, len(amountsIn))
	invariantRatioWithFees := bn.DecFloatPoint(0)
	for i, balance := range balances {
		currentWeight := _divDownFixed(balance, sumBalances)
		balanceRatiosWithFee[i] = _divDownFixed(balance.Add(amountsIn[i]), balance)
		invariantRatioWithFees = invariantRatioWithFees.Add(_mulDownFixed(balanceRatiosWithFee[i], currentWeight))
	}

	newBalances := make([]*bn.DecFloatPointNumber, len(balances))
	for i, balance := range balances {
		var amountInWithoutFee *bn.DecFloatPointNumber
		if balanceRatiosWithFee[i].Cmp(invariantRatioWithFees) > 0 {
			nonTaxableAmount := _mulDownFixed(balance,
				invariantRatioWithFees.Sub(bn.DecFloatPoint(ether)))
			taxableAmount := amountsIn[i].Sub(nonTaxableAmount)
			amountInWithoutFee = nonTaxableAmount.Add(_mulDownFixed(
				taxableAmount,
				bn.DecFloatPoint(ether).Sub(swapFeePercentage),
			))
		} else {
			amountInWithoutFee = amountsIn[i]
		}
		feeAmountIn = feeAmountIn.Add(amountsIn[i].Sub(amountInWithoutFee))
		newBalances[i] = balance.Add(amountInWithoutFee)
	}

	newInvariant, err := _calculateInvariant(amp, newBalances, false)
	if err != nil {
		return nil, nil, err
	}

	invariantRatio := _divDownFixed(newInvariant, invariant)
	if invariantRatio.Cmp(bn.DecFloatPoint(ether)) > 0 {
		return _mulDownFixed(bptTotalSupply, invariantRatio.Sub(bn.DecFloatPoint(ether))), feeAmountIn, nil
	}
	return bn.DecFloatPoint(0), feeAmountIn, nil
}

// _calcTokenOutGivenExactBptIn implements same functionality with the following url:
// https://github.com/balancer/balancer-v2-monorepo/blob/master/pkg/pool-stable/contracts/StableMath.sol#L354
func _calcTokenOutGivenExactBptIn(
	amp *bn.DecFloatPointNumber,
	balances []*bn.DecFloatPointNumber,
	tokenIndex int,
	bptAmountIn *bn.DecFloatPointNumber,
	bptTotalSupply, invariant, swapFeePercentage *bn.DecFloatPointNumber,
) (*bn.DecFloatPointNumber, *bn.DecFloatPointNumber, error) {

	newInvariant := _mulUpFixed(_divUpFixed(bptTotalSupply.Sub(bptAmountIn), bptTotalSupply), invariant)
	newBalanceTokenIndex, err := _getTokenBalanceGivenInvariantAndAllOtherBalances(amp, balances, newInvariant, tokenIndex)
	if err != nil {
		return nil, nil, err
	}
	amountOutWithoutFee := balances[tokenIndex].Sub(newBalanceTokenIndex)

	sumBalances := bn.DecFloatPoint(0)
	for _, balance := range balances {
		sumBalances = sumBalances.Add(balance)
	}

	currentWeight := _divDownFixed(balances[tokenIndex], sumBalances)
	taxablePercentage := _complementFixed(currentWeight)

	taxableAmount := _mulUpFixed(amountOutWithoutFee, taxablePercentage)
	nonTaxableAmount := amountOutWithoutFee.Sub(taxableAmount)

	feeOfTaxableAmount := _mulDownFixed(
		taxableAmount,
		bn.DecFloatPoint(ether).Sub(swapFeePercentage),
	)

	feeAmount := taxableAmount.Sub(feeOfTaxableAmount)
	return nonTaxableAmount.Add(feeOfTaxableAmount), feeAmount, nil
}

// https://github.com/balancer/balancer-v2-monorepo/blob/master/pkg/pool-stable/contracts/StableMath.sol#L399
func _getTokenBalanceGivenInvariantAndAllOtherBalances(
	a *bn.DecFloatPointNumber,
	balances []*bn.DecFloatPointNumber,
	invariant *bn.DecFloatPointNumber,
	tokenIndex int,
) (*bn.DecFloatPointNumber, error) {

	var nTokens = len(balances)
	var nTokensBi = bn.DecFloatPoint(nTokens)
	var ampTotal = a.Mul(nTokensBi)
	var sum = balances[0]
	var PD = balances[0].Mul(nTokensBi) // P_D
	for j := 1; j < nTokens; j++ {
		PD = _divDown(PD.Mul(balances[j]).Mul(nTokensBi), invariant)
		sum = sum.Add(balances[j])
	}
	sum = sum.Sub(balances[tokenIndex])
	var inv2 = invariant.Mul(invariant)
	var c = _divUp(inv2, ampTotal.Mul(PD)).Mul(ampPrecision).Mul(balances[tokenIndex])
	var b = sum.Add(_divDown(invariant, ampTotal).Mul(ampPrecision))
	var prevTokenBalance *bn.DecFloatPointNumber
	var tokenBalance = _divUp(inv2.Add(c), invariant.Add(b))
	for i := 0; i < 255; i++ {
		prevTokenBalance = tokenBalance
		tokenBalance = _divUp(
			tokenBalance.Mul(tokenBalance).Add(c),
			tokenBalance.Mul(bigIntTwo).Add(b).Sub(invariant),
		)
		if tokenBalance.Cmp(prevTokenBalance) > 0 {
			if tokenBalance.Sub(prevTokenBalance).Cmp(bigIntOne) <= 0 {
				return tokenBalance, nil
			}
		} else if prevTokenBalance.Sub(tokenBalance).Cmp(bigIntOne) <= 0 {
			return tokenBalance, nil
		}
	}
	return nil, fmt.Errorf("STABLE_GET_BALANCE_DIDNT_CONVERGE")
}

// https://github.com/balancer/balancer-v2-monorepo/blob/master/pkg/pool-stable/contracts/StableMath.sol#L124
func _calcOutGivenIn(
	a *bn.DecFloatPointNumber,
	balances []*bn.DecFloatPointNumber,
	tokenIndexIn int,
	tokenIndexOut int,
	tokenAmountIn *bn.DecFloatPointNumber,
	invariant *bn.DecFloatPointNumber,
) (*bn.DecFloatPointNumber, error) {

	balances[tokenIndexIn] = balances[tokenIndexIn].Add(tokenAmountIn)
	var finalBalanceOut, err = _getTokenBalanceGivenInvariantAndAllOtherBalances(a, balances, invariant, tokenIndexOut)
	if err != nil {
		return nil, err
	}
	balances[tokenIndexIn] = balances[tokenIndexIn].Sub(tokenAmountIn)
	return balances[tokenIndexOut].Sub(finalBalanceOut).Sub(bigIntOne), nil
}
