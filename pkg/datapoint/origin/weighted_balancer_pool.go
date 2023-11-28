//  Copyright (C) 2021-2023 Chronicle Labs, Inc.
//
//  This program is free software: you can redistribute it and/or modify
//  it under the terms of the GNU Affero General Public License as
//  published by the Free Software Foundation, either version 3 of the
//  License, or (at your option) any later version.
//
//  This program is distributed in the hope that it will be useful,
//  but WITHOUT ANY WARRANTY; without even the implied warranty of
//  MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.  See the
//  GNU Affero General Public License for more details.
//
//  You should have received a copy of the GNU Affero General Public License
//  along with this program.  If not, see <http://www.gnu.org/licenses/>.

package origin

import (
	"context"
	"fmt"
	"math/big"

	"github.com/defiweb/go-eth/rpc"
	"github.com/defiweb/go-eth/types"
	"golang.org/x/exp/maps"

	"github.com/chronicleprotocol/oracle-suite/pkg/datapoint/value"
	"github.com/chronicleprotocol/oracle-suite/pkg/ethereum"
	"github.com/chronicleprotocol/oracle-suite/pkg/util/bn"
	"github.com/chronicleprotocol/oracle-suite/pkg/util/errutil"
)

type WeightedPoolConfig struct {
	Pair    value.Pair
	Address types.Address
}

type WeightedPool struct {
	pair    value.Pair
	address types.Address

	tokens            []types.Address
	balances          []*bn.DecFloatPointNumber
	swapFeePercentage *bn.DecFloatPointNumber
	scalingFactors    []*bn.DecFloatPointNumber
	normalizedWeights []*bn.DecFloatPointNumber
}

type WeightedPools struct {
	client       rpc.RPC
	erc20        *ERC20
	pools        []*WeightedPool
	tokenDetails map[string]ERC20Details
}

func NewWeightedPools(configs []WeightedPoolConfig, client rpc.RPC) (*WeightedPools, error) {
	if client == nil {
		return nil, fmt.Errorf("ethereum client not set")
	}
	var pools []*WeightedPool
	for _, config := range configs {
		pools = append(pools, &WeightedPool{
			pair:    config.Pair,
			address: config.Address,
		})
	}

	erc20, err := NewERC20(client)
	if err != nil {
		return nil, err
	}

	return &WeightedPools{
		client: client,
		erc20:  erc20,
		pools:  pools,
	}, nil
}

func (w *WeightedPools) getPoolTokens(ctx context.Context, blockNumber types.BlockNumber) error {
	var calls []types.Call

	for _, pool := range w.pools {
		// Calls for `getPoolID`
		callData, _ := getPoolID.EncodeArgs()
		calls = append(calls, types.Call{
			To:    &pool.address,
			Input: callData,
		})
		// Calls for `getVault`
		callData, _ = getVault.EncodeArgs()
		calls = append(calls, types.Call{
			To:    &pool.address,
			Input: callData,
		})
	}

	resp, err := ethereum.MultiCall(ctx, w.client, calls, blockNumber)
	if err != nil {
		return err
	}
	calls = make([]types.Call, 0)
	n := len(resp) / len(w.pools)
	for i := range w.pools {
		poolID := types.Bytes(resp[i*n]).PadLeft(32)
		vault := types.MustAddressFromBytes(resp[i*n+1][len(resp[i*n+1])-types.AddressLength:])

		// Calls for `getPoolTokens`
		callData, _ := getPoolTokens.EncodeArgs(poolID.Bytes())
		calls = append(calls, types.Call{
			To:    &vault,
			Input: callData,
		})
	}

	// Get pool tokens from vault by given pool id
	resp, err = ethereum.MultiCall(ctx, w.client, calls, blockNumber)
	if err != nil {
		return err
	}

	tokensMap := make(map[types.Address]struct{})
	for i, pool := range w.pools {
		var tokens []types.Address
		var balances []*big.Int
		if err := getPoolTokens.DecodeValues(resp[i], &tokens, &balances, nil); err != nil {
			return fmt.Errorf("failed decoding pool tokens calls: %s, %w", pool.pair.String(), err)
		}
		for _, address := range tokens {
			tokensMap[address] = struct{}{}
		}
		pool.tokens = tokens
		var decBalances []*bn.DecFloatPointNumber
		for _, balance := range balances {
			decBalances = append(decBalances, bn.DecFloatPoint(balance))
		}
		pool.balances = decBalances
	}

	w.tokenDetails, err = w.erc20.GetSymbolAndDecimals(ctx, maps.Keys(tokensMap))
	if err != nil {
		return nil
	}
	return nil
}

