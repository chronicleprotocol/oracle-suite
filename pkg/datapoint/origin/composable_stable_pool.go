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
)

type ComposableStablePoolConfig struct {
	Pair    value.Pair
	Address types.Address
}

type LastJoinExitData struct {
	lastJoinExitAmplification *bn.DecFloatPointNumber
	lastPostJoinExitInvariant *bn.DecFloatPointNumber
}

type TokenRateCache struct {
	rate     *bn.DecFloatPointNumber
	oldRate  *bn.DecFloatPointNumber
	duration *bn.DecFloatPointNumber
	expires  *bn.DecFloatPointNumber
}

type AmplificationParameter struct {
	value      *bn.DecFloatPointNumber
	isUpdating bool
	precision  *bn.DecFloatPointNumber
}

type Extra struct {
	amplificationParameter              AmplificationParameter
	scalingFactors                      []*bn.DecFloatPointNumber
	lastJoinExit                        LastJoinExitData
	tokensExemptFromYieldProtocolFee    []bool
	tokenRateCaches                     []TokenRateCache
	protocolFeePercentageCacheSwapType  *bn.DecFloatPointNumber
	protocolFeePercentageCacheYieldType *bn.DecFloatPointNumber
}

type ComposableStablePool struct {
	pair    value.Pair
	address types.Address

	tokens            []types.Address
	balances          []*bn.DecFloatPointNumber
	bptIndex          int
	rateProviders     []types.Address
	totalSupply       *bn.DecFloatPointNumber
	swapFeePercentage *bn.DecFloatPointNumber
	extra             Extra
}

type ComposableStablePools struct {
	client       rpc.RPC
	erc20        *ERC20
	pools        []*ComposableStablePool
	tokenDetails map[string]ERC20Details
}

func NewComposableStablePools(configs []ComposableStablePoolConfig, client rpc.RPC) (*ComposableStablePools, error) {
	if client == nil {
		return nil, fmt.Errorf("ethereum client not set")
	}

	var pools []*ComposableStablePool
	for _, config := range configs {
		pools = append(pools, &ComposableStablePool{
			pair:    config.Pair,
			address: config.Address,
		})
	}

	erc20, err := NewERC20(client)
	if err != nil {
		return nil, err
	}

	return &ComposableStablePools{
		client: client,
		erc20:  erc20,
		pools:  pools,
	}, nil
}

func (c *ComposableStablePools) InitializePools(ctx context.Context, blockNumber types.BlockNumber) error {
	err := c.getPoolTokens(ctx, blockNumber)
	if err != nil {
		return err
	}
	err = c.getPoolParameters(ctx, blockNumber)
	if err != nil {
		return err
	}
	err = c.getPoolRateCache(ctx, blockNumber)
	if err != nil {
		return err
	}
	return nil
}

