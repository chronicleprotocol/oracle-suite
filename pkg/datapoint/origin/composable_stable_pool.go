package origin

import (
	"fmt"
	"math/big"

	"github.com/defiweb/go-eth/types"

	"github.com/chronicleprotocol/oracle-suite/pkg/datapoint/value"
	"github.com/chronicleprotocol/oracle-suite/pkg/util/bn"
)

type ComposableStablePoolConfig struct {
	Pair            value.Pair
	ContractAddress types.Address
}

type LastJoinExitData struct {
	LastJoinExitAmplification *bn.DecFloatPointNumber
	LastPostJoinExitInvariant *bn.DecFloatPointNumber
}

type TokenRateCache struct {
	Rate     *bn.DecFloatPointNumber
	OldRate  *bn.DecFloatPointNumber
	Duration *bn.DecFloatPointNumber
	Expires  *bn.DecFloatPointNumber
}

type AmplificationParameter struct {
	Value      *bn.DecFloatPointNumber
	IsUpdating bool
	Precision  *bn.DecFloatPointNumber
}

type Extra struct {
	AmplificationParameter              AmplificationParameter
	ScalingFactors                      []*bn.DecFloatPointNumber
	LastJoinExit                        LastJoinExitData
	TokensExemptFromYieldProtocolFee    []bool
	TokenRateCaches                     []TokenRateCache
	ProtocolFeePercentageCacheSwapType  *bn.DecFloatPointNumber
	ProtocolFeePercentageCacheYieldType *bn.DecFloatPointNumber
}

type ComposableStablePoolFullConfig struct {
	Pair              value.Pair
	ContractAddress   types.Address
	PoolID            types.Bytes
	Vault             types.Address
	Tokens            []types.Address
	BptIndex          int
	RateProviders     []types.Address
	Balances          []*bn.DecFloatPointNumber
	TotalSupply       *bn.DecFloatPointNumber
	SwapFeePercentage *bn.DecFloatPointNumber
	Extra             Extra
}

type ComposableStablePool struct {
	config ComposableStablePoolFullConfig
}

func NewComposableStablePool(config ComposableStablePoolConfig) (*ComposableStablePool, error) {
	return &ComposableStablePool{
		config: ComposableStablePoolFullConfig{
			Pair:            config.Pair,
			ContractAddress: config.ContractAddress,
		},
	}, nil
}

func NewComposableStablePoolFull(config ComposableStablePoolFullConfig) (*ComposableStablePool, error) {
	return &ComposableStablePool{
		config,
	}, nil
}

// createInitCalls create the calls for `multicall` to get vault address and pool id
func (c *ComposableStablePool) createInitCalls() ([]types.Call, error) {
	if c.config.ContractAddress == types.ZeroAddress {
		return nil, fmt.Errorf("unknown contract address: %s", c.config.Pair.String())
	}

	var calls []types.Call
	// Calls for `getPoolID`
	callData, _ := getPoolID.EncodeArgs()
	calls = append(calls, types.Call{
		To:    &c.config.ContractAddress,
		Input: callData,
	})
	// Calls for `getVault`
	callData, _ = getVault.EncodeArgs()
	calls = append(calls, types.Call{
		To:    &c.config.ContractAddress,
		Input: callData,
	})
	// Calls for `getBptIndex`
	callData, _ = getBptIndex.EncodeArgs()
	calls = append(calls, types.Call{
		To:    &c.config.ContractAddress,
		Input: callData,
	})
	// Calls for `getRateProviders`
	callData, _ = getRateProviders.EncodeArgs()
	calls = append(calls, types.Call{
		To:    &c.config.ContractAddress,
		Input: callData,
	})
	return calls, nil
}

// decodeInitCalls decode the returned bytes of `multicall` that was triggered for `createInitCalls`
func (c *ComposableStablePool) decodeInitCalls(resp [][]byte) error {
	if len(resp) != 4 {
		return fmt.Errorf("not matched response for init calls: %s, %d", c.config.Pair.String(), len(resp))
	}

	var poolID = types.Bytes(resp[0]).PadLeft(32)
	var vault = types.MustAddressFromBytes(resp[1][len(resp[1])-types.AddressLength:])
	var bptIndex = new(big.Int).SetBytes(resp[2]).Int64()
	var rateProviders []types.Address
	if err := getRateProviders.DecodeValues(resp[3], &rateProviders); err != nil {
		return fmt.Errorf("failed decoding rate providers calls: %s, %w", c.config.Pair.String(), err)
	}
	c.config.PoolID = poolID
	c.config.Vault = vault
	c.config.BptIndex = int(bptIndex)
	c.config.RateProviders = rateProviders
	return nil
}

