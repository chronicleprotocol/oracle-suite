package origin

import (
	"fmt"
	"math/big"
)

var bigIntZero = big.NewInt(0)
var bigIntOne = big.NewInt(1)
var bigIntTwo = big.NewInt(2)
var bigIntEther = big.NewInt(ether)

const AmpPrecision = 1e3

var ampPrecision = big.NewInt(AmpPrecision)

func _mulDownFixed(a *big.Int, b *big.Int) *big.Int {
	var ret = new(big.Int).Mul(a, b)
	return new(big.Int).Div(ret, bigIntEther)
}

func _mulUpFixed(a *big.Int, b *big.Int) *big.Int {
	var ret = new(big.Int).Mul(a, b)
	if ret.Cmp(bigIntZero) == 0 {
		return ret
	}
	return new(big.Int).Add(new(big.Int).Div(new(big.Int).Sub(ret, bigIntOne), bigIntEther), bigIntOne)
}

func _divRounding(a *big.Int, b *big.Int, roundUp bool) *big.Int {
	if roundUp {
		return _divUp(a, b)
	}
	return _divDown(a, b)
}

func _divDown(a *big.Int, b *big.Int) *big.Int {
	return new(big.Int).Div(a, b)
}

func _divUp(a *big.Int, b *big.Int) *big.Int {
	if a.Cmp(bigIntZero) == 0 {
		return bigIntZero
	}
	return new(big.Int).Add(new(big.Int).Div(new(big.Int).Sub(a, bigIntOne), b), bigIntOne)
}

func _divUpFixed(a *big.Int, b *big.Int) *big.Int {
	if a.Cmp(bigIntZero) == 0 {
		return bigIntZero
	}
	aInflated := new(big.Int).Mul(a, bigIntEther)
	return new(big.Int).Add(new(big.Int).Div(new(big.Int).Sub(aInflated, bigIntOne), b), bigIntOne)
}

func _divDownFixed(a *big.Int, b *big.Int) *big.Int {
	if a.Cmp(bigIntZero) == 0 {
		return bigIntZero
	}
	var ret = new(big.Int).Mul(a, bigIntEther)
	return new(big.Int).Div(ret, b)
}

func _complementFixed(x *big.Int) *big.Int {
	if x.Cmp(bigIntEther) < 0 {
		return new(big.Int).Sub(bigIntEther, x)
	}
	return big.NewInt(0)
}

// https://github.com/balancer/balancer-v2-monorepo/blob/master/pkg/pool-stable/contracts/StableMath.sol#L57
func _calculateInvariant(amp *big.Int, balances []*big.Int, roundUp bool) (*big.Int, error) { //nolint:unparam
	var sum = bigIntZero
	var numTokens = len(balances)
	var numTokensBi = big.NewInt(int64(numTokens))
	for i := 0; i < numTokens; i++ {
		sum = new(big.Int).Add(sum, balances[i])
	}
	if sum.Cmp(bigIntZero) == 0 {
		return bigIntZero, nil
	}
	var prevInvariant *big.Int
	var invariant = sum
	var ampTotal = new(big.Int).Mul(amp, numTokensBi)
	for i := 0; i < 255; i++ {
		var PD = new(big.Int).Mul(balances[0], numTokensBi) // P_D
		for j := 1; j < numTokens; j++ {
			PD = _divRounding(new(big.Int).Mul(new(big.Int).Mul(PD, balances[j]), numTokensBi), invariant, roundUp)
		}
		prevInvariant = invariant
		invariant = _divRounding(
			new(big.Int).Add(
				new(big.Int).Mul(new(big.Int).Mul(numTokensBi, invariant), invariant),
				_divRounding(new(big.Int).Mul(new(big.Int).Mul(ampTotal, sum), PD), ampPrecision, roundUp),
			),
			new(big.Int).Add(
				new(big.Int).Mul(new(big.Int).Add(numTokensBi, bigIntOne), invariant),
				_divRounding(new(big.Int).Mul(new(big.Int).Sub(ampTotal, ampPrecision), PD), ampPrecision, !roundUp),
			),
			roundUp,
		)
		if invariant.Cmp(prevInvariant) > 0 {
			if new(big.Int).Sub(invariant, prevInvariant).Cmp(bigIntOne) <= 0 {
				return invariant, nil
			}
		} else if new(big.Int).Sub(prevInvariant, invariant).Cmp(bigIntOne) <= 0 {
			return invariant, nil
		}
	}
	return nil, fmt.Errorf("STABLE_INVARIANT_DIDNT_CONVERGE")
}

