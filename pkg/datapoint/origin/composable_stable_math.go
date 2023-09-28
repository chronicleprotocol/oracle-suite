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

func MulDownFixed(a *big.Int, b *big.Int) *big.Int {
	var ret = new(big.Int).Mul(a, b)
	return new(big.Int).Div(ret, bigIntEther)
}

func MulUpFixed(a *big.Int, b *big.Int) *big.Int {
	var ret = new(big.Int).Mul(a, b)
	if ret.Cmp(bigIntZero) == 0 {
		return ret
	}
	return new(big.Int).Add(new(big.Int).Div(new(big.Int).Sub(ret, bigIntOne), bigIntEther), bigIntOne)
}

func DivRounding(a *big.Int, b *big.Int, roundUp bool) *big.Int {
	if roundUp {
		return DivUp(a, b)
	}
	return DivDown(a, b)
}

func DivDown(a *big.Int, b *big.Int) *big.Int {
	return new(big.Int).Div(a, b)
}

func DivUp(a *big.Int, b *big.Int) *big.Int {
	if a.Cmp(bigIntZero) == 0 {
		return bigIntZero
	}
	return new(big.Int).Add(new(big.Int).Div(new(big.Int).Sub(a, bigIntOne), b), bigIntOne)
}

func DivUpFixed(a *big.Int, b *big.Int) *big.Int {
	if a.Cmp(bigIntZero) == 0 {
		return bigIntZero
	}
	aInflated := new(big.Int).Mul(a, bigIntEther)
	return new(big.Int).Add(new(big.Int).Div(new(big.Int).Sub(aInflated, bigIntOne), b), bigIntOne)
}

func DivDownFixed(a *big.Int, b *big.Int) *big.Int {
	if a.Cmp(bigIntZero) == 0 {
		return bigIntZero
	}
	var ret = new(big.Int).Mul(a, bigIntEther)
	return new(big.Int).Div(ret, b)
}

func ComplementFixed(x *big.Int) *big.Int {
	if x.Cmp(bigIntEther) < 0 {
		return new(big.Int).Sub(bigIntEther, x)
	}
	return big.NewInt(0)
}

// https://github.com/balancer/balancer-v2-monorepo/blob/master/pkg/pool-stable/contracts/StableMath.sol#L57
func CalculateInvariant(amp *big.Int, balances []*big.Int, roundUp bool) (*big.Int, error) {
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
			PD = DivRounding(new(big.Int).Mul(new(big.Int).Mul(PD, balances[j]), numTokensBi), invariant, roundUp)
		}
		prevInvariant = invariant
		invariant = DivRounding(
			new(big.Int).Add(
				new(big.Int).Mul(new(big.Int).Mul(numTokensBi, invariant), invariant),
				DivRounding(new(big.Int).Mul(new(big.Int).Mul(ampTotal, sum), PD), ampPrecision, roundUp),
			),
			new(big.Int).Add(
				new(big.Int).Mul(new(big.Int).Add(numTokensBi, bigIntOne), invariant),
				DivRounding(new(big.Int).Mul(new(big.Int).Sub(ampTotal, ampPrecision), PD), ampPrecision, !roundUp),
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

// https://github.com/balancer/balancer-v2-monorepo/blob/master/pkg/pool-stable/contracts/StableMath.sol#L399
func getTokenBalanceGivenInvariantAndAllOtherBalances(
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
		PD = DivDown(new(big.Int).Mul(new(big.Int).Mul(PD, balances[j]), nTokensBi), invariant)
		sum = new(big.Int).Add(sum, balances[j])
	}
	sum = new(big.Int).Sub(sum, balances[tokenIndex])
	var inv2 = new(big.Int).Mul(invariant, invariant)
	var c = new(big.Int).Mul(
		new(big.Int).Mul(DivUp(inv2, new(big.Int).Mul(ampTotal, PD)), ampPrecision),
		balances[tokenIndex],
	)
	var b = new(big.Int).Add(sum, new(big.Int).Mul(DivDown(invariant, ampTotal), ampPrecision))
	var prevTokenBalance *big.Int
	var tokenBalance = DivUp(new(big.Int).Add(inv2, c), new(big.Int).Add(invariant, b))
	for i := 0; i < 255; i++ {
		prevTokenBalance = tokenBalance
		tokenBalance = DivUp(
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
func CalcOutGivenIn(
	a *big.Int,
	balances []*big.Int,
	tokenIndexIn int,
	tokenIndexOut int,
	tokenAmountIn *big.Int,
	invariant *big.Int,
) (*big.Int, error) {

	balances[tokenIndexIn] = new(big.Int).Add(balances[tokenIndexIn], tokenAmountIn)
	var finalBalanceOut, err = getTokenBalanceGivenInvariantAndAllOtherBalances(a, balances, invariant, tokenIndexOut)
	if err != nil {
		return nil, err
	}
	balances[tokenIndexIn] = new(big.Int).Sub(balances[tokenIndexIn], tokenAmountIn)
	return new(big.Int).Sub(new(big.Int).Sub(balances[tokenIndexOut], finalBalanceOut), bigIntOne), nil
}
