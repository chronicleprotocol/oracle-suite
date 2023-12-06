package origin

import (
	"fmt"

	"github.com/chronicleprotocol/oracle-suite/pkg/util/bn"
)

const AmpPrecision = 1e3

var ampPrecision = bn.DecFloatPoint(AmpPrecision)

var STABLE_INVARIANT_DIDNT_CONVERGE = fmt.Errorf("STABLE_INVARIANT_DIDNT_CONVERGE")     //nolint:revive,stylecheck
var STABLE_GET_BALANCE_DIDNT_CONVERGE = fmt.Errorf("STABLE_GET_BALANCE_DIDNT_CONVERGE") //nolint:revive,stylecheck

// Note on unchecked arithmetic:
// This contract performs a large number of additions, subtractions, multiplications and divisions, often inside
// loops. Since many of these operations are gas-sensitive (as they happen e.g. during a swap), it is important to
// not make any unnecessary checks. We rely on a set of invariants to avoid having to use checked arithmetic (the
// Math library), including:
//  - the number of tokens is bounded by _MAX_STABLE_TOKENS
//  - the amplification parameter is bounded by _MAX_AMP * _AMP_PRECISION, which fits in 23 bits
//  - the token balances are bounded by 2^112 (guaranteed by the Vault) times 1e18 (the maximum scaling factor),
//    which fits in 172 bits
//
// This means e.g. we can safely multiply a balance by the amplification parameter without worrying about overflow.

// About swap fees on joins and exits:
// Any join or exit that is not perfectly balanced (e.g. all single token joins or exits) is mathematically
// equivalent to a perfectly balanced join or  exit followed by a series of swaps. Since these swaps would charge
// swap fees, it follows that (some) joins and exits should as well.
// On these operations, we split the token amounts in 'taxable' and 'non-taxable' portions, where the 'taxable' part
// is the one to which swap fees are applied.

// Computes the invariant given the current balances, using the Newton-Raphson approximation.
// The amplification parameter equals: A n^(n-1)
// See: https://github.com/curvefi/curve-contract/blob/b0bbf77f8f93c9c5f4e415bce9cd71f0cdee960e/contracts/pool-templates/base/SwapTemplateBase.vy#L206
// solhint-disable-previous-line max-line-length
// Reference: https://github.com/balancer/balancer-v2-monorepo/blob/master/pkg/pool-stable/contracts/StableMath.sol#L57
func _calculateInvariant(amplificationParameter *bn.DecFloatPointNumber, balances []*bn.DecFloatPointNumber) (
	*bn.DecFloatPointNumber,
	error,
) {
	/**********************************************************************************************
	// invariant                                                                                 //
	// D = invariant                                                  D^(n+1)                    //
	// A = amplification coefficient      A  n^n S + D = A D n^n + -----------                   //
	// S = sum of balances                                             n^n P                     //
	// P = product of balances                                                                   //
	// n = number of tokens                                                                      //
	**********************************************************************************************/

	// Always round down, to match Vyper's arithmetic (which always truncates).

	var sum = bnZero // S in the Curve version
	var numTokens = len(balances)
	var numTokensBi = bn.DecFloatPoint(numTokens)
	for i := 0; i < numTokens; i++ {
		sum = sum.Add(balances[i])
	}
	if sum.Cmp(bnZero) == 0 {
		return bnZero, nil
	}
	var prevInvariant *bn.DecFloatPointNumber                   // Dprev in the Curve version
	var invariant = sum                                         // D in the Curve version
	var ampTimesTotal = amplificationParameter.Mul(numTokensBi) // Ann in the Curve version
	for i := 0; i < 255; i++ {
		var DP = invariant // D_P
		for j := 0; j < numTokens; j++ {
			// (D_P * invariant) / (balances[j] * numTokens)
			DP = DP.Mul(invariant).DivDown(balances[j].Mul(numTokensBi))
		}
		prevInvariant = invariant
		// ((ampTimesTotal * sum) / AMP_PRECISION + D_P * numTokens) * invariant
		numerator := ampTimesTotal.Mul(sum).Mul(invariant).DivDown(ampPrecision).Add(
			DP.Mul(numTokensBi).Mul(invariant))
		// ((ampTimesTotal - _AMP_PRECISION) * invariant) / _AMP_PRECISION + (numTokens + 1) * D_P
		denominator := ampTimesTotal.Sub(ampPrecision).Mul(invariant).DivDown(ampPrecision).Add(
			numTokensBi.Add(bnOne).Mul(DP))
		invariant = numerator.DivDown(denominator)
		if invariant.Cmp(prevInvariant) > 0 {
			if invariant.Sub(prevInvariant).Cmp(bnOne) <= 0 {
				return invariant, nil
			}
		} else if prevInvariant.Sub(invariant).Cmp(bnOne) <= 0 {
			return invariant, nil
		}
	}
	return nil, STABLE_INVARIANT_DIDNT_CONVERGE
}

