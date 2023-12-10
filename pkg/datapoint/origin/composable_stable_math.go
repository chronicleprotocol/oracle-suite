package origin

import (
	"fmt"

	"github.com/chronicleprotocol/oracle-suite/pkg/util/bn"
)

const AmpPrecision = 1e3

var ampPrecision = bn.DecFixedPoint(AmpPrecision, 0)

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
func _calculateInvariant(amplificationParameter *bn.DecFixedPointNumber, balances []*bn.DecFixedPointNumber) (
	*bn.DecFixedPointNumber,
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
	var numTokensBi = bn.DecFixedPoint(numTokens, 0)
	for i := 0; i < numTokens; i++ {
		sum = sum.Add(balances[i])
	}
	if sum.Cmp(bnZero) == 0 {
		return bnZero, nil
	}
	var prevInvariant *bn.DecFixedPointNumber                   // Dprev in the Curve version
	var invariant = sum                                         // D in the Curve version
	var ampTimesTotal = amplificationParameter.Mul(numTokensBi) // Ann in the Curve version
	for i := 0; i < 255; i++ {
		var DP = invariant // D_P
		for j := 0; j < numTokens; j++ {
			// (D_P * invariant) / (balances[j] * numTokens)
			DP = _divDown(DP.Mul(invariant), balances[j].Mul(numTokensBi))
		}
		prevInvariant = invariant
		// ((ampTimesTotal * sum) / AMP_PRECISION + D_P * numTokens) * invariant
		numerator := _divDown(ampTimesTotal.Mul(sum).Mul(invariant), ampPrecision).Add(
			DP.Mul(numTokensBi).Mul(invariant))
		// ((ampTimesTotal - _AMP_PRECISION) * invariant) / _AMP_PRECISION + (numTokens + 1) * D_P
		denominator := _divDown(ampTimesTotal.Sub(ampPrecision).Mul(invariant), ampPrecision).Add(
			numTokensBi.Add(bnOne).Mul(DP))
		invariant = _divDown(numerator, denominator)
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

// This function calculates the balance of a given token (tokenIndex)
// given all the other balances and the invariant
// Reference: https://github.com/balancer/balancer-v2-monorepo/blob/master/pkg/pool-stable/contracts/StableMath.sol#L399
func _getTokenBalanceGivenInvariantAndAllOtherBalances(
	amplificationParameter *bn.DecFixedPointNumber,
	balances []*bn.DecFixedPointNumber,
	invariant *bn.DecFixedPointNumber,
	tokenIndex int,
) (*bn.DecFixedPointNumber, error) {
	// Rounds result up overall
	var nTokens = len(balances)
	var nTokensBi = bn.DecFixedPoint(nTokens, 0)
	var ampTimesTotal = amplificationParameter.Mul(nTokensBi)
	var sum = balances[0]
	var PD = balances[0].Mul(nTokensBi) // P_D
	for j := 1; j < nTokens; j++ {
		PD = _divDown(PD.Mul(balances[j]).Mul(nTokensBi), invariant)
		sum = sum.Add(balances[j])
	}
	// No need to use safe math, based on the loop above `sum` is greater than or equal to `balances[tokenIndex]`
	sum = sum.Sub(balances[tokenIndex])
	var inv2 = invariant.Mul(invariant)
	// We remove the balance from c by multiplying it
	var c = _divUp(inv2, ampTimesTotal.Mul(PD)).Mul(ampPrecision).Mul(balances[tokenIndex])
	var b = sum.Add(_divDown(invariant, ampTimesTotal).Mul(ampPrecision))
	// We iterate to find the balance
	var prevTokenBalance *bn.DecFixedPointNumber
	// We multiply the first iteration outside the loop with the invariant to set the value of the
	// initial approximation.
	var tokenBalance = _divUp(inv2.Add(c), invariant.Add(b))
	for i := 0; i < 255; i++ {
		prevTokenBalance = tokenBalance
		tokenBalance =
			_divUp(tokenBalance.Mul(tokenBalance).Add(c),
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
	amplificationParameter *bn.DecFixedPointNumber,
	balances []*bn.DecFixedPointNumber,
	tokenIndexIn int,
	tokenIndexOut int,
	tokenAmountIn *bn.DecFixedPointNumber,
	invariant *bn.DecFixedPointNumber,
) (*bn.DecFixedPointNumber, error) {

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