func (c *ComposableStablePool) createPoolTokensCall() (types.Call, error) {
	if c.config.PoolID.String() == "" || c.config.Vault == types.ZeroAddress {
		return types.Call{}, fmt.Errorf("unknown vault or pool id: %s", c.config.Pair.String())
	}

	// Calls for `getPoolTokens`
	callData, _ := getPoolTokens.EncodeArgs(c.config.PoolID.Bytes())
	return types.Call{
		To:    &c.config.Vault,
		Input: callData,
	}, nil
}

func (c *ComposableStablePool) decodePoolTokensCall(resp []byte) error {
	var tokens []types.Address
	var balances []*big.Int
	if err := getPoolTokens.DecodeValues(resp, &tokens, &balances, nil); err != nil {
		return fmt.Errorf("failed decoding pool tokens calls: %s, %w", c.config.Pair.String(), err)
	}
	c.config.Tokens = tokens
	c.config.Balances = make([]*bn.DecFloatPointNumber, len(balances))
	for i, balance := range balances {
		c.config.Balances[i] = bn.DecFloatPoint(balance)
	}
	return nil
}

func (c *ComposableStablePool) createPoolParamsCalls() ([]types.Call, error) {
	if c.config.ContractAddress == types.ZeroAddress {
		return nil, fmt.Errorf("unknown contract address: %s", c.config.Pair.String())
	}

	var calls []types.Call
	// Calls for `getSwapFeePercentage`
	callData, _ := getSwapFeePercentage.EncodeArgs()
	calls = append(calls, types.Call{
		To:    &c.config.ContractAddress,
		Input: callData,
	})
	// Calls for `getAmplificationParameter`
	callData, _ = getAmplificationParameter.EncodeArgs()
	calls = append(calls, types.Call{
		To:    &c.config.ContractAddress,
		Input: callData,
	})
	// Calls for `getScalingFactors`
	callData, _ = getScalingFactors.EncodeArgs()
	calls = append(calls, types.Call{
		To:    &c.config.ContractAddress,
		Input: callData,
	})
	// Calls for `getLastJoinExitData`
	callData, _ = getLastJoinExitData.EncodeArgs()
	calls = append(calls, types.Call{
		To:    &c.config.ContractAddress,
		Input: callData,
	})
	// Calls for `getTotalSupply`
	callData, _ = getTotalSupply.EncodeArgs()
	calls = append(calls, types.Call{
		To:    &c.config.ContractAddress,
		Input: callData,
	})
	// Calls for `getProtocolFeePercentageCache(SWAP)`
	callData, _ = getProtocolFeePercentageCache.EncodeArgs(0)
	calls = append(calls, types.Call{
		To:    &c.config.ContractAddress,
		Input: callData,
	})
	// Calls for `getProtocolFeePercentageCache(YIELD)`
	callData, _ = getProtocolFeePercentageCache.EncodeArgs(2)
	calls = append(calls, types.Call{
		To:    &c.config.ContractAddress,
		Input: callData,
	})
	for _, token := range c.config.Tokens {
		// Calls for `_isTokenExemptFromYieldProtocolFee(token)`
		callData, _ = isTokenExemptFromYieldProtocolFee.EncodeArgs(token)
		calls = append(calls, types.Call{
			To:    &c.config.ContractAddress,
			Input: callData,
		})
	}
	return calls, nil
}

func (c *ComposableStablePool) decodePoolParamsCalls(resp [][]byte) error {
	if len(resp) != 7+len(c.config.Tokens) {
		return fmt.Errorf("not matched response for pool params calls: %s, %d", c.config.Pair.String(), len(resp))
	}
	var swapFeePercentage = new(big.Int).SetBytes(resp[0])
	var amplificationParameter, amplificationPrecision *big.Int
	var isUpdating bool
	if err := getAmplificationParameter.DecodeValues(resp[1], &amplificationParameter, &isUpdating, &amplificationPrecision); err != nil {
		return fmt.Errorf("failed decoding amplification parameter calls: %s, %w", c.config.Pair.String(), err)
	}
	var scalingFactors []*big.Int
	if err := getScalingFactors.DecodeValues(resp[2], &scalingFactors); err != nil {
		return fmt.Errorf("failed decoding scaling factors calls: %s, %w", c.config.Pair.String(), err)
	}
	var lastJoinExitAmplification, lastPostJoinExitInvariant *big.Int
	if err := getLastJoinExitData.DecodeValues(resp[3], &lastJoinExitAmplification, &lastPostJoinExitInvariant); err != nil {
		return fmt.Errorf("failed decoding last join exit calls: %s, %w", c.config.Pair.String(), err)
	}
	var totalSupply = new(big.Int).SetBytes(resp[4])
	var feePercentageCacheSwap = new(big.Int).SetBytes(resp[5])
	var feePercentageCacheYield = new(big.Int).SetBytes(resp[6])
	n := 7
	for i := 0; i < len(c.config.Tokens); i++ {
		var isTokenExempt bool
		if new(big.Int).SetBytes(resp[n]).Cmp(big.NewInt(0)) > 0 {
			isTokenExempt = true
		}
		n++
		c.config.Extra.TokensExemptFromYieldProtocolFee = append(c.config.Extra.TokensExemptFromYieldProtocolFee, isTokenExempt)
	}

	c.config.SwapFeePercentage = bn.DecFloatPoint(swapFeePercentage)
	c.config.Extra.AmplificationParameter.Value = bn.DecFloatPoint(amplificationParameter)
	c.config.Extra.AmplificationParameter.IsUpdating = isUpdating
	c.config.Extra.AmplificationParameter.Precision = bn.DecFloatPoint(amplificationPrecision)
	c.config.Extra.ScalingFactors = make([]*bn.DecFloatPointNumber, len(scalingFactors))
	for i, factor := range scalingFactors {
		c.config.Extra.ScalingFactors[i] = bn.DecFloatPoint(factor)
	}
	c.config.Extra.LastJoinExit.LastJoinExitAmplification = bn.DecFloatPoint(lastJoinExitAmplification)
	c.config.Extra.LastJoinExit.LastPostJoinExitInvariant = bn.DecFloatPoint(lastPostJoinExitInvariant)
	c.config.TotalSupply = bn.DecFloatPoint(totalSupply)
	c.config.Extra.ProtocolFeePercentageCacheSwapType = bn.DecFloatPoint(feePercentageCacheSwap)
	c.config.Extra.ProtocolFeePercentageCacheYieldType = bn.DecFloatPoint(feePercentageCacheYield)
	return nil
}

