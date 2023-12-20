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

type AmplificationParameter struct {
	value      *bn.DecFixedPointNumber
	isUpdating bool
	precision  *bn.DecFixedPointNumber
}

type Extra struct {
	amplificationParameter AmplificationParameter
	scalingFactors         []*bn.DecFixedPointNumber
}

type ComposableStablePool struct {
	pair    value.Pair
	address types.Address

	tokens            []types.Address
	balances          []*bn.DecFixedPointNumber
	bptIndex          int
	totalSupply       *bn.DecFixedPointNumber
	swapFeePercentage *bn.DecFixedPointNumber
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
		var decBalances []*bn.DecFixedPointNumber
		for _, balance := range balances {
			decBalances = append(decBalances, bn.DecFixedPoint(balance, 0))
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
		// Calls for `getSwapFeePercentage`
		callData, _ = getSwapFeePercentage.EncodeArgs()
		calls = append(calls, types.Call{To: &pool.address, Input: callData})
		// Calls for `getAmplificationParameter`
		callData, _ = getAmplificationParameter.EncodeArgs()
		calls = append(calls, types.Call{To: &pool.address, Input: callData})
		// Calls for `getScalingFactors`
		callData, _ = getScalingFactors.EncodeArgs()
		calls = append(calls, types.Call{To: &pool.address, Input: callData})
		// Calls for `getTotalSupply`
		callData, _ = getTotalSupply.EncodeArgs()
		calls = append(calls, types.Call{To: &pool.address, Input: callData})
	}

	resp, err := ethereum.MultiCall(ctx, c.client, calls, blockNumber)
	if err != nil {
		return err
	}
	n := len(resp) / len(c.pools)
	for i, pool := range c.pools {
		pool.bptIndex = int(new(big.Int).SetBytes(resp[i*n]).Int64())
		pool.swapFeePercentage = bn.DecFixedPoint(new(big.Int).SetBytes(resp[i*n+1]), 0)
		var amplificationParameter, amplificationPrecision *big.Int
		var isUpdating bool
		if err := getAmplificationParameter.DecodeValues(resp[i*n+2], &amplificationParameter, &isUpdating, &amplificationPrecision); err != nil {
			return fmt.Errorf("failed decoding amplification parameter calls: %s, %w", pool.pair.String(), err)
		}
		var scalingFactors []*big.Int
		if err := getScalingFactors.DecodeValues(resp[i*n+3], &scalingFactors); err != nil {
			return fmt.Errorf("failed decoding scaling factors calls: %s, %w", pool.pair.String(), err)
		}
		pool.totalSupply = bn.DecFixedPoint(new(big.Int).SetBytes(resp[i*n+4]), 0)
		pool.extra.amplificationParameter.value = bn.DecFixedPoint(amplificationParameter, 0)
		pool.extra.amplificationParameter.isUpdating = isUpdating
		pool.extra.amplificationParameter.precision = bn.DecFixedPoint(amplificationPrecision, 0)
		pool.extra.scalingFactors = make([]*bn.DecFixedPointNumber, len(scalingFactors))
		for j, factor := range scalingFactors {
			pool.extra.scalingFactors[j] = bn.DecFixedPoint(factor, 0)
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

func (p *ComposableStablePool) CalcAmountOut(tokenIn, tokenOut types.Address, amountIn *bn.DecFixedPointNumber) (
	*bn.DecFixedPointNumber,
	*bn.DecFixedPointNumber,
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

	if tokenIn == p.address || tokenOut == p.address {
		return nil, nil, fmt.Errorf("unsupported token swap")
	}
	return p._swapGivenIn(indexIn, indexOut, amountIn)
}

// _onRegularSwap implements same functionality with the following url:
// https://github.com/balancer/balancer-v2-monorepo/blob/master/pkg/pool-stable/contracts/ComposableStablePool.sol#L283
func (p *ComposableStablePool) _onRegularSwap(
	amountIn *bn.DecFixedPointNumber,
	registeredBalances []*bn.DecFixedPointNumber,
	registeredIndexIn,
	registeredIndexOut int,
) (*bn.DecFixedPointNumber, error) {
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
	amountIn *bn.DecFixedPointNumber,
	registeredBalances []*bn.DecFixedPointNumber,
	indexIn,
	indexOut int,
) (*bn.DecFixedPointNumber, error) {

	return p._onRegularSwap(amountIn, registeredBalances, indexIn, indexOut)
}

// Remove the item at `_bptIndex` from an arbitrary array (e.g., amountsIn).
// Reference: https://github.com/balancer/balancer-v2-monorepo/blob/master/pkg/pool-stable/contracts/ComposableStablePoolStorage.sol#L246
func (p *ComposableStablePool) _dropBptItem(amounts []*bn.DecFixedPointNumber) []*bn.DecFixedPointNumber {
	amountsWithoutBpt := make([]*bn.DecFixedPointNumber, len(amounts)-1)
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
func (p *ComposableStablePool) _swapGivenIn(indexIn, indexOut int, amountIn *bn.DecFixedPointNumber) (
	*bn.DecFixedPointNumber,
	*bn.DecFixedPointNumber,
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

	return _divDownFixed18(amountOut, p.extra.scalingFactors[indexOut]), feeAmount, nil
}

// Subtracts swap fee amount from `amount`, returning a lower value.
// Reference: https://github.com/balancer/balancer-v2-monorepo/blob/master/pkg/pool-utils/contracts/BasePool.sol#L603
func (p *ComposableStablePool) _subtractSwapFeeAmount(amount, swapFeePercentage *bn.DecFixedPointNumber) (
	*bn.DecFixedPointNumber,
	*bn.DecFixedPointNumber,
) {

	feeAmount := _mulUpFixed18(amount, swapFeePercentage)
	return amount.Sub(feeAmount), feeAmount
}

// Same as `_upscale`, but for an entire array. This function does not return anything, but instead *mutates*
// the `amounts` array.
// Reference: https://github.com/balancer/balancer-v2-monorepo/blob/master/pkg/solidity-utils/contracts/helpers/ScalingHelpers.sol#L64C1-L64C1
func (p *ComposableStablePool) _upscaleArray(amounts, scalingFactors []*bn.DecFixedPointNumber) []*bn.DecFixedPointNumber {
	result := make([]*bn.DecFixedPointNumber, len(amounts))
	for i, amount := range amounts {
		result[i] = _mulUpFixed18(amount, scalingFactors[i])
	}
	return result
}

// Applies `scalingFactor` to `amount`, resulting in a larger or equal value depending on whether it needed
// scaling or not.
// https://github.com/balancer/balancer-v2-monorepo/blob/master/pkg/solidity-utils/contracts/helpers/ScalingHelpers.sol#L32
func (p *ComposableStablePool) _upscale(amount, scalingFactor *bn.DecFixedPointNumber) *bn.DecFixedPointNumber {
	// Upscale rounding wouldn't necessarily always go in the same direction: in a swap for example the balance of
	// token in should be rounded up, and that of token out rounded down. This is the only place where we round in
	// the same direction for all amounts, as the impact of this rounding is expected to be minimal.
	return _mulUpFixed18(amount, scalingFactor)
}