func (w *WeightedPools) getPoolParameters(ctx context.Context, blockNumber types.BlockNumber) error {
	var calls []types.Call

	for _, pool := range w.pools {
		// Calls for `getSwapFeePercentage`
		callData, _ := getSwapFeePercentage.EncodeArgs()
		calls = append(calls, types.Call{
			To:    &pool.address,
			Input: callData,
		})
		// Calls for `getScalingFactors`
		callData, _ = getScalingFactors.EncodeArgs()
		calls = append(calls, types.Call{
			To:    &pool.address,
			Input: callData,
		})
		// Calls for `getNormalizedWeights`
		callData, _ = getNormalizedWeights.EncodeArgs()
		calls = append(calls, types.Call{
			To:    &pool.address,
			Input: callData,
		})
	}

	resp, err := ethereum.MultiCall(ctx, w.client, calls, blockNumber)
	if err != nil {
		return err
	}
	n := len(resp) / len(w.pools)
	for i, pool := range w.pools {
		var swapFeePercentage = new(big.Int).SetBytes(resp[i*n])
		var scalingFactors []*big.Int
		if err := getScalingFactors.DecodeValues(resp[i*n+1], &scalingFactors); err != nil {
			return fmt.Errorf("failed decoding scaling factors calls: %s, %w", pool.pair.String(), err)
		}
		var normalizedWeights []*big.Int
		if err := getNormalizedWeights.DecodeValues(resp[i*n+2], &normalizedWeights); err != nil {
			return fmt.Errorf("failed decoding normal weights calls: %s, %w", pool.pair.String(), err)
		}

		pool.swapFeePercentage = bn.DecFloatPoint(swapFeePercentage)
		pool.scalingFactors = make([]*bn.DecFloatPointNumber, len(scalingFactors))
		for j, factor := range scalingFactors {
			pool.scalingFactors[j] = bn.DecFloatPoint(factor)
		}
		pool.normalizedWeights = make([]*bn.DecFloatPointNumber, len(normalizedWeights))
		for j, weight := range normalizedWeights {
			pool.normalizedWeights[j] = bn.DecFloatPoint(weight)
		}
	}

	return nil
}

func (w *WeightedPools) findPoolByPair(pair value.Pair) *WeightedPool {
	for _, pool := range w.pools {
		if pool.pair == pair {
			return pool
		}
	}
	return nil
}

func (p *WeightedPool) calcAmountOut(tokenIn, tokenOut types.Address, amountIn *bn.DecFloatPointNumber) (
	*bn.DecFloatPointNumber,
	*bn.DecFloatPointNumber,
	error,
) {
	// Make sure that tokenIn and tokenOut are the tokens of pool
	indexIn := -1
	indexOut := -1
	for i, address := range p.tokens {
		if address == tokenIn {
			indexIn = i
		}
		if address == tokenOut {
			indexOut = i
		}
	}
	if indexIn < 0 || indexOut < 0 || indexIn == indexOut {
		return nil, nil, fmt.Errorf("not found tokens in %s: %s, %s",
			p.pair.String(), tokenIn.String(), tokenOut.String())
	}

	amountOut, feeAmount := p._swapGivenIn(indexIn, indexOut, amountIn)
	return bn.DecFloatPoint(amountOut), bn.DecFloatPoint(feeAmount), nil
}

func (p *WeightedPool) _swapGivenIn(indexIn, indexOut int, amountIn *bn.DecFloatPointNumber) (
	*bn.DecFloatPointNumber,
	*bn.DecFloatPointNumber,
) {
	// uint256 scalingFactorTokenIn = _scalingFactor(request.tokenIn);
	scalingFactorTokenIn, _ := p._scalingFactor(indexIn)

	// uint256 scalingFactorTokenOut = _scalingFactor(request.tokenOut);
	scalingFactorTokenOut, _ := p._scalingFactor(indexOut)

	// balanceTokenIn = _upscale(balanceTokenIn, scalingFactorTokenIn);
	balanceTokenInUpscale := p._upscale(p.balances[indexIn], scalingFactorTokenIn)

	// balanceTokenOut = _upscale(balanceTokenOut, scalingFactorTokenOut);
	balanceTokenoutUpscale := p._upscale(p.balances[indexOut], scalingFactorTokenOut)

	// if (request.kind == IVault.SwapKind.GIVEN_IN)

	// Fees are subtracted before scaling, to reduce the complexity of the rounding direction analysis.
	// request.amount = _subtractSwapFeeAmount(request.amount);
	amountAfterFee, feeAmount := p._subtractSwapFeeAmount(amountIn)

	// All token amounts are upscaled.
	amountUpscale := p._upscale(amountAfterFee, scalingFactorTokenIn)

	// uint256 amountOut = _onSwapGivenIn(request, balanceTokenIn, balanceTokenOut);
	amountOut := p._onSwapGivenIn(indexIn, indexOut, amountUpscale, balanceTokenInUpscale, balanceTokenoutUpscale)

	// amountOut tokens are exiting the Pool, so we round down.
	// return _downscaleDown(amountOut, scalingFactorTokenOut);
	return p._downscaleDown(amountOut, scalingFactorTokenOut), feeAmount
}