func (c *ComposableStablePool) createTokenRateCacheCalls() ([]types.Call, error) {
	if len(c.config.Tokens) < 1 || len(c.config.Tokens) != len(c.config.RateProviders) {
		return nil, fmt.Errorf("not found tokens in the pool: %s", c.config.Pair.String())
	}

	var calls []types.Call
	for i, token := range c.config.Tokens {
		if token == c.config.ContractAddress || c.config.RateProviders[i] == types.ZeroAddress {
			continue
		}
		// Calls for `getTokenRateCache(token)`
		callData, _ := getTokenRateCache.EncodeArgs(token)
		calls = append(calls, types.Call{
			To:    &c.config.ContractAddress,
			Input: callData,
		})
	}
	return calls, nil
}

func (c *ComposableStablePool) decodeTokenRateCacheCalls(resp [][]byte) error {
	c.config.Extra.TokenRateCaches = make([]TokenRateCache, len(c.config.Tokens))
	n := 0

	for i, token := range c.config.Tokens {
		if token == c.config.ContractAddress || c.config.RateProviders[i] == types.ZeroAddress {
			continue
		}
		if n >= len(resp) {
			return fmt.Errorf("invalid response for rate cache calls: %s, %d", c.config.Pair.String(), len(resp))
		}
		var rate, oldRate, duration, expires *big.Int
		if err := getTokenRateCache.DecodeValues(resp[n], &rate, &oldRate, &duration, &expires); err != nil {
			return fmt.Errorf("failed decoding token rate cache calls: %s, %w", c.config.Pair.String(), err)
		}
		c.config.Extra.TokenRateCaches[i] = TokenRateCache{
			Rate:     bn.DecFloatPoint(rate),
			OldRate:  bn.DecFloatPoint(oldRate),
			Duration: bn.DecFloatPoint(duration),
			Expires:  bn.DecFloatPoint(expires),
		}
		n++
	}
	return nil
}

func (c *ComposableStablePool) calcAmountOut(tokenIn ERC20Details, tokenOut ERC20Details, amountIn *bn.DecFloatPointNumber) (
	*bn.DecFloatPointNumber,
	*bn.DecFloatPointNumber,
	error,
) {

	indexIn := -1
	indexOut := -1
	for i, address := range c.config.Tokens {
		if address == tokenIn.address {
			indexIn = i
		}
		if address == tokenOut.address {
			indexOut = i
		}
	}
	if indexIn < 0 || indexOut < 0 || indexIn == indexOut {
		return nil, nil, fmt.Errorf("not found tokens in %s: %s, %s", c.config.Pair.String(), tokenIn.symbol, tokenOut.symbol)
	}

	var amountOut, feeAmount *bn.DecFloatPointNumber
	var err error
	if tokenIn.address == c.config.ContractAddress || tokenOut.address == c.config.ContractAddress {
		amountOut, feeAmount, err = c._swapWithBptGivenIn(indexIn, indexOut, amountIn)
	} else {
		amountOut, feeAmount, err = c._swapGivenIn(indexIn, indexOut, amountIn)
	}
	return bn.DecFloatPoint(amountOut), bn.DecFloatPoint(feeAmount), err
}