// _calcBptOutGivenExactTokensIn implements same functionality with the following url:
// https://github.com/balancer/balancer-v2-monorepo/blob/master/pkg/pool-stable/contracts/StableMath.sol#L201
func _calcBptOutGivenExactTokensIn(
	amp *big.Int,
	balances []*big.Int,
	amountsIn []*big.Int,
	bptTotalSupply, invariant, swapFeePercentage *big.Int,
) (*big.Int, *big.Int, error) {

	feeAmountIn := big.NewInt(0)
	sumBalances := big.NewInt(0)
	for _, balance := range balances {
		sumBalances.Add(sumBalances, balance)
	}

	balanceRatiosWithFee := make([]*big.Int, len(amountsIn))
	invariantRatioWithFees := big.NewInt(0)
	for i, balance := range balances {
		currentWeight := _divDownFixed(balance, sumBalances)
		balanceRatiosWithFee[i] = _divDownFixed(new(big.Int).Add(balance, amountsIn[i]), balance)
		invariantRatioWithFees.Add(invariantRatioWithFees, _mulDownFixed(balanceRatiosWithFee[i], currentWeight))
	}

	newBalances := make([]*big.Int, len(balances))
	for i, balance := range balances {
		var amountInWithoutFee *big.Int
		if balanceRatiosWithFee[i].Cmp(invariantRatioWithFees) > 0 {
			nonTaxableAmount := _mulDownFixed(balance, new(big.Int).Sub(invariantRatioWithFees, big.NewInt(ether)))
			taxableAmount := new(big.Int).Sub(amountsIn[i], nonTaxableAmount)
			amountInWithoutFee = new(big.Int).Add(
				nonTaxableAmount,
				_mulDownFixed(
					taxableAmount,
					new(big.Int).Sub(big.NewInt(ether), swapFeePercentage),
				),
			)
		} else {
			amountInWithoutFee = amountsIn[i]
		}
		feeAmountIn = feeAmountIn.Add(feeAmountIn, new(big.Int).Sub(amountsIn[i], amountInWithoutFee))
		newBalances[i] = new(big.Int).Add(balance, amountInWithoutFee)
	}

	newInvariant, err := _calculateInvariant(amp, newBalances, false)
	if err != nil {
		return nil, nil, err
	}

	invariantRatio := _divDownFixed(newInvariant, invariant)
	if invariantRatio.Cmp(big.NewInt(ether)) > 0 {
		return _mulDownFixed(bptTotalSupply, new(big.Int).Sub(invariantRatio, big.NewInt(ether))), feeAmountIn, nil
	}
	return big.NewInt(0), feeAmountIn, nil
}

// _calcTokenOutGivenExactBptIn implements same functionality with the following url:
// https://github.com/balancer/balancer-v2-monorepo/blob/master/pkg/pool-stable/contracts/StableMath.sol#L354
func _calcTokenOutGivenExactBptIn(
	amp *big.Int,
	balances []*big.Int,
	tokenIndex int,
	bptAmountIn *big.Int,
	bptTotalSupply, invariant, swapFeePercentage *big.Int,
) (*big.Int, *big.Int, error) {

	newInvariant := _mulUpFixed(_divUpFixed(new(big.Int).Sub(bptTotalSupply, bptAmountIn), bptTotalSupply), invariant)
	newBalanceTokenIndex, err := _getTokenBalanceGivenInvariantAndAllOtherBalances(amp, balances, newInvariant, tokenIndex)
	if err != nil {
		return nil, nil, err
	}
	amountOutWithoutFee := new(big.Int).Sub(balances[tokenIndex], newBalanceTokenIndex)

	sumBalances := big.NewInt(0)
	for _, balance := range balances {
		sumBalances.Add(sumBalances, balance)
	}

	currentWeight := _divDownFixed(balances[tokenIndex], sumBalances)
	taxablePercentage := _complementFixed(currentWeight)

	taxableAmount := _mulUpFixed(amountOutWithoutFee, taxablePercentage)
	nonTaxableAmount := new(big.Int).Sub(amountOutWithoutFee, taxableAmount)

	feeOfTaxableAmount := _mulDownFixed(
		taxableAmount,
		new(big.Int).Sub(big.NewInt(ether), swapFeePercentage),
	)

	feeAmount := new(big.Int).Sub(taxableAmount, feeOfTaxableAmount)
	return new(big.Int).Add(nonTaxableAmount, feeOfTaxableAmount), feeAmount, nil
}