func (p *WeightedPool) _onSwapGivenIn(
	indexIn, indexOut int,
	amountIn, currentBalanceTokenIn, currentBalanceTokenOut *bn.DecFloatPointNumber,
) *bn.DecFloatPointNumber {
	// return
	//	WeightedMath._calcOutGivenIn(
	//		currentBalanceTokenIn,
	//		_getNormalizedWeight(swapRequest.tokenIn),
	//		currentBalanceTokenOut,
	//		_getNormalizedWeight(swapRequest.tokenOut),
	//		swapRequest.amount
	//	);

	weightIn := errutil.Must(p._getNormalizedWeight(indexIn))
	weightOut := errutil.Must(p._getNormalizedWeight(indexOut))

	return p._calcOutGivenIn(
		currentBalanceTokenIn,
		weightIn,
		currentBalanceTokenOut,
		weightOut,
		amountIn,
	)
}

// _calcOutGivenIn computes how many tokens can be taken out of a pool if `amountIn` are sent, given the
// current balances and weights.
func (p *WeightedPool) _calcOutGivenIn(
	balanceIn, weightIn, balanceOut, weightOut, amountIn *bn.DecFloatPointNumber,
) *bn.DecFloatPointNumber {
	/**********************************************************************************************
	// outGivenIn                                                                                //
	// aO = amountOut                                                                            //
	// bO = balanceOut                                                                           //
	// bI = balanceIn              /      /            bI             \    (wI / wO) \           //
	// aI = amountIn    aO = bO * |  1 - | --------------------------  | ^            |          //
	// wI = weightIn               \      \       ( bI + aI )         /              /           //
	// wO = weightOut                                                                            //
	**********************************************************************************************/

	// Amount out, so we round down overall.

	// The multiplication rounds down, and the subtrahend (power) rounds up (so the base rounds up too).
	// Because bI / (bI + aI) <= 1, the exponent rounds down.

	// Cannot exceed maximum in ratio
	// _require(amountIn <= balanceIn.mulDown(_MAX_IN_RATIO), Errors.MAX_IN_RATIO);

	// uint256 denominator = balanceIn.add(amountIn);
	denominator := balanceIn.Add(amountIn)

	// uint256 base = balanceIn.divUp(denominator);
	base := balanceIn.DivUpFixed(denominator, balancerV2Precision)

	// uint256 exponent = weightIn.divDown(weightOut);
	exponent := weightIn.DivDownFixed(weightOut, balancerV2Precision)

	// uint256 power = base.powUp(exponent);
	power := _powUpFixed(base, exponent, balancerV2Precision)

	// return balanceOut.mulDown(power.complement());
	return balanceOut.MulDownFixed(_complementFixed(power), balancerV2Precision)
}

func (p *WeightedPool) _scalingFactor(index int) (*bn.DecFloatPointNumber, error) {
	if index < 0 || index >= len(p.scalingFactors) {
		return nil, fmt.Errorf("unsupported token")
	}
	return p.scalingFactors[index], nil
}

// Returns the normalized weight of `token`. Weights are fixed point numbers that sum to FixedPoint.ONE.
func (p *WeightedPool) _getNormalizedWeight(index int) (*bn.DecFloatPointNumber, error) {
	if index < 0 || index >= len(p.scalingFactors) {
		return nil, fmt.Errorf("unsupported token")
	}
	return p.normalizedWeights[index], nil
}

// _upscale applies `scalingFactor` to `amount`, resulting in a larger or equal value depending on whether it needed
// scaling or not.
func (p *WeightedPool) _upscale(amount, scalingFactor *bn.DecFloatPointNumber) *bn.DecFloatPointNumber {
	// Upscale rounding wouldn't necessarily always go in the same direction: in a swap for example the balance of
	// token in should be rounded up, and that of token out rounded down. This is the only place where we round in
	// the same direction for all amounts, as the impact of this rounding is expected to be minimal (and there's no
	// rounding error unless `_scalingFactor()` is overriden).
	// return FixedPoint.mulDown(amount, scalingFactor);
	return amount.MulDownFixed(scalingFactor, balancerV2Precision)
}

// _subtractSwapFeeAmount subtracts swap fee amount from `amount`, returning a lower value.
func (p *WeightedPool) _subtractSwapFeeAmount(amount *bn.DecFloatPointNumber) (
	*bn.DecFloatPointNumber,
	*bn.DecFloatPointNumber,
) {
	// This returns amount - fee amount, so we round up (favoring a higher fee amount).
	// uint256 feeAmount = amount.mulUp(getSwapFeePercentage());
	// return amount.sub(feeAmount);
	feeAmount := amount.MulUpFixed(p.swapFeePercentage, balancerV2Precision)
	return amount.Sub(feeAmount), feeAmount
}

// _downscaleDown reverses the `scalingFactor` applied to `amount`, resulting in a smaller or equal value depending on
// whether it needed scaling or not. The result is rounded down.
func (p *WeightedPool) _downscaleDown(amount, scalingFactor *bn.DecFloatPointNumber) *bn.DecFloatPointNumber {
	// return FixedPoint.divDown(amount, scalingFactor);
	return amount.DivDownFixed(scalingFactor, balancerV2Precision)
}