// _onRegularSwap implements same functionality with the following url:
// https://github.com/balancer/balancer-v2-monorepo/blob/master/pkg/pool-stable/contracts/ComposableStablePool.sol#L283
func (c *ComposableStablePool) _onRegularSwap(
	amountIn *bn.DecFloatPointNumber,
	registeredBalances []*bn.DecFloatPointNumber,
	registeredIndexIn,
	registeredIndexOut int,
) (*bn.DecFloatPointNumber, error) {
	// Adjust indices and balances for BPT token
	// uint256[] memory balances = _dropBptItem(registeredBalances);
	// uint256 indexIn = _skipBptIndex(indexIn);
	// uint256 indexOut = _skipBptIndex(indexOut);

	droppedBalances := c._dropBptItem(registeredBalances)
	indexIn := c._skipBptIndex(registeredIndexIn)
	indexOut := c._skipBptIndex(registeredIndexOut)

	// (uint256 currentAmp, ) = _getAmplificationParameter();
	// uint256 invariant = StableMath._calculateInvariant(currentAmp, balances);
	currentAmp := c.config.Extra.AmplificationParameter.Value
	invariant, err := _calculateInvariant(currentAmp, droppedBalances)
	if err != nil {
		return nil, err
	}

	// StableMath._calcOutGivenIn(currentAmp, balances, indexIn, indexOut, amountGiven, invariant);
	return _calcOutGivenIn(currentAmp, droppedBalances, indexIn, indexOut, amountIn, invariant)
}

// _onSwapGivenIn implements same functionality with the following url:
// https://github.com/balancer/balancer-v2-monorepo/blob/master/pkg/pool-stable/contracts/ComposableStablePool.sol#L242
func (c *ComposableStablePool) _onSwapGivenIn(
	amountIn *bn.DecFloatPointNumber,
	registeredBalances []*bn.DecFloatPointNumber,
	indexIn,
	indexOut int,
) (*bn.DecFloatPointNumber, error) {

	return c._onRegularSwap(amountIn, registeredBalances, indexIn, indexOut)
}