// _calcBptOutGivenExactTokensIn implements same functionality with the following url:
// Reference: https://github.com/balancer/balancer-v2-monorepo/blob/master/pkg/pool-stable/contracts/StableMath.sol#L201
func _calcBptOutGivenExactTokensIn(
	amp *bn.DecFloatPointNumber,
	balances []*bn.DecFloatPointNumber,
	amountsIn []*bn.DecFloatPointNumber,
	bptTotalSupply, invariant, swapFeePercentage *bn.DecFloatPointNumber,
) (*bn.DecFloatPointNumber, *bn.DecFloatPointNumber, error) {

	// BPT out, so we round down overall.

	// First loop calculates the sum of all token balances, which will be used to calculate
	// the current weights of each token, relative to this sum
	feeAmountIn := bn.DecFloatPoint(0)
	sumBalances := bn.DecFloatPoint(0)
	for _, balance := range balances {
		sumBalances = sumBalances.Add(balance)
	}

	// Calculate the weighted balance ratio without considering fees
	balanceRatiosWithFee := make([]*bn.DecFloatPointNumber, len(amountsIn))
	// The weighted sum of token balance ratios with fee
	invariantRatioWithFees := bn.DecFloatPoint(0)
	for i, balance := range balances {
		currentWeight := _divDownFixed18(balance, sumBalances)
		balanceRatiosWithFee[i] = _divDownFixed18(balance.Add(amountsIn[i]), balance)
		invariantRatioWithFees = invariantRatioWithFees.Add(_mulDownFixed18(balanceRatiosWithFee[i], currentWeight))
	}

	// Second loop calculates new amounts in, taking into account the fee on the percentage excess
	newBalances := make([]*bn.DecFloatPointNumber, len(balances))
	for i, balance := range balances {
		var amountInWithoutFee *bn.DecFloatPointNumber
		// Check if the balance ratio is greater than the ideal ratio to charge fees or not
		if balanceRatiosWithFee[i].Cmp(invariantRatioWithFees) > 0 {
			nonTaxableAmount := _mulDownFixed18(balance, invariantRatioWithFees.Sub(bnOne))
			taxableAmount := amountsIn[i].Sub(nonTaxableAmount)
			// No need to use checked arithmetic for the swap fee, it is guaranteed to be lower than 50%
			amountInWithoutFee = nonTaxableAmount.Add(_mulDownFixed18(taxableAmount, bnOne.Sub(swapFeePercentage)))
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
	invariantRatio := _divDownFixed18(newInvariant, invariant)
	// If the invariant didn't increase for any reason, we simply don't mint BPT
	if invariantRatio.Cmp(bnOne) > 0 {
		return _mulDownFixed18(bptTotalSupply, invariantRatio.Sub(bnOne)),
			feeAmountIn,
			nil
	}
	return bnZero, feeAmountIn, nil
}

// _calcTokenOutGivenExactBptIn implements same functionality with the following url:
// Reference: https://github.com/balancer/balancer-v2-monorepo/blob/master/pkg/pool-stable/contracts/StableMath.sol#L354
func _calcTokenOutGivenExactBptIn(
	amp *bn.DecFloatPointNumber,
	balances []*bn.DecFloatPointNumber,
	tokenIndex int,
	bptAmountIn *bn.DecFloatPointNumber,
	bptTotalSupply, currentInvariant, swapFeePercentage *bn.DecFloatPointNumber,
) (*bn.DecFloatPointNumber, *bn.DecFloatPointNumber, error) {
	// Token out, so we round down overall.
	newInvariant := _mulUpFixed18(_divUpFixed18(bptTotalSupply.Sub(bptAmountIn), bptTotalSupply), currentInvariant)
	// Calculate amount out without fee
	newBalanceTokenIndex, err := _getTokenBalanceGivenInvariantAndAllOtherBalances(amp, balances, newInvariant, tokenIndex)
	if err != nil {
		return nil, nil, err
	}
	amountOutWithoutFee := balances[tokenIndex].Sub(newBalanceTokenIndex)

	// First calculate the sum of all token balances, which will be used to calculate
	// the current weight of each token
	sumBalances := bn.DecFloatPoint(0)
	for _, balance := range balances {
		sumBalances = sumBalances.Add(balance)
	}

	// We can now compute how much excess balance is being withdrawn as a result of the virtual swaps, which result
	// in swap fees.
	currentWeight := _divDownFixed18(balances[tokenIndex], sumBalances)
	taxablePercentage := _complementFixed(currentWeight)

	// Swap fees are typically charged on 'token in', but there is no 'token in' here, so we apply it
	// to 'token out'. This results in slightly larger price impact. Fees are rounded up.
	taxableAmount := _mulUpFixed18(amountOutWithoutFee, taxablePercentage)
	nonTaxableAmount := amountOutWithoutFee.Sub(taxableAmount)

	// No need to use checked arithmetic for the swap fee, it is guaranteed to be lower than 50%
	feeOfTaxableAmount := _mulDownFixed18(taxableAmount, bnOne.Sub(swapFeePercentage))
	feeAmount := taxableAmount.Sub(feeOfTaxableAmount)
	return nonTaxableAmount.Add(feeOfTaxableAmount), feeAmount, nil
}

// This function calculates the balance of a given token (tokenIndex)
// given all the other balances and the invariant
// Reference: https://github.com/balancer/balancer-v2-monorepo/blob/master/pkg/pool-stable/contracts/StableMath.sol#L399
func _getTokenBalanceGivenInvariantAndAllOtherBalances(
	amplificationParameter *bn.DecFloatPointNumber,
	balances []*bn.DecFloatPointNumber,
	invariant *bn.DecFloatPointNumber,
	tokenIndex int,
) (*bn.DecFloatPointNumber, error) {
	// Rounds result up overall
	var nTokens = len(balances)
	var nTokensBi = bn.DecFloatPoint(nTokens)
	var ampTimesTotal = amplificationParameter.Mul(nTokensBi)
	var sum = balances[0]
	var PD = balances[0].Mul(nTokensBi) // P_D
	for j := 1; j < nTokens; j++ {
		PD = PD.Mul(balances[j]).Mul(nTokensBi).DivDown(invariant)
		sum = sum.Add(balances[j])
	}
	// No need to use safe math, based on the loop above `sum` is greater than or equal to `balances[tokenIndex]`
	sum = sum.Sub(balances[tokenIndex])
	var inv2 = invariant.Mul(invariant)
	// We remove the balance from c by multiplying it
	var c = inv2.DivUp(ampTimesTotal.Mul(PD)).Mul(ampPrecision).Mul(balances[tokenIndex])
	var b = sum.Add(invariant.DivDown(ampTimesTotal).Mul(ampPrecision))
	// We iterate to find the balance
	var prevTokenBalance *bn.DecFloatPointNumber
	// We multiply the first iteration outside the loop with the invariant to set the value of the
	// initial approximation.
	var tokenBalance = inv2.Add(c).DivUp(invariant.Add(b))
	for i := 0; i < 255; i++ {
		prevTokenBalance = tokenBalance
		tokenBalance =
			tokenBalance.Mul(tokenBalance).Add(c).DivUp(
				tokenBalance.Mul(bnTwo).Add(b).Sub(invariant))
		if tokenBalance.Cmp(prevTokenBalance) > 0 {
			if tokenBalance.Sub(prevTokenBalance).Cmp(bnOne) <= 0 {
				return tokenBalance, nil
			}
		} else if prevTokenBalance.Sub(tokenBalance).Cmp(bnOne) <= 0 {
			return tokenBalance, nil
		}
	}
	return nil, STABLE_GET_BALANCE_DIDNT_CONVERGE
}

// Computes how many tokens can be taken out of a pool if `tokenAmountIn` are sent, given the current balances.
// The amplification parameter equals: A n^(n-1)
// Reference: https://github.com/balancer/balancer-v2-monorepo/blob/master/pkg/pool-stable/contracts/StableMath.sol#L124
func _calcOutGivenIn(
	amplificationParameter *bn.DecFloatPointNumber,
	balances []*bn.DecFloatPointNumber,
	tokenIndexIn int,
	tokenIndexOut int,
	tokenAmountIn *bn.DecFloatPointNumber,
	invariant *bn.DecFloatPointNumber,
) (*bn.DecFloatPointNumber, error) {

	/**************************************************************************************************************
	// outGivenIn token x for y - polynomial equation to solve                                                   //
	// ay = amount out to calculate                                                                              //
	// by = balance token out                                                                                    //
	// y = by - ay (finalBalanceOut)                                                                             //
	// D = invariant                                               D                     D^(n+1)                 //
	// A = amplification coefficient               y^2 + ( S + ----------  - D) * y -  ------------- = 0         //
	// n = number of tokens                                    (A * n^n)               A * n^2n * P              //
	// S = sum of final balances but y                                                                           //
	// P = product of final balances but y                                                                       //
	**************************************************************************************************************/

	// Amount out, so we round down overall.
	balances[tokenIndexIn] = balances[tokenIndexIn].Add(tokenAmountIn)
	var finalBalanceOut, err = _getTokenBalanceGivenInvariantAndAllOtherBalances(
		amplificationParameter, balances, invariant, tokenIndexOut)
	if err != nil {
		return nil, err
	}
	// No need to use checked arithmetic since `tokenAmountIn` was actually added to the same balance right before
	// calling `_getTokenBalanceGivenInvariantAndAllOtherBalances` which doesn't alter the balances array.
	balances[tokenIndexIn] = balances[tokenIndexIn].Sub(tokenAmountIn)
	return balances[tokenIndexOut].Sub(finalBalanceOut).Sub(bnOne), nil
}