func (c *ComposableStablePools) getPoolTokens(ctx context.Context, blockNumber types.BlockNumber) error {
	var calls []types.Call
	for _, pool := range c.pools {
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

	resp, err := ethereum.MultiCall(ctx, c.client, calls, blockNumber)
	if err != nil {
		return err
	}
	calls = make([]types.Call, 0)
	n := len(resp) / len(c.pools)
	for i := range c.pools {
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
	resp, err = ethereum.MultiCall(ctx, c.client, calls, blockNumber)
	if err != nil {
		return err
	}

	tokensMap := make(map[types.Address]struct{})
	for i, pool := range c.pools {
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

	c.tokenDetails, err = c.erc20.GetSymbolAndDecimals(ctx, maps.Keys(tokensMap))
	if err != nil {
		return nil
	}
	return nil
}

func (c *ComposableStablePools) getPoolParameters(ctx context.Context, blockNumber types.BlockNumber) error { //nolint:funlen
	var calls []types.Call
	for _, pool := range c.pools {
		// Calls for `getBptIndex`
		callData, _ := getBptIndex.EncodeArgs()
		calls = append(calls, types.Call{To: &pool.address, Input: callData})
		// Calls for `getRateProviders`
		callData, _ = getRateProviders.EncodeArgs()
		calls = append(calls, types.Call{To: &pool.address, Input: callData})
		// Calls for `getSwapFeePercentage`
		callData, _ = getSwapFeePercentage.EncodeArgs()
		calls = append(calls, types.Call{To: &pool.address, Input: callData})
		// Calls for `getAmplificationParameter`
		callData, _ = getAmplificationParameter.EncodeArgs()
		calls = append(calls, types.Call{To: &pool.address, Input: callData})
		// Calls for `getScalingFactors`
		callData, _ = getScalingFactors.EncodeArgs()
		calls = append(calls, types.Call{To: &pool.address, Input: callData})
		// Calls for `getLastJoinExitData`
		callData, _ = getLastJoinExitData.EncodeArgs()
		calls = append(calls, types.Call{To: &pool.address, Input: callData})
		// Calls for `getTotalSupply`
		callData, _ = getTotalSupply.EncodeArgs()
		calls = append(calls, types.Call{To: &pool.address, Input: callData})
		// Calls for `getProtocolFeePercentageCache(SWAP)`
		callData, _ = getProtocolFeePercentageCache.EncodeArgs(0)
		calls = append(calls, types.Call{To: &pool.address, Input: callData})
		// Calls for `getProtocolFeePercentageCache(YIELD)`
		callData, _ = getProtocolFeePercentageCache.EncodeArgs(2)
		calls = append(calls, types.Call{To: &pool.address, Input: callData})
		for _, token := range pool.tokens {
			// Calls for `_isTokenExemptFromYieldProtocolFee(token)`
			callData, _ = isTokenExemptFromYieldProtocolFee.EncodeArgs(token)
			calls = append(calls, types.Call{To: &pool.address, Input: callData})
		}
	}

	resp, err := ethereum.MultiCall(ctx, c.client, calls, blockNumber)
	if err != nil {
		return err
	}
	n := len(resp) / len(c.pools)
	for i, pool := range c.pools {
		pool.bptIndex = int(new(big.Int).SetBytes(resp[i*n]).Int64())
		pool.rateProviders = make([]types.Address, 0)
		if err := getRateProviders.DecodeValues(resp[i*n+1], &pool.rateProviders); err != nil {
			return fmt.Errorf("failed decoding rate providers calls: %s, %w", pool.pair.String(), err)
		}
		pool.swapFeePercentage = bn.DecFloatPoint(new(big.Int).SetBytes(resp[i*n+2]))
		var amplificationParameter, amplificationPrecision *big.Int
		var isUpdating bool
		if err := getAmplificationParameter.DecodeValues(resp[i*n+3], &amplificationParameter, &isUpdating, &amplificationPrecision); err != nil {
			return fmt.Errorf("failed decoding amplification parameter calls: %s, %w", pool.pair.String(), err)
		}
		var scalingFactors []*big.Int
		if err := getScalingFactors.DecodeValues(resp[i*n+4], &scalingFactors); err != nil {
			return fmt.Errorf("failed decoding scaling factors calls: %s, %w", pool.pair.String(), err)
		}
		var lastJoinExitAmplification, lastPostJoinExitInvariant *big.Int
		if err := getLastJoinExitData.DecodeValues(resp[i*n+5], &lastJoinExitAmplification, &lastPostJoinExitInvariant); err != nil {
			return fmt.Errorf("failed decoding last join exit calls: %s, %w", pool.pair.String(), err)
		}
		pool.totalSupply = bn.DecFloatPoint(new(big.Int).SetBytes(resp[i*n+6]))
		pool.extra.protocolFeePercentageCacheSwapType = bn.DecFloatPoint(new(big.Int).SetBytes(resp[i*n+7]))
		pool.extra.protocolFeePercentageCacheYieldType = bn.DecFloatPoint(new(big.Int).SetBytes(resp[i*n+8]))
		pool.extra.tokensExemptFromYieldProtocolFee = make([]bool, len(pool.tokens))
		for j := 0; j < len(pool.tokens); j++ {
			var isTokenExempt bool
			if new(big.Int).SetBytes(resp[i*n+9+j]).Cmp(big.NewInt(0)) > 0 {
				isTokenExempt = true
			}
			pool.extra.tokensExemptFromYieldProtocolFee[j] = isTokenExempt
		}
		pool.extra.amplificationParameter.value = bn.DecFloatPoint(amplificationParameter)
		pool.extra.amplificationParameter.isUpdating = isUpdating
		pool.extra.amplificationParameter.precision = bn.DecFloatPoint(amplificationPrecision)
		pool.extra.scalingFactors = make([]*bn.DecFloatPointNumber, len(scalingFactors))
		for j, factor := range scalingFactors {
			pool.extra.scalingFactors[j] = bn.DecFloatPoint(factor)
		}
		pool.extra.lastJoinExit.lastJoinExitAmplification = bn.DecFloatPoint(lastJoinExitAmplification)
		pool.extra.lastJoinExit.lastPostJoinExitInvariant = bn.DecFloatPoint(lastPostJoinExitInvariant)
	}
	return nil
}

func (c *ComposableStablePools) getPoolRateCache(ctx context.Context, blockNumber types.BlockNumber) error {
	var calls []types.Call
	for _, pool := range c.pools {
		if len(pool.tokens) < 1 || len(pool.tokens) != len(pool.rateProviders) {
			return fmt.Errorf("not found proper rate providers in the pool: %s", pool.pair.String())
		}
		for i, token := range pool.tokens {
			if token == pool.address || pool.rateProviders[i] == types.ZeroAddress {
				continue
			}
			// Calls for `getTokenRateCache(token)`
			callData, _ := getTokenRateCache.EncodeArgs(token)
			calls = append(calls, types.Call{
				To:    &pool.address,
				Input: callData,
			})
		}
	}

	resp, err := ethereum.MultiCall(ctx, c.client, calls, blockNumber)
	if err != nil {
		return err
	}
	n := len(resp) / len(c.pools)
	for i, pool := range c.pools {
		for j, token := range pool.tokens {
			if token == pool.address || pool.rateProviders[j] == types.ZeroAddress {
				continue
			}
			var rate, oldRate, duration, expires *big.Int
			if err := getTokenRateCache.DecodeValues(resp[i*n+j], &rate, &oldRate, &duration, &expires); err != nil {
				return fmt.Errorf("failed decoding token rate cache calls: %s, %w", pool.pair.String(), err)
			}
			pool.extra.tokenRateCaches[j] = TokenRateCache{
				rate:     bn.DecFloatPoint(rate),
				oldRate:  bn.DecFloatPoint(oldRate),
				duration: bn.DecFloatPoint(duration),
				expires:  bn.DecFloatPoint(expires),
			}
		}
	}
	return nil
}

func (c *ComposableStablePools) FindPoolByPair(pair value.Pair) *ComposableStablePool {
	for _, pool := range c.pools {
		if pool.pair == pair {
			return pool
		}
	}
	return nil
}

func (p *ComposableStablePool) CalcAmountOut(tokenIn, tokenOut types.Address, amountIn *bn.DecFloatPointNumber) (
	*bn.DecFloatPointNumber,
	*bn.DecFloatPointNumber,
	error,
) {

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

	var amountOut, feeAmount *bn.DecFloatPointNumber
	var err error
	if tokenIn == p.address || tokenOut == p.address {
		amountOut, feeAmount, err = p._swapWithBptGivenIn(indexIn, indexOut, amountIn)
	} else {
		amountOut, feeAmount, err = p._swapGivenIn(indexIn, indexOut, amountIn)
	}
	return bn.DecFloatPoint(amountOut), bn.DecFloatPoint(feeAmount), err
}

// _onRegularSwap implements same functionality with the following url:
// https://github.com/balancer/balancer-v2-monorepo/blob/master/pkg/pool-stable/contracts/ComposableStablePool.sol#L283
func (p *ComposableStablePool) _onRegularSwap(
	amountIn *bn.DecFloatPointNumber,
	registeredBalances []*bn.DecFloatPointNumber,
	registeredIndexIn,
	registeredIndexOut int,
) (*bn.DecFloatPointNumber, error) {
	// Adjust indices and balances for BPT token
	// uint256[] memory balances = _dropBptItem(registeredBalances);
	// uint256 indexIn = _skipBptIndex(indexIn);
	// uint256 indexOut = _skipBptIndex(indexOut);

	droppedBalances := p._dropBptItem(registeredBalances)
	indexIn := p._skipBptIndex(registeredIndexIn)
	indexOut := p._skipBptIndex(registeredIndexOut)

	// (uint256 currentAmp, ) = _getAmplificationParameter();
	// uint256 invariant = StableMath._calculateInvariant(currentAmp, balances);
	currentAmp := p.extra.amplificationParameter.value
	invariant, err := _calculateInvariant(currentAmp, droppedBalances)
	if err != nil {
		return nil, err
	}

	// StableMath._calcOutGivenIn(currentAmp, balances, indexIn, indexOut, amountGiven, invariant);
	return _calcOutGivenIn(currentAmp, droppedBalances, indexIn, indexOut, amountIn, invariant)
}

// _onSwapGivenIn implements same functionality with the following url:
// https://github.com/balancer/balancer-v2-monorepo/blob/master/pkg/pool-stable/contracts/ComposableStablePool.sol#L242
func (p *ComposableStablePool) _onSwapGivenIn(
	amountIn *bn.DecFloatPointNumber,
	registeredBalances []*bn.DecFloatPointNumber,
	indexIn,
	indexOut int,
) (*bn.DecFloatPointNumber, error) {

	return p._onRegularSwap(amountIn, registeredBalances, indexIn, indexOut)
}

// Perform a swap involving the BPT token, equivalent to a single-token join or exit.
// As with the standard joins and swaps, we first pay any protocol fees pending from swaps that occurred since the previous join or exit,
// then perform the operation (joinSwap or exitSwap), and finally store the "post operation" invariant and amp,
// which establishes the new basis for protocol fees.
//
// At this point, the scaling factors (including rates) have been computed by the base class,
// but not yet applied to the balances.
//
// Reference: https://github.com/balancer/balancer-v2-monorepo/blob/master/pkg/pool-stable/contracts/ComposableStablePool.sol#L314
func (p *ComposableStablePool) _swapWithBptGivenIn(indexIn, indexOut int, amountIn *bn.DecFloatPointNumber) (
	*bn.DecFloatPointNumber,
	*bn.DecFloatPointNumber,
	error,
) {

	var amountCalculated, feeAmount *bn.DecFloatPointNumber

	// bool isGivenIn = swapRequest.kind == IVault.SwapKind.GIVEN_IN;
	// _upscaleArray(registeredBalances, scalingFactors);
	// swapRequest.amount = _upscale(
	//	swapRequest.amount,
	//	scalingFactors[isGivenIn ? indexIn : indexOut]
	balancesUpscaled := p._upscaleArray(p.balances, p.extra.scalingFactors)
	tokenAmountInScaled := p._upscale(amountIn, p.extra.scalingFactors[indexIn])

	// (
	//	uint256 preJoinExitSupply,
	//	uint256[] memory balances,
	//	uint256 currentAmp,
	//	uint256 preJoinExitInvariant
	// ) = _beforeJoinExit(registeredBalances);
	preJoinExitSupply, balances, currentAmp, preJoinExitInvariant, err := p._beforeJoinExit(balancesUpscaled)
	if err != nil {
		return nil, nil, err
	}
	if indexOut == p.bptIndex {
		// _doJoinSwap(
		//	isGivenIn,
		//	swapRequest.amount,
		//	balances,
		//	_skipBptIndex(registeredIndexIn),
		//	currentAmp,
		//	preJoinExitSupply,
		//	preJoinExitInvariant
		// )
		amountCalculated, _, feeAmount, err = p._doJoinSwap(
			true,
			tokenAmountInScaled,
			balances,
			p._skipBptIndex(indexIn),
			currentAmp,
			preJoinExitSupply,
			preJoinExitInvariant,
		)
		if err != nil {
			return nil, nil, err
		}
	} else {
		// _doExitSwap(
		//	isGivenIn,
		//	swapRequest.amount,
		//	balances,
		//	_skipBptIndex(registeredIndexOut),
		//	currentAmp,
		//	preJoinExitSupply,
		//	preJoinExitInvariant
		// )
		amountCalculated, _, feeAmount, err = p._doExitSwap(
			true,
			tokenAmountInScaled,
			balances,
			p._skipBptIndex(indexOut),
			currentAmp,
			preJoinExitSupply,
			preJoinExitInvariant,
		)
		if err != nil {
			return nil, nil, err
		}
	}
	if amountCalculated == nil {
		return nil, nil, fmt.Errorf("INVALID_AMOUNT_OUT_CALCULATED")
	}
	// _downscaleDown(amountCalculated, scalingFactors[registeredIndexOut]) // Amount out, round down
	return amountCalculated.DivDownFixed(p.extra.scalingFactors[indexOut], balancerV2Precision), feeAmount, nil
}

// Since this is an exit, we know the tokenIn is BPT. Since it is GivenIn, we know the BPT amount, and must calculate the token amount out.
// We are moving BPT out of circulation and into the Vault, which decreases the virtual supply.
//
// Reference: https://github.com/balancer/balancer-v2-monorepo/blob/master/pkg/pool-stable/contracts/ComposableStablePool.sol#L504
func (p *ComposableStablePool) _exitSwapExactBptInForTokenOut(
	bptAmount *bn.DecFloatPointNumber,
	balances []*bn.DecFloatPointNumber,
	indexOut int,
	currentAmp *bn.DecFloatPointNumber,
	actualSupply *bn.DecFloatPointNumber,
	preJoinExitInvariant *bn.DecFloatPointNumber,
) (*bn.DecFloatPointNumber, *bn.DecFloatPointNumber, *bn.DecFloatPointNumber, error) {

	amountOut, feeAmount, err := _calcTokenOutGivenExactBptIn(
		currentAmp, balances, indexOut, bptAmount, actualSupply, preJoinExitInvariant, p.swapFeePercentage)
	if err != nil {
		return nil, nil, nil, err
	}

	balances[indexOut] = balances[indexOut].Sub(amountOut)
	postJoinExitSupply := actualSupply.Sub(bptAmount)

	return amountOut, postJoinExitSupply, feeAmount, nil
}

// This mutates `balances` so that they become the post-joinswap balances.
// The StableMath interfaces are different depending on the swap direction, so we forward to the appropriate low-level join function.
//
// Reference: https://github.com/balancer/balancer-v2-monorepo/blob/master/pkg/pool-stable/contracts/ComposableStablePool.sol#L375
func (p *ComposableStablePool) _doJoinSwap(
	isGivenIn bool,
	amount *bn.DecFloatPointNumber,
	balances []*bn.DecFloatPointNumber,
	indexIn int,
	currentAmp *bn.DecFloatPointNumber,
	actualSupply *bn.DecFloatPointNumber,
	preJoinExitInvariant *bn.DecFloatPointNumber,
) (*bn.DecFloatPointNumber, *bn.DecFloatPointNumber, *bn.DecFloatPointNumber, error) {

	if isGivenIn {
		return p._joinSwapExactTokenInForBptOut(amount, balances, indexIn, currentAmp, actualSupply, preJoinExitInvariant)
	}
	// Currently ignore givenOut case
	return nil, nil, nil, nil
}

// This mutates balances so that they become the post-exitswap balances.
// The StableMath interfaces are different depending on the swap direction, so we forward to the appropriate low-level exit function.
//
// Reference: https://github.com/balancer/balancer-v2-monorepo/blob/master/pkg/pool-stable/contracts/ComposableStablePool.sol#L470
func (p *ComposableStablePool) _doExitSwap(
	isGivenIn bool,
	amount *bn.DecFloatPointNumber,
	balances []*bn.DecFloatPointNumber,
	indexOut int,
	currentAmp *bn.DecFloatPointNumber,
	actualSupply *bn.DecFloatPointNumber,
	preJoinExitInvariant *bn.DecFloatPointNumber,
) (*bn.DecFloatPointNumber, *bn.DecFloatPointNumber, *bn.DecFloatPointNumber, error) {

	if isGivenIn {
		return p._exitSwapExactBptInForTokenOut(amount, balances, indexOut, currentAmp, actualSupply, preJoinExitInvariant)
	}
	// Currently ignore givenOut case
	return nil, nil, nil, nil
}

// Since this is a join, we know the tokenOut is BPT.
// Since it is GivenIn, we know the tokenIn amount, and must calculate the BPT amount out.
// We are moving preminted BPT out of the Vault, which increases the virtual supply.
//
// Reference: https://github.com/balancer/balancer-v2-monorepo/blob/master/pkg/pool-stable/contracts/ComposableStablePool.sol#L409
func (p *ComposableStablePool) _joinSwapExactTokenInForBptOut(
	amountIn *bn.DecFloatPointNumber,
	balances []*bn.DecFloatPointNumber,
	indexIn int,
	currentAmp *bn.DecFloatPointNumber,
	actualSupply *bn.DecFloatPointNumber,
	preJoinExitInvariant *bn.DecFloatPointNumber,
) (*bn.DecFloatPointNumber, *bn.DecFloatPointNumber, *bn.DecFloatPointNumber, error) {
	// The StableMath function was created with joins in mind, so it expects a full amounts array.
	// We create an empty one and only set the amount for the token involved.
	amountsIn := make([]*bn.DecFloatPointNumber, len(balances))
	for i := range amountsIn {
		amountsIn[i] = bn.DecFloatPoint(0)
	}
	amountsIn[indexIn] = amountIn
	bptOut, feeAmountIn, err := _calcBptOutGivenExactTokensIn(
		currentAmp, balances, amountsIn, actualSupply, preJoinExitInvariant, p.swapFeePercentage)
	if err != nil {
		return nil, nil, nil, err
	}
	balances[indexIn] = balances[indexIn].Add(amountIn)
	postJoinExitSupply := actualSupply.Add(bptOut)

	return bptOut, postJoinExitSupply, feeAmountIn, nil
}

// Pay any due protocol fees and calculate values necessary for performing the join/exit.
//
// Reference: https://github.com/balancer/balancer-v2-monorepo/blob/master/pkg/pool-stable/contracts/ComposableStablePool.sol#L701
func (p *ComposableStablePool) _beforeJoinExit(registeredBalances []*bn.DecFloatPointNumber) (
	*bn.DecFloatPointNumber,
	[]*bn.DecFloatPointNumber,
	*bn.DecFloatPointNumber,
	*bn.DecFloatPointNumber,
	error,
) {

	preJoinExitSupply, balances, oldAmpPreJoinExitInvariant, err := p._payProtocolFeesBeforeJoinExit(registeredBalances)
	if err != nil {
		return nil, nil, nil, nil, err
	}
	currentAmp := p.extra.amplificationParameter.value

	// If the amplification factor is the same as it was during the last join/exit then we can reuse the
	// value calculated using the "old" amplification factor. If not, then we have to calculate this now.
	var (
		preJoinExitInvariant *bn.DecFloatPointNumber
	)
	if currentAmp.Cmp(p.extra.lastJoinExit.lastJoinExitAmplification) == 0 {
		preJoinExitInvariant = oldAmpPreJoinExitInvariant
	} else {
		preJoinExitInvariant, err = _calculateInvariant(currentAmp, balances)
	}
	if err != nil {
		return nil, nil, nil, nil, err
	}

	return preJoinExitSupply, balances, currentAmp, preJoinExitInvariant, nil
}

// Calculates due protocol fees originating from accumulated swap fees and yield of non-exempt tokens,
// pays them by minting BPT, and returns the actual supply and current balances.
//
// We also return the current invariant computed using the amplification factor at the last join or exit,
// which can be useful to skip computations in scenarios where the amplification factor is not changing.
//
// Reference: https://github.com/balancer/balancer-v2-monorepo/blob/master/pkg/pool-stable/contracts/ComposableStablePoolProtocolFees.sol#L64
func (p *ComposableStablePool) _payProtocolFeesBeforeJoinExit(
	registeredBalances []*bn.DecFloatPointNumber,
) (*bn.DecFloatPointNumber, []*bn.DecFloatPointNumber, *bn.DecFloatPointNumber, error) {

	virtualSupply, droppedBalances := p._dropBptItemFromBalances(registeredBalances)
	// First, we'll compute what percentage of the Pool the protocol should own due to charging protocol fees on
	// swap fees and yield.
	expectedProtocolOwnershipPercentage, currentInvariantWithLastJoinExitAmp, err := p._getProtocolPoolOwnershipPercentage(droppedBalances)
	if err != nil {
		return nil, nil, nil, err
	}
	// Now that we know what percentage of the Pool's current value the protocol should own, we can compute how
	// much BPT we need to mint to get to this state. Since we're going to mint BPT for the protocol, the value
	// of each BPT is going to be reduced as all LPs get diluted.
	protocolFeeAmount := p._bptForPoolOwnershipPercentage(virtualSupply, expectedProtocolOwnershipPercentage)

	// _payProtocolFees(protocolFeeAmount)

	// We pay fees before a join or exit to ensure the pool is debt-free. This increases the virtual supply (making
	// it match the actual supply).
	//
	// For this addition to overflow, `totalSupply` would also have already overflowed.
	return virtualSupply.Add(protocolFeeAmount), droppedBalances, currentInvariantWithLastJoinExitAmp, nil
}

// Reference: https://github.com/balancer/balancer-v2-monorepo/blob/master/pkg/pool-stable/contracts/ComposableStablePoolProtocolFees.sol#L102
func (p *ComposableStablePool) _getProtocolPoolOwnershipPercentage(balances []*bn.DecFloatPointNumber) (
	*bn.DecFloatPointNumber,
	*bn.DecFloatPointNumber,
	error,
) {
	// We compute three invariants, adjusting the balances of tokens that have rate providers by undoing the current
	// rate adjustment and then applying the old rate. This is equivalent to multiplying by old rate / current rate.
	//
	// In all cases we compute invariants with the last join-exit amplification factor, so that changes to the
	// amplification are not translated into changes to the invariant. Since amplification factor changes are both
	// infrequent and slow, they should have little effect on the pool balances, making this a very good
	// approximation.
	//
	// With this technique we obtain an invariant that does not include yield at all, meaning any growth will be due
	// exclusively to swap fees. We call this the 'swap fee growth invariant'.
	// A second invariant will exclude the yield of exempt tokens, and therefore include both swap fees and
	// non-exempt yield. This is called the 'non exempt growth invariant'.
	// Finally, a third invariant includes the yield of all tokens by using only the current rates. We call this the
	// 'total growth invariant', since it includes both swap fee growth, non-exempt yield growth and exempt yield
	// growth. If the last join-exit amplification equals the current one, this invariant equals the current
	// invariant.

	swapFeeGrowthInvariant, totalNonExemptGrowthInvariant, totalGrowthInvariant, err := p._getGrowthInvariants(balances)
	if err != nil {
		return nil, nil, err
	}
	// By comparing the invariant increase attributable to each source of growth to the total growth invariant,
	// we can calculate how much of the current Pool value originates from that source, and then apply the
	// corresponding protocol fee percentage to that amount.

	// We have two sources of growth: swap fees, and non-exempt yield. As we illustrate graphically below:
	//
	// growth due to swap fees        = (swap fee growth invariant - last post join-exit invariant)
	// growth due to non-exempt yield = (non-exempt growth invariant - swap fee growth invariant)
	//
	// These can be converted to additive percentages by normalizing against the total growth invariant value:
	// growth due to swap fees / total growth invariant = % pool ownership due from swap fees
	// growth due to non-exempt yield / total growth invariant = % pool ownership due from non-exempt yield
	//
	//   ┌───────────────────────┐ ──┐
	//   │  exempt yield         │   │  total growth invariant
	//   ├───────────────────────┤   │ ──┐
	//   │  non-exempt yield     │   │   │  non-exempt growth invariant
	//   ├───────────────────────┤   │   │ ──┐
	//   │  swap fees            │   │   │   │  swap fee growth invariant
	//   ├───────────────────────┤   │   │   │ ──┐
	//   │   original value      │   │   │   │   │  last post join-exit invariant
	//   └───────────────────────┘ ──┘ ──┘ ──┘ ──┘
	//
	// Each invariant should be larger than its precedessor. In case any rounding error results in them being
	// smaller, we adjust the subtraction to equal 0.

	// Note: in the unexpected scenario where the rates of the tokens shrink over time instead of growing (i.e. if
	// the yield is negative), the non-exempt growth invariant might actually be *smaller* than the swap fee growth
	// invariant, and the total growth invariant might be *smaller* than the non-exempt growth invariant. Depending
	// on the order in which swaps, joins/exits and rate changes happen, as well as their relative magnitudes, it is
	// possible for the Pool to either pay more or less protocol fees than it should.
	// Due to the complexity that handling all of these cases would introduce, this behavior is considered out of
	// scope, and is expected to be handled on a case-by-case basis if the token rates were to ever decrease (which
	// would also mean that the Pool value has dropped).

	// Calculate the delta for swap fee growth invariant
	swapFeeGrowthInvariantDelta := swapFeeGrowthInvariant.Sub(p.extra.lastJoinExit.lastPostJoinExitInvariant)
	if swapFeeGrowthInvariantDelta.Cmp(bnZero) < 0 {
		swapFeeGrowthInvariantDelta = bn.DecFloatPoint(0)
	}

	// Calculate the delta for non-exempt yield growth invariant
	nonExemptYieldGrowthInvariantDelta := totalNonExemptGrowthInvariant.Sub(swapFeeGrowthInvariant)
	if nonExemptYieldGrowthInvariantDelta.Cmp(bnZero) < 0 {
		nonExemptYieldGrowthInvariantDelta = bn.DecFloatPoint(0)
	}

	// We can now derive what percentage of the Pool's total value each invariant delta represents by dividing by
	// the total growth invariant. These values, multiplied by the protocol fee percentage for each growth type,
	// represent the percentage of Pool ownership the protocol should have due to each source.

	// swapFeeGrowthInvariantDelta/totalGrowthInvariant*getProtocolFeePercentageCache
	protocolSwapFeePercentage :=
		swapFeeGrowthInvariantDelta.DivDownFixed(totalGrowthInvariant, balancerV2Precision).MulDownFixed(
			p.extra.protocolFeePercentageCacheSwapType, balancerV2Precision)

	protocolYieldPercentage :=
		nonExemptYieldGrowthInvariantDelta.DivDownFixed(totalGrowthInvariant, balancerV2Precision).MulDownFixed(
			p.extra.protocolFeePercentageCacheYieldType, balancerV2Precision)

	// These percentages can then be simply added to compute the total protocol Pool ownership percentage.
	// This is naturally bounded above by FixedPoint.ONE so this addition cannot overflow.

	// Calculate the total protocol ComposableStablePool ownership percentage
	protocolPoolOwnershipPercentage := protocolSwapFeePercentage.Add(protocolYieldPercentage)
	return protocolPoolOwnershipPercentage, totalGrowthInvariant, nil
}

// Reference: https://github.com/balancer/balancer-v2-monorepo/blob/master/pkg/pool-stable/contracts/ComposableStablePoolProtocolFees.sol#L189
func (p *ComposableStablePool) _getGrowthInvariants(balances []*bn.DecFloatPointNumber) (
	*bn.DecFloatPointNumber,
	*bn.DecFloatPointNumber,
	*bn.DecFloatPointNumber,
	error,
) {
	// We always calculate the swap fee growth invariant, since we cannot easily know whether swap fees have
	// accumulated or not.

	var (
		swapFeeGrowthInvariant        *bn.DecFloatPointNumber
		totalNonExemptGrowthInvariant *bn.DecFloatPointNumber
		totalGrowthInvariant          *bn.DecFloatPointNumber
		err                           error
	)

	// This invariant result is calc by _divDown (round down)
	// https://github.com/balancer/balancer-v2-monorepo/blob/b46023f7c5deefaf58a0a42559a36df420e1639f/pkg/pool-stable/contracts/StableMath.sol#L96
	swapFeeGrowthInvariant, err = _calculateInvariant(
		p.extra.lastJoinExit.lastJoinExitAmplification,
		p._getAdjustedBalances(balances, true))
	if err != nil {
		return nil, nil, nil, err
	}

	// For the other invariants, we can potentially skip some work. In the edge cases where none or all of the
	// tokens are exempt from yield, there's one fewer invariant to compute.
	switch {
	case p._areNoTokensExempt():
		// If there are no tokens with fee-exempt yield, then the total non-exempt growth will equal the total
		// growth: all yield growth is non-exempt. There's also no point in adjusting balances, since we
		// already know none are exempt.
		totalNonExemptGrowthInvariant, err = _calculateInvariant(p.extra.lastJoinExit.lastJoinExitAmplification, balances)
		if err != nil {
			return nil, nil, nil, err
		}

		totalGrowthInvariant = totalNonExemptGrowthInvariant
	case p._areAllTokensExempt():
		// If no tokens are charged fees on yield, then the non-exempt growth is equal to the swap fee growth - no
		// yield fees will be collected.
		totalNonExemptGrowthInvariant = swapFeeGrowthInvariant
		totalGrowthInvariant, err = _calculateInvariant(p.extra.lastJoinExit.lastJoinExitAmplification, balances)
		if err != nil {
			return nil, nil, nil, err
		}
	default:
		// In the general case, we need to calculate two invariants: one with some adjusted balances, and one with
		// the current balances.

		totalNonExemptGrowthInvariant, err = _calculateInvariant(
			p.extra.lastJoinExit.lastJoinExitAmplification,
			p._getAdjustedBalances(balances, false), // Only adjust non-exempt balances
		)
		if err != nil {
			return nil, nil, nil, err
		}

		totalGrowthInvariant, err = _calculateInvariant(
			p.extra.lastJoinExit.lastJoinExitAmplification,
			balances)
		if err != nil {
			return nil, nil, nil, err
		}
	}
	return swapFeeGrowthInvariant, totalNonExemptGrowthInvariant, totalGrowthInvariant, nil
}

// Same as `_dropBptItem`, except the virtual supply is also returned,
// and `balances` is assumed to be the current Pool balances (including BPT).
//
// Reference: https://github.com/balancer/balancer-v2-monorepo/blob/master/pkg/pool-stable/contracts/ComposableStablePoolStorage.sol#L259
func (p *ComposableStablePool) _dropBptItemFromBalances(balances []*bn.DecFloatPointNumber) (*bn.DecFloatPointNumber, []*bn.DecFloatPointNumber) {
	return p._getVirtualSupply(balances[p.bptIndex]), p._dropBptItem(balances)
}

// Returns the number of tokens in circulation.
// WARNING: in the vast majority of cases this is not a useful value, since it does not include the debt the Pool
// accrued in the form of unminted BPT for the ProtocolFeesCollector. Look into `getActualSupply()` and how that's
// different.
//
// In other pools, this would be the same as `totalSupply`, but since this pool pre-mints BPT and holds it in the
// Vault as a token, we need to subtract the Vault's balance to get the total "circulating supply". Both the
// totalSupply and Vault balance can change. If users join or exit using swaps, some of the preminted BPT are
// exchanged, so the Vault's balance increases after joins and decreases after exits. If users call the regular
// joins/exit functions, the totalSupply can change as BPT are minted for joins or burned for exits.
//
// Reference: https://github.com/balancer/balancer-v2-monorepo/blob/master/pkg/pool-stable/contracts/ComposableStablePoolStorage.sol#L386
func (p *ComposableStablePool) _getVirtualSupply(bptBalance *bn.DecFloatPointNumber) *bn.DecFloatPointNumber {
	// The initial amount of BPT pre-minted is _PREMINTED_TOKEN_BALANCE, and it goes entirely to the pool balance in
	// the vault. So the virtualSupply (the amount of BPT supply in circulation) is defined as:
	// virtualSupply = totalSupply() - _balances[_bptIndex]
	return p.totalSupply.Sub(bptBalance)
}

// Return true if the token at this index has a rate provider
// Reference: https://github.com/balancer/balancer-v2-monorepo/blob/master/pkg/pool-stable/contracts/ComposableStablePoolStorage.sol#L337
func (p *ComposableStablePool) _hasRateProvider(tokenIndex int) bool {
	return p.rateProviders[tokenIndex] != types.ZeroAddress
}

// Returns whether the token is exempt from protocol fees on the yield.
// If the BPT token is passed in (which doesn't make much sense, but shouldn't fail,
// since it is a valid pool token), the corresponding flag will be false.
//
// Reference: https://github.com/balancer/balancer-v2-monorepo/blob/master/pkg/pool-stable/contracts/ComposableStablePoolStorage.sol#L362
func (p *ComposableStablePool) _isTokenExemptFromYieldProtocolFee(tokenIndex int) bool {
	return p.extra.tokensExemptFromYieldProtocolFee[tokenIndex]
}

// Return true if no tokens are exempt from yield fees.
// Reference: https://github.com/balancer/balancer-v2-monorepo/blob/master/pkg/pool-stable/contracts/ComposableStablePoolStorage.sol#L351
func (p *ComposableStablePool) _areNoTokensExempt() bool {
	for _, exempt := range p.extra.tokensExemptFromYieldProtocolFee {
		if exempt {
			return false
		}
	}
	return true
}

// Return true if all tokens are exempt from yield fees.
// Reference: https://github.com/balancer/balancer-v2-monorepo/blob/master/pkg/pool-stable/contracts/ComposableStablePoolStorage.sol#L344
func (p *ComposableStablePool) _areAllTokensExempt() bool {
	for _, exempt := range p.extra.tokensExemptFromYieldProtocolFee {
		if !exempt {
			return false
		}
	}
	return true
}

// Apply the token ratios to a set of balances, optionally adjusting for exempt yield tokens.
// The `balances` array is assumed to not include BPT to ensure that token indices align.
// Reference: https://github.com/balancer/balancer-v2-monorepo/blob/master/pkg/pool-stable/contracts/ComposableStablePoolRates.sol#L222
func (p *ComposableStablePool) _getAdjustedBalances(balances []*bn.DecFloatPointNumber, ignoreExemptFlags bool) []*bn.DecFloatPointNumber {
	totalTokensWithoutBpt := len(balances)
	adjustedBalances := make([]*bn.DecFloatPointNumber, totalTokensWithoutBpt)

	for i := 0; i < totalTokensWithoutBpt; i++ {
		skipBptIndex := i
		if i >= p.bptIndex {
			skipBptIndex++
		}

		if p._isTokenExemptFromYieldProtocolFee(skipBptIndex) || (ignoreExemptFlags && p._hasRateProvider(skipBptIndex)) {
			adjustedBalances[i] = p._adjustedBalance(balances[i], &p.extra.tokenRateCaches[skipBptIndex])
		} else {
			adjustedBalances[i] = balances[i]
		}
	}

	return adjustedBalances
}

// Compute balance * oldRate/currentRate, doing division last to minimize rounding error.
// Reference: https://github.com/balancer/balancer-v2-monorepo/blob/master/pkg/pool-stable/contracts/ComposableStablePoolRates.sol#L242
func (p *ComposableStablePool) _adjustedBalance(balance *bn.DecFloatPointNumber, cache *TokenRateCache) *bn.DecFloatPointNumber {
	return balance.Mul(cache.oldRate).DivDown(cache.rate)
}

// Remove the item at `_bptIndex` from an arbitrary array (e.g., amountsIn).
// Reference: https://github.com/balancer/balancer-v2-monorepo/blob/master/pkg/pool-stable/contracts/ComposableStablePoolStorage.sol#L246
func (p *ComposableStablePool) _dropBptItem(amounts []*bn.DecFloatPointNumber) []*bn.DecFloatPointNumber {
	amountsWithoutBpt := make([]*bn.DecFloatPointNumber, len(amounts)-1)
	bptIndex := p.bptIndex

	for i := 0; i < len(amountsWithoutBpt); i++ {
		if i < bptIndex {
			amountsWithoutBpt[i] = amounts[i]
		} else {
			amountsWithoutBpt[i] = amounts[i+1]
		}
	}
	return amountsWithoutBpt
}

// Calculates the amount of BPT necessary to give ownership of a given percentage of the Pool to an external
// third party. In the case of protocol fees, this is the DAO, but could also be a pool manager, etc.
// Note that this function reverts if `poolPercentage` >= 100%, it's expected that the caller will enforce this.
// @param totalSupply - The total supply of the pool prior to minting BPT.
// @param poolOwnershipPercentage - The desired ownership percentage of the pool to have as a result of minting BPT.
// @return bptAmount - The amount of BPT to mint such that it is `poolPercentage` of the resultant total supply.
//
// Reference: https://github.com/balancer/balancer-v2-monorepo/blob/master/pkg/pool-utils/contracts/external-fees/ExternalFees.sol#L31
func (p *ComposableStablePool) _bptForPoolOwnershipPercentage(totalSupply, poolOwnershipPercentage *bn.DecFloatPointNumber) *bn.DecFloatPointNumber {
	// If we mint some amount `bptAmount` of BPT then the percentage ownership of the pool this grants is given by:
	// `poolOwnershipPercentage = bptAmount / (totalSupply + bptAmount)`.
	// Solving for `bptAmount`, we arrive at:
	// `bptAmount = totalSupply * poolOwnershipPercentage / (1 - poolOwnershipPercentage)`.
	return totalSupply.Mul(poolOwnershipPercentage).DivPrec(_complementFixed(poolOwnershipPercentage), 0)
}

// Convert from an index into an array including BPT (the Vault's registered token list), to an index
// into an array excluding BPT (usually from user input, such as amountsIn/Out).
// `index` must not be the BPT token index itself.
//
// Reference: https://github.com/balancer/balancer-v2-monorepo/blob/master/pkg/pool-stable/contracts/ComposableStablePoolStorage.sol#L235
func (p *ComposableStablePool) _skipBptIndex(index int) int {
	if index < p.bptIndex {
		return index
	}
	return index - 1
}

// Override this hook called by the base class `onSwap`, to check whether we are doing a regular swap,
// or a swap involving BPT, which is equivalent to a single token join or exit. Since one of the Pool's
// tokens is the preminted BPT, we need to handle swaps where BPT is involved separately.
//
// At this point, the balances are unscaled. The indices are coming from the Vault, so they are indices into
// the array of registered tokens (including BPT).
//
// If this is a swap involving BPT, call `_swapWithBpt`, which computes the amountOut using the swapFeePercentage
// and charges protocol fees, in the same manner as single token join/exits. Otherwise, perform the default
// processing for a regular swap.
//
// Reference: https://github.com/balancer/balancer-v2-monorepo/blob/master/pkg/pool-utils/contracts/BaseGeneralPool.sol#L49
func (p *ComposableStablePool) _swapGivenIn(indexIn, indexOut int, amountIn *bn.DecFloatPointNumber) (
	*bn.DecFloatPointNumber,
	*bn.DecFloatPointNumber,
	error,
) {
	// Fees are subtracted before scaling, to reduce the complexity of the rounding direction analysis.
	// swapRequest.amount = _subtractSwapFeeAmount(swapRequest.amount);
	amountAfterFee, feeAmount := p._subtractSwapFeeAmount(amountIn, p.swapFeePercentage)

	// _upscaleArray(balances, scalingFactors);
	// swapRequest.amount = _upscale(swapRequest.amount, scalingFactors[indexIn]);
	upscaledBalances := p._upscaleArray(p.balances, p.extra.scalingFactors)
	amountUpScale := p._upscale(amountAfterFee, p.extra.scalingFactors[indexIn])

	// uint256 amountOut = _onSwapGivenIn(swapRequest, balances, indexIn, indexOut);
	amountOut, err := p._onSwapGivenIn(amountUpScale, upscaledBalances, indexIn, indexOut)
	if err != nil {
		return nil, nil, err
	}

	return amountOut.DivDownFixed(p.extra.scalingFactors[indexOut], balancerV2Precision), feeAmount, nil
}

// Subtracts swap fee amount from `amount`, returning a lower value.
// Reference: https://github.com/balancer/balancer-v2-monorepo/blob/master/pkg/pool-utils/contracts/BasePool.sol#L603
func (p *ComposableStablePool) _subtractSwapFeeAmount(amount, swapFeePercentage *bn.DecFloatPointNumber) (
	*bn.DecFloatPointNumber,
	*bn.DecFloatPointNumber,
) {

	feeAmount := amount.MulUpFixed(swapFeePercentage, balancerV2Precision)
	return amount.Sub(feeAmount), feeAmount
}

// Same as `_upscale`, but for an entire array. This function does not return anything, but instead *mutates*
// the `amounts` array.
// Reference: https://github.com/balancer/balancer-v2-monorepo/blob/master/pkg/solidity-utils/contracts/helpers/ScalingHelpers.sol#L64C1-L64C1
func (p *ComposableStablePool) _upscaleArray(amounts, scalingFactors []*bn.DecFloatPointNumber) []*bn.DecFloatPointNumber {
	result := make([]*bn.DecFloatPointNumber, len(amounts))
	for i, amount := range amounts {
		result[i] = amount.MulUpFixed(scalingFactors[i], balancerV2Precision)
	}
	return result
}

// Applies `scalingFactor` to `amount`, resulting in a larger or equal value depending on whether it needed
// scaling or not.
// https://github.com/balancer/balancer-v2-monorepo/blob/master/pkg/solidity-utils/contracts/helpers/ScalingHelpers.sol#L32
func (p *ComposableStablePool) _upscale(amount, scalingFactor *bn.DecFloatPointNumber) *bn.DecFloatPointNumber {
	// Upscale rounding wouldn't necessarily always go in the same direction: in a swap for example the balance of
	// token in should be rounded up, and that of token out rounded down. This is the only place where we round in
	// the same direction for all amounts, as the impact of this rounding is expected to be minimal.
	return amount.MulUpFixed(scalingFactor, balancerV2Precision)
}
