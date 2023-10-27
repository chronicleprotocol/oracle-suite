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
const composableStablePrecision = 18

var ampPrecision = bn.DecFloatPoint(AmpPrecision)

func _complementFixed(x *bn.DecFloatPointNumber) *bn.DecFloatPointNumber {
	if x.Cmp(bigIntEther) < 0 {
		return bigIntEther.Sub(x)
	}
	return bigIntZero
}

// https://github.com/balancer/balancer-v2-monorepo/blob/master/pkg/pool-stable/contracts/StableMath.sol#L57
func _calculateInvariant(amp *bn.DecFloatPointNumber, balances []*bn.DecFloatPointNumber) (
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
			PD = PD.Mul(balances[j]).Mul(numTokensBi).Div(invariant)
		}
		prevInvariant = invariant
		numerator := numTokensBi.Mul(invariant).Mul(invariant).Add(
			ampTotal.Mul(sum).Mul(PD).Div(ampPrecision))
		denominator := numTokensBi.Add(bigIntOne).Mul(invariant).Add(
			ampTotal.Sub(ampPrecision).Mul(PD).Div(ampPrecision))
		invariant = numerator.Div(denominator)
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
		currentWeight := balance.DivDownFixed(sumBalances, composableStablePrecision)
		balanceRatiosWithFee[i] = balance.Add(amountsIn[i]).DivDownFixed(balance, composableStablePrecision)
		invariantRatioWithFees = invariantRatioWithFees.Add(
			balanceRatiosWithFee[i].MulDownFixed(currentWeight, composableStablePrecision))
	}

	newBalances := make([]*bn.DecFloatPointNumber, len(balances))
	for i, balance := range balances {
		var amountInWithoutFee *bn.DecFloatPointNumber
		if balanceRatiosWithFee[i].Cmp(invariantRatioWithFees) > 0 {
			nonTaxableAmount :=
				balance.MulDownFixed(
					invariantRatioWithFees.Sub(bn.DecFloatPoint(ether)), composableStablePrecision)
			taxableAmount := amountsIn[i].Sub(nonTaxableAmount)
			amountInWithoutFee = nonTaxableAmount.Add(
				taxableAmount.MulDownFixed(
					bn.DecFloatPoint(ether).Sub(swapFeePercentage), composableStablePrecision),
			)
		} else {
			amountInWithoutFee = amountsIn[i]
		}
		feeAmountIn = feeAmountIn.Add(amountsIn[i].Sub(amountInWithoutFee))
		newBalances[i] = balance.Add(amountInWithoutFee)
	}

	newInvariant, err := _calculateInvariant(amp, newBalances)
	if err != nil {
		return nil, nil, err
	}

	invariantRatio := newInvariant.DivDownFixed(invariant, composableStablePrecision)
	if invariantRatio.Cmp(bn.DecFloatPoint(ether)) > 0 {
		return bptTotalSupply.MulDownFixed(invariantRatio.Sub(bn.DecFloatPoint(ether)), composableStablePrecision),
			feeAmountIn,
			nil
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

	newInvariant :=
		bptTotalSupply.Sub(bptAmountIn).DivUpFixed(bptTotalSupply, composableStablePrecision).MulUpFixed(
			invariant, composableStablePrecision)
	newBalanceTokenIndex, err := _getTokenBalanceGivenInvariantAndAllOtherBalances(amp, balances, newInvariant, tokenIndex)
	if err != nil {
		return nil, nil, err
	}
	amountOutWithoutFee := balances[tokenIndex].Sub(newBalanceTokenIndex)

	sumBalances := bn.DecFloatPoint(0)
	for _, balance := range balances {
		sumBalances = sumBalances.Add(balance)
	}

	currentWeight := balances[tokenIndex].DivDownFixed(sumBalances, composableStablePrecision)
	taxablePercentage := _complementFixed(currentWeight)

	taxableAmount := amountOutWithoutFee.MulUpFixed(taxablePercentage, composableStablePrecision)
	nonTaxableAmount := amountOutWithoutFee.Sub(taxableAmount)

	feeOfTaxableAmount :=
		taxableAmount.MulDownFixed(
			bn.DecFloatPoint(ether).Sub(swapFeePercentage), composableStablePrecision)

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
		PD = PD.Mul(balances[j]).Mul(nTokensBi).Div(invariant)
		sum = sum.Add(balances[j])
	}
	sum = sum.Sub(balances[tokenIndex])
	var inv2 = invariant.Mul(invariant)
	var c = inv2.DivUp(ampTotal.Mul(PD)).Mul(ampPrecision).Mul(balances[tokenIndex])
	var b = sum.Add(invariant.Div(ampTotal).Mul(ampPrecision))
	var prevTokenBalance *bn.DecFloatPointNumber
	var tokenBalance = inv2.Add(c).DivUp(invariant.Add(b))
	for i := 0; i < 255; i++ {
		prevTokenBalance = tokenBalance
		tokenBalance =
			tokenBalance.Mul(tokenBalance).Add(c).DivUp(
				tokenBalance.Mul(bigIntTwo).Add(b).Sub(invariant))
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