// https://github.com/balancer/balancer-v2-monorepo/blob/master/pkg/pool-stable/contracts/StableMath.sol#L399
func _getTokenBalanceGivenInvariantAndAllOtherBalances(
	a *big.Int,
	balances []*big.Int,
	invariant *big.Int,
	tokenIndex int,
) (*big.Int, error) {

	var nTokens = len(balances)
	var nTokensBi = big.NewInt(int64(nTokens))
	var ampTotal = new(big.Int).Mul(a, nTokensBi)
	var sum = balances[0]
	var PD = new(big.Int).Mul(balances[0], nTokensBi) // P_D
	for j := 1; j < nTokens; j++ {
		PD = _divDown(new(big.Int).Mul(new(big.Int).Mul(PD, balances[j]), nTokensBi), invariant)
		sum = new(big.Int).Add(sum, balances[j])
	}
	sum = new(big.Int).Sub(sum, balances[tokenIndex])
	var inv2 = new(big.Int).Mul(invariant, invariant)
	var c = new(big.Int).Mul(
		new(big.Int).Mul(_divUp(inv2, new(big.Int).Mul(ampTotal, PD)), ampPrecision),
		balances[tokenIndex],
	)
	var b = new(big.Int).Add(sum, new(big.Int).Mul(_divDown(invariant, ampTotal), ampPrecision))
	var prevTokenBalance *big.Int
	var tokenBalance = _divUp(new(big.Int).Add(inv2, c), new(big.Int).Add(invariant, b))
	for i := 0; i < 255; i++ {
		prevTokenBalance = tokenBalance
		tokenBalance = _divUp(
			new(big.Int).Add(new(big.Int).Mul(tokenBalance, tokenBalance), c),
			new(big.Int).Sub(new(big.Int).Add(new(big.Int).Mul(tokenBalance, bigIntTwo), b), invariant),
		)
		if tokenBalance.Cmp(prevTokenBalance) > 0 {
			if new(big.Int).Sub(tokenBalance, prevTokenBalance).Cmp(bigIntOne) <= 0 {
				return tokenBalance, nil
			}
		} else if new(big.Int).Sub(prevTokenBalance, tokenBalance).Cmp(bigIntOne) <= 0 {
			return tokenBalance, nil
		}
	}
	return nil, fmt.Errorf("STABLE_GET_BALANCE_DIDNT_CONVERGE")
}

// https://github.com/balancer/balancer-v2-monorepo/blob/master/pkg/pool-stable/contracts/StableMath.sol#L124
func _calcOutGivenIn(
	a *big.Int,
	balances []*big.Int,
	tokenIndexIn int,
	tokenIndexOut int,
	tokenAmountIn *big.Int,
	invariant *big.Int,
) (*big.Int, error) {

	balances[tokenIndexIn] = new(big.Int).Add(balances[tokenIndexIn], tokenAmountIn)
	var finalBalanceOut, err = _getTokenBalanceGivenInvariantAndAllOtherBalances(a, balances, invariant, tokenIndexOut)
	if err != nil {
		return nil, err
	}
	balances[tokenIndexIn] = new(big.Int).Sub(balances[tokenIndexIn], tokenAmountIn)
	return new(big.Int).Sub(new(big.Int).Sub(balances[tokenIndexOut], finalBalanceOut), bigIntOne), nil
}