// _swapWithBptGivenIn implements same functionality with the following url:
// https://github.com/balancer/balancer-v2-monorepo/blob/master/pkg/pool-stable/contracts/ComposableStablePool.sol#L314
func (c *ComposableStablePool) _swapWithBptGivenIn(indexIn, indexOut int, amountIn *bn.DecFloatPointNumber) (
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
	balancesUpscaled := c._upscaleArray(c.config.Balances, c.config.Extra.ScalingFactors)
	tokenAmountInScaled := c._upscale(amountIn, c.config.Extra.ScalingFactors[indexIn])

	// (
	//	uint256 preJoinExitSupply,
	//	uint256[] memory balances,
	//	uint256 currentAmp,
	//	uint256 preJoinExitInvariant
	// ) = _beforeJoinExit(registeredBalances);
	preJoinExitSupply, balances, currentAmp, preJoinExitInvariant, err := c._beforeJoinExit(balancesUpscaled)
	if err != nil {
		return nil, nil, err
	}
	if indexOut == c.config.BptIndex {
		// _doJoinSwap(
		//	isGivenIn,
		//	swapRequest.amount,
		//	balances,
		//	_skipBptIndex(registeredIndexIn),
		//	currentAmp,
		//	preJoinExitSupply,
		//	preJoinExitInvariant
		// )
		amountCalculated, _, feeAmount, err = c._doJoinSwap(
			true,
			tokenAmountInScaled,
			balances,
			c._skipBptIndex(indexIn),
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
		amountCalculated, _, feeAmount, err = c._doExitSwap(
			true,
			tokenAmountInScaled,
			balances,
			c._skipBptIndex(indexOut),
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
	return amountCalculated.DivDownFixed(c.config.Extra.ScalingFactors[indexOut], composableStablePrecision), feeAmount, nil
}

// _exitSwapExactBptInForTokenOut implements same functionality with the following url:
// https://github.com/balancer/balancer-v2-monorepo/blob/master/pkg/pool-stable/contracts/ComposableStablePool.sol#L504
func (c *ComposableStablePool) _exitSwapExactBptInForTokenOut(
	bptAmount *bn.DecFloatPointNumber,
	balances []*bn.DecFloatPointNumber,
	indexOut int,
	currentAmp *bn.DecFloatPointNumber,
	actualSupply *bn.DecFloatPointNumber,
	preJoinExitInvariant *bn.DecFloatPointNumber,
) (*bn.DecFloatPointNumber, *bn.DecFloatPointNumber, *bn.DecFloatPointNumber, error) {

	amountOut, feeAmount, err := _calcTokenOutGivenExactBptIn(
		currentAmp, balances, indexOut, bptAmount, actualSupply, preJoinExitInvariant, c.config.SwapFeePercentage)
	if err != nil {
		return nil, nil, nil, err
	}

	balances[indexOut] = balances[indexOut].Sub(amountOut)
	postJoinExitSupply := actualSupply.Sub(bptAmount)

	return amountOut, postJoinExitSupply, feeAmount, nil
}

// _doJoinSwap implements same functionality with the following url:
// https://github.com/balancer/balancer-v2-monorepo/blob/master/pkg/pool-stable/contracts/ComposableStablePool.sol#L375
func (c *ComposableStablePool) _doJoinSwap(
	isGivenIn bool,
	amount *bn.DecFloatPointNumber,
	balances []*bn.DecFloatPointNumber,
	indexIn int,
	currentAmp *bn.DecFloatPointNumber,
	actualSupply *bn.DecFloatPointNumber,
	preJoinExitInvariant *bn.DecFloatPointNumber,
) (*bn.DecFloatPointNumber, *bn.DecFloatPointNumber, *bn.DecFloatPointNumber, error) {

	if isGivenIn {
		return c._joinSwapExactTokenInForBptOut(amount, balances, indexIn, currentAmp, actualSupply, preJoinExitInvariant)
	}
	// Currently ignore givenOut case
	return nil, nil, nil, nil
}

// _doExitSwap implements same functionality with the following url:
// https://github.com/balancer/balancer-v2-monorepo/blob/master/pkg/pool-stable/contracts/ComposableStablePool.sol#L470
func (c *ComposableStablePool) _doExitSwap(
	isGivenIn bool,
	amount *bn.DecFloatPointNumber,
	balances []*bn.DecFloatPointNumber,
	indexOut int,
	currentAmp *bn.DecFloatPointNumber,
	actualSupply *bn.DecFloatPointNumber,
	preJoinExitInvariant *bn.DecFloatPointNumber,
) (*bn.DecFloatPointNumber, *bn.DecFloatPointNumber, *bn.DecFloatPointNumber, error) {

	if isGivenIn {
		return c._exitSwapExactBptInForTokenOut(amount, balances, indexOut, currentAmp, actualSupply, preJoinExitInvariant)
	}
	// Currently ignore givenOut case
	return nil, nil, nil, nil
}

// _joinSwapExactTokenInForBptOut implements same functionality with the following url:
// https://github.com/balancer/balancer-v2-monorepo/blob/master/pkg/pool-stable/contracts/ComposableStablePool.sol#L409
func (c *ComposableStablePool) _joinSwapExactTokenInForBptOut(
	amountIn *bn.DecFloatPointNumber,
	balances []*bn.DecFloatPointNumber,
	indexIn int,
	currentAmp *bn.DecFloatPointNumber,
	actualSupply *bn.DecFloatPointNumber,
	preJoinExitInvariant *bn.DecFloatPointNumber,
) (*bn.DecFloatPointNumber, *bn.DecFloatPointNumber, *bn.DecFloatPointNumber, error) {

	amountsIn := make([]*bn.DecFloatPointNumber, len(balances))
	for i := range amountsIn {
		amountsIn[i] = bn.DecFloatPoint(0)
	}
	amountsIn[indexIn] = amountIn
	bptOut, feeAmountIn, err := _calcBptOutGivenExactTokensIn(
		currentAmp, balances, amountsIn, actualSupply, preJoinExitInvariant, c.config.SwapFeePercentage)
	if err != nil {
		return nil, nil, nil, err
	}
	balances[indexIn] = balances[indexIn].Add(amountIn)
	postJoinExitSupply := actualSupply.Add(bptOut)

	return bptOut, postJoinExitSupply, feeAmountIn, nil
}

// _beforeJoinExit implements same functionality with the following url:
// https://github.com/balancer/balancer-v2-monorepo/blob/master/pkg/pool-stable/contracts/ComposableStablePool.sol#L701
func (c *ComposableStablePool) _beforeJoinExit(registeredBalances []*bn.DecFloatPointNumber) (
	*bn.DecFloatPointNumber,
	[]*bn.DecFloatPointNumber,
	*bn.DecFloatPointNumber,
	*bn.DecFloatPointNumber,
	error,
) {

	preJoinExitSupply, balances, oldAmpPreJoinExitInvariant, err := c._payProtocolFeesBeforeJoinExit(registeredBalances)
	if err != nil {
		return nil, nil, nil, nil, err
	}
	currentAmp := c.config.Extra.AmplificationParameter.Value

	var (
		preJoinExitInvariant *bn.DecFloatPointNumber
	)

	if currentAmp.Cmp(c.config.Extra.LastJoinExit.LastJoinExitAmplification) == 0 {
		preJoinExitInvariant = oldAmpPreJoinExitInvariant
	} else {
		preJoinExitInvariant, err = _calculateInvariant(currentAmp, balances)
	}
	if err != nil {
		return nil, nil, nil, nil, err
	}

	return preJoinExitSupply, balances, currentAmp, preJoinExitInvariant, nil
}

// _payProtocolFeesBeforeJoinExit implements same functionality with the following url:
// https://github.com/balancer/balancer-v2-monorepo/blob/master/pkg/pool-stable/contracts/ComposableStablePoolProtocolFees.sol#L64
func (c *ComposableStablePool) _payProtocolFeesBeforeJoinExit(
	registeredBalances []*bn.DecFloatPointNumber,
) (*bn.DecFloatPointNumber, []*bn.DecFloatPointNumber, *bn.DecFloatPointNumber, error) {

	virtualSupply, droppedBalances := c._dropBptItemFromBalances(registeredBalances)
	expectedProtocolOwnershipPercentage, currentInvariantWithLastJoinExitAmp, err := c._getProtocolPoolOwnershipPercentage(droppedBalances)
	if err != nil {
		return nil, nil, nil, err
	}
	protocolFeeAmount := c._bptForPoolOwnershipPercentage(virtualSupply, expectedProtocolOwnershipPercentage)

	return virtualSupply.Add(protocolFeeAmount), droppedBalances, currentInvariantWithLastJoinExitAmp, nil
}

// _getProtocolPoolOwnershipPercentage implements same functionality with the following url:
// https://github.com/balancer/balancer-v2-monorepo/blob/master/pkg/pool-stable/contracts/ComposableStablePoolProtocolFees.sol#L102
func (c *ComposableStablePool) _getProtocolPoolOwnershipPercentage(balances []*bn.DecFloatPointNumber) (
	*bn.DecFloatPointNumber,
	*bn.DecFloatPointNumber,
	error,
) {

	swapFeeGrowthInvariant, totalNonExemptGrowthInvariant, totalGrowthInvariant, err := c._getGrowthInvariants(balances)
	if err != nil {
		return nil, nil, err
	}
	// Calculate the delta for swap fee growth invariant
	swapFeeGrowthInvariantDelta := swapFeeGrowthInvariant.Sub(c.config.Extra.LastJoinExit.LastPostJoinExitInvariant)
	if swapFeeGrowthInvariantDelta.Cmp(bigIntZero) < 0 {
		swapFeeGrowthInvariantDelta = bn.DecFloatPoint(0)
	}

	// Calculate the delta for non-exempt yield growth invariant
	nonExemptYieldGrowthInvariantDelta := totalNonExemptGrowthInvariant.Sub(swapFeeGrowthInvariant)
	if nonExemptYieldGrowthInvariantDelta.Cmp(bigIntZero) < 0 {
		nonExemptYieldGrowthInvariantDelta = bn.DecFloatPoint(0)
	}

	// swapFeeGrowthInvariantDelta/totalGrowthInvariant*getProtocolFeePercentageCache
	protocolSwapFeePercentage :=
		swapFeeGrowthInvariantDelta.DivDownFixed(totalGrowthInvariant, composableStablePrecision).MulDownFixed(
			c.config.Extra.ProtocolFeePercentageCacheSwapType, composableStablePrecision)

	protocolYieldPercentage :=
		nonExemptYieldGrowthInvariantDelta.DivDownFixed(totalGrowthInvariant, composableStablePrecision).MulDownFixed(
			c.config.Extra.ProtocolFeePercentageCacheYieldType, composableStablePrecision)

	// Calculate the total protocol ComposableStablePool ownership percentage
	protocolPoolOwnershipPercentage := protocolSwapFeePercentage.Add(protocolYieldPercentage)

	return protocolPoolOwnershipPercentage, totalGrowthInvariant, nil
}

// _getGrowthInvariants implements same functionality with the following url:
// https://github.com/balancer/balancer-v2-monorepo/blob/master/pkg/pool-stable/contracts/ComposableStablePoolProtocolFees.sol#L189
func (c *ComposableStablePool) _getGrowthInvariants(balances []*bn.DecFloatPointNumber) (
	*bn.DecFloatPointNumber,
	*bn.DecFloatPointNumber,
	*bn.DecFloatPointNumber,
	error,
) {

	var (
		swapFeeGrowthInvariant        *bn.DecFloatPointNumber
		totalNonExemptGrowthInvariant *bn.DecFloatPointNumber
		totalGrowthInvariant          *bn.DecFloatPointNumber
		err                           error
	)

	// This invariant result is calc by _divDown (round down)
	// https://github.com/balancer/balancer-v2-monorepo/blob/b46023f7c5deefaf58a0a42559a36df420e1639f/pkg/pool-stable/contracts/StableMath.sol#L96
	swapFeeGrowthInvariant, err = _calculateInvariant(
		c.config.Extra.LastJoinExit.LastJoinExitAmplification,
		c._getAdjustedBalances(balances, true))
	if err != nil {
		return nil, nil, nil, err
	}

	// For the other invariants, we can potentially skip some work. In the edge cases where none or all of the
	// tokens are exempt from yield, there's one fewer invariant to compute.
	switch {
	case c._areNoTokensExempt():
		// If there are no tokens with fee-exempt yield, then the total non-exempt growth will equal the total
		// growth: all yield growth is non-exempt. There's also no point in adjusting balances, since we
		// already know none are exempt.
		totalNonExemptGrowthInvariant, err = _calculateInvariant(c.config.Extra.LastJoinExit.LastJoinExitAmplification, balances)
		if err != nil {
			return nil, nil, nil, err
		}

		totalGrowthInvariant = totalNonExemptGrowthInvariant
	case c._areAllTokensExempt():
		// If no tokens are charged fees on yield, then the non-exempt growth is equal to the swap fee growth - no
		// yield fees will be collected.
		totalNonExemptGrowthInvariant = swapFeeGrowthInvariant
		totalGrowthInvariant, err = _calculateInvariant(c.config.Extra.LastJoinExit.LastJoinExitAmplification, balances)
		if err != nil {
			return nil, nil, nil, err
		}
	default:
		// In the general case, we need to calculate two invariants: one with some adjusted balances, and one with
		// the current balances.

		totalNonExemptGrowthInvariant, err = _calculateInvariant(
			c.config.Extra.LastJoinExit.LastJoinExitAmplification,
			c._getAdjustedBalances(balances, false), // Only adjust non-exempt balances
		)
		if err != nil {
			return nil, nil, nil, err
		}

		totalGrowthInvariant, err = _calculateInvariant(
			c.config.Extra.LastJoinExit.LastJoinExitAmplification,
			balances)
		if err != nil {
			return nil, nil, nil, err
		}
	}
	return swapFeeGrowthInvariant, totalNonExemptGrowthInvariant, totalGrowthInvariant, nil
}

// _dropBptItemFromBalances implements same functionality with the following url:
// https://github.com/balancer/balancer-v2-monorepo/blob/master/pkg/pool-stable/contracts/ComposableStablePoolStorage.sol#L259
func (c *ComposableStablePool) _dropBptItemFromBalances(balances []*bn.DecFloatPointNumber) (*bn.DecFloatPointNumber, []*bn.DecFloatPointNumber) {
	return c._getVirtualSupply(balances[c.config.BptIndex]), c._dropBptItem(balances)
}

// _getVirtualSupply implements same functionality with the following url:
// https://github.com/balancer/balancer-v2-monorepo/blob/master/pkg/pool-stable/contracts/ComposableStablePoolStorage.sol#L386
func (c *ComposableStablePool) _getVirtualSupply(bptBalance *bn.DecFloatPointNumber) *bn.DecFloatPointNumber {
	return c.config.TotalSupply.Sub(bptBalance)
}

// _hasRateProvider implements same functionality with the following url:
// https://github.com/balancer/balancer-v2-monorepo/blob/master/pkg/pool-stable/contracts/ComposableStablePoolStorage.sol#L337
func (c *ComposableStablePool) _hasRateProvider(tokenIndex int) bool {
	return c.config.RateProviders[tokenIndex] != types.ZeroAddress
}

// isTokenExemptFromYieldProtocolFee implements same functionality with the following url:
// https://github.com/balancer/balancer-v2-monorepo/blob/master/pkg/pool-stable/contracts/ComposableStablePoolStorage.sol#L362
func (c *ComposableStablePool) _isTokenExemptFromYieldProtocolFee(tokenIndex int) bool {
	return c.config.Extra.TokensExemptFromYieldProtocolFee[tokenIndex]
}

// _areNoTokensExempt implements same functionality with the following url:
// https://github.com/balancer/balancer-v2-monorepo/blob/master/pkg/pool-stable/contracts/ComposableStablePoolStorage.sol#L351
func (c *ComposableStablePool) _areNoTokensExempt() bool {
	for _, exempt := range c.config.Extra.TokensExemptFromYieldProtocolFee {
		if exempt {
			return false
		}
	}
	return true
}

// _areAllTokensExempt implements same functionality with the following url:
// https://github.com/balancer/balancer-v2-monorepo/blob/master/pkg/pool-stable/contracts/ComposableStablePoolStorage.sol#L344
func (c *ComposableStablePool) _areAllTokensExempt() bool {
	for _, exempt := range c.config.Extra.TokensExemptFromYieldProtocolFee {
		if !exempt {
			return false
		}
	}
	return true
}

// _getAdjustedBalances implements same functionality with the following url:
// https://github.com/balancer/balancer-v2-monorepo/blob/master/pkg/pool-stable/contracts/ComposableStablePoolRates.sol#L222
func (c *ComposableStablePool) _getAdjustedBalances(balances []*bn.DecFloatPointNumber, ignoreExemptFlags bool) []*bn.DecFloatPointNumber {
	totalTokensWithoutBpt := len(balances)
	adjustedBalances := make([]*bn.DecFloatPointNumber, totalTokensWithoutBpt)

	for i := 0; i < totalTokensWithoutBpt; i++ {
		skipBptIndex := i
		if i >= c.config.BptIndex {
			skipBptIndex++
		}

		if c._isTokenExemptFromYieldProtocolFee(skipBptIndex) || (ignoreExemptFlags && c._hasRateProvider(skipBptIndex)) {
			adjustedBalances[i] = c._adjustedBalance(balances[i], &c.config.Extra.TokenRateCaches[skipBptIndex])
		} else {
			adjustedBalances[i] = balances[i]
		}
	}

	return adjustedBalances
}

// _adjustedBalance implements same functionality with the following url:
// https://github.com/balancer/balancer-v2-monorepo/blob/master/pkg/pool-stable/contracts/ComposableStablePoolRates.sol#L242
func (c *ComposableStablePool) _adjustedBalance(balance *bn.DecFloatPointNumber, cache *TokenRateCache) *bn.DecFloatPointNumber {
	return balance.Mul(cache.OldRate).DivPrec(cache.Rate, 0)
}

// _dropBptItem implements same functionality with the following url:
// https://github.com/balancer/balancer-v2-monorepo/blob/master/pkg/pool-stable/contracts/ComposableStablePoolStorage.sol#L246
func (c *ComposableStablePool) _dropBptItem(amounts []*bn.DecFloatPointNumber) []*bn.DecFloatPointNumber {
	amountsWithoutBpt := make([]*bn.DecFloatPointNumber, len(amounts)-1)
	bptIndex := c.config.BptIndex

	for i := 0; i < len(amountsWithoutBpt); i++ {
		if i < bptIndex {
			amountsWithoutBpt[i] = amounts[i]
		} else {
			amountsWithoutBpt[i] = amounts[i+1]
		}
	}
	return amountsWithoutBpt
}

func (c *ComposableStablePool) _bptForPoolOwnershipPercentage(totalSupply, poolOwnershipPercentage *bn.DecFloatPointNumber) *bn.DecFloatPointNumber {
	// If we mint some amount `bptAmount` of BPT then the percentage ownership of the pool this grants is given by:
	// `poolOwnershipPercentage = bptAmount / (totalSupply + bptAmount)`.
	// Solving for `bptAmount`, we arrive at:
	// `bptAmount = totalSupply * poolOwnershipPercentage / (1 - poolOwnershipPercentage)`.
	return totalSupply.Mul(poolOwnershipPercentage).DivPrec(_complementFixed(poolOwnershipPercentage), 0)
}

// _skipBptIndex implements same functionality with the following url:
// https://github.com/balancer/balancer-v2-monorepo/blob/master/pkg/pool-stable/contracts/ComposableStablePoolStorage.sol#L235
func (c *ComposableStablePool) _skipBptIndex(index int) int {
	if index < c.config.BptIndex {
		return index
	}
	return index - 1
}

// _swapGivenIn simulates the functionality of `_swapGivenIn` in `ComposableStablePool`
// https://github.com/balancer/balancer-v2-monorepo/blob/master/pkg/pool-stable/contracts/ComposableStablePool.sol#L187
func (c *ComposableStablePool) _swapGivenIn(indexIn, indexOut int, amountIn *bn.DecFloatPointNumber) (
	*bn.DecFloatPointNumber,
	*bn.DecFloatPointNumber,
	error,
) {
	// Fees are subtracted before scaling, to reduce the complexity of the rounding direction analysis.
	// swapRequest.amount = _subtractSwapFeeAmount(swapRequest.amount);
	amountAfterFee, feeAmount := c._subtractSwapFeeAmount(amountIn, c.config.SwapFeePercentage)

	// _upscaleArray(balances, scalingFactors);
	// swapRequest.amount = _upscale(swapRequest.amount, scalingFactors[indexIn]);
	upscaledBalances := c._upscaleArray(c.config.Balances, c.config.Extra.ScalingFactors)
	amountUpScale := c._upscale(amountAfterFee, c.config.Extra.ScalingFactors[indexIn])

	// uint256 amountOut = _onSwapGivenIn(swapRequest, balances, indexIn, indexOut);
	amountOut, err := c._onSwapGivenIn(amountUpScale, upscaledBalances, indexIn, indexOut)
	if err != nil {
		return nil, nil, err
	}

	return amountOut.DivDownFixed(c.config.Extra.ScalingFactors[indexOut], composableStablePrecision), feeAmount, nil
}

func (c *ComposableStablePool) _subtractSwapFeeAmount(amount, swapFeePercentage *bn.DecFloatPointNumber) (
	*bn.DecFloatPointNumber,
	*bn.DecFloatPointNumber,
) {

	feeAmount := amount.MulUpFixed(swapFeePercentage, composableStablePrecision)
	return amount.Sub(feeAmount), feeAmount
}

func (c *ComposableStablePool) _upscaleArray(amounts, scalingFactors []*bn.DecFloatPointNumber) []*bn.DecFloatPointNumber {
	result := make([]*bn.DecFloatPointNumber, len(amounts))
	for i, amount := range amounts {
		result[i] = amount.MulUpFixed(scalingFactors[i], composableStablePrecision)
	}
	return result
}

func (c *ComposableStablePool) _upscale(amount, scalingFactor *bn.DecFloatPointNumber) *bn.DecFloatPointNumber {
	return amount.MulUpFixed(scalingFactor, composableStablePrecision)
}
