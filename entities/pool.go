package entities

import (
	"errors"
	"math/big"

	"github.com/daoleno/uniswap-sdk-core/entities"
	"github.com/ethereum/go-ethereum/common"

	"github.com/KyberNetwork/elastic-go-sdk/v2/constants"
	"github.com/KyberNetwork/elastic-go-sdk/v2/utils"
)

const (
	MaxTickDistance = 480
)

var (
	ErrFeeTooHigh          = errors.New("fee too high")
	ErrInvalidSqrtRatioX96 = errors.New("invalid sqrtRatioX96")
	ErrTokenNotInvolved    = errors.New("token not involved in pool")
	ErrBadLimitSqrtP       = errors.New("bad limitSqrtP")
)

type SwapData struct {
	specifiedAmount *big.Int // the specified amount (could be tokenIn or tokenOut)
	returnedAmount  *big.Int // the opposite amout of sourceQty
	sqrtP           *big.Int // current sqrt(price), multiplied by 2^96
	currentTick     int      // the tick associated with the current price
	nextTick        int      // the next initialized tick
	nextSqrtP       *big.Int // the price of nextTick
	isToken0        bool     // true if specifiedAmount is in token0, false if in token1
	isExactInput    bool     // true = input qty, false = output qty
	baseL           *big.Int // the cached base pool liquidity without reinvestment liquidity
	reinvestL       *big.Int // the cached reinvestment liquidity
	startSqrtP      *big.Int // the start sqrt price before each iteration
}

// Represents a V3 pool
type Pool struct {
	Token0             *entities.Token
	Token1             *entities.Token
	Fee                constants.FeeAmount
	SqrtP              *big.Int
	BaseL              *big.Int
	ReinvestL          *big.Int
	CurrentTick        int
	NearestCurrentTick int
	Ticks              map[int]TickData
	InitializedTicks   map[int]LinkedListData

	token0Price *entities.Price
	token1Price *entities.Price
}

func GetAddress(
	tokenA, tokenB *entities.Token, fee constants.FeeAmount, initCodeHashManualOverride string,
) (common.Address, error) {
	return utils.ComputePoolAddress(constants.FactoryAddress, tokenA, tokenB, fee, initCodeHashManualOverride)
}

/**
 * Construct a pool
 * @param tokenA One of the tokens in the pool
 * @param tokenB The other token in the pool
 * @param fee The fee in hundredths of a bips of the input amount of every swap that is collected by the pool
 * @param sqrtRatioX96 The sqrt of the current ratio of amounts of token1 to token0
 * @param liquidity The current value of in range liquidity
 * @param tickCurrent The current tick of the pool
 * @param ticks The current state of the pool ticks or a data provider that can return tick data
 */
func NewPool(
	tokenA *entities.Token,
	tokenB *entities.Token,
	fee constants.FeeAmount,
	sqrtRatioX96 *big.Int,
	liquidity *big.Int,
	reinvestLiquidity *big.Int,
	tickCurrent int,
	tickDataProvider TickDataProvider,
) (*Pool, error) {
	if fee >= constants.FeeMax {
		return nil, ErrFeeTooHigh
	}

	token0 := tokenA
	token1 := tokenB
	isSorted, err := tokenA.SortsBefore(tokenB)
	if err != nil {
		return nil, err
	}
	if !isSorted {
		token0 = tokenB
		token1 = tokenA
	}

	tickCurrentSqrtRatioX96, err := utils.GetSqrtRatioAtTick(tickCurrent)
	if err != nil {
		return nil, err
	}
	nextTickSqrtRatioX96, err := utils.GetSqrtRatioAtTick(tickCurrent + 1)
	if err != nil {
		return nil, err
	}

	if sqrtRatioX96.Cmp(tickCurrentSqrtRatioX96) < 0 || sqrtRatioX96.Cmp(nextTickSqrtRatioX96) > 0 {
		return nil, ErrInvalidSqrtRatioX96
	}

	var nearestCurrentTick int
	if tickDataProvider == nil {
		tickDataProvider, err = NewTickListDataProvider([]Tick{}, constants.TickSpacings[fee])
		if err != nil {
			return nil, err
		}

		nearestCurrentTick = utils.MinTick
	} else {
		nearestCurrentTick, err = tickDataProvider.GetNearestCurrentTick(tickCurrent)
		if err != nil {
			return nil, err
		}
	}

	ticks, initializedTicks := tickDataProvider.TransformToMap()

	return &Pool{
		Token0:             token0,
		Token1:             token1,
		Fee:                fee,
		SqrtP:              sqrtRatioX96,
		BaseL:              liquidity,
		ReinvestL:          reinvestLiquidity,
		CurrentTick:        tickCurrent,
		NearestCurrentTick: nearestCurrentTick,
		Ticks:              ticks,
		InitializedTicks:   initializedTicks,
	}, nil
}

/**
 * Returns true if the token is either token0 or token1
 * @param token The token to check
 * @returns True if token is either token0 or token
 */
func (p *Pool) InvolvesToken(token *entities.Token) bool {
	return p.Token0.Equal(token) || p.Token1.Equal(token)
}

// Token0Price returns the current mid price of the pool in terms of token0, i.e. the ratio of token1 over token0
func (p *Pool) Token0Price() *entities.Price {
	if p.token0Price != nil {
		return p.token0Price
	}
	p.token0Price = entities.NewPrice(
		p.Token0, p.Token1, constants.Q192, new(big.Int).Mul(p.SqrtP, p.SqrtP),
	)
	return p.token0Price
}

// Token1Price returns the current mid price of the pool in terms of token1, i.e. the ratio of token0 over token1
func (p *Pool) Token1Price() *entities.Price {
	if p.token1Price != nil {
		return p.token1Price
	}
	p.token1Price = entities.NewPrice(
		p.Token1, p.Token0, new(big.Int).Mul(p.SqrtP, p.SqrtP), constants.Q192,
	)
	return p.token1Price
}

/**
 * Return the price of the given token in terms of the other token in the pool.
 * @param token The token to return price of
 * @returns The price of the given token, in terms of the other.
 */
func (p *Pool) PriceOf(token *entities.Token) (*entities.Price, error) {
	if !p.InvolvesToken(token) {
		return nil, ErrTokenNotInvolved
	}
	if p.Token0.Equal(token) {
		return p.Token0Price(), nil
	}
	return p.Token1Price(), nil
}

// ChainId returns the chain ID of the tokens in the pool.
func (p *Pool) ChainID() uint {
	return p.Token0.ChainId()
}

/**
 * Given an input amount of a token, return the computed output amount, and a pool with state updated after the trade
 * @param inputAmount The input amount for which to quote the output amount
 * @param sqrtPriceLimitX96 The Q64.96 sqrt price limit
 * @returns The output amount and the pool with updated state
 */
func (p *Pool) GetOutputAmount(
	inputAmount *entities.CurrencyAmount, limitSqrtP *big.Int,
) (*entities.CurrencyAmount, *Pool, error) {
	if !(inputAmount.Currency.IsToken() && p.InvolvesToken(inputAmount.Currency.Wrapped())) {
		return nil, nil, ErrTokenNotInvolved
	}
	zeroForOne := inputAmount.Currency.Equal(p.Token0)
	returnedAmount, baseL, reinvestL, sqrtP, currentTick, nextTick, err := p.swap(
		zeroForOne,
		inputAmount.Quotient(),
		limitSqrtP,
	)
	if err != nil {
		return nil, nil, err
	}

	var outputToken *entities.Token
	if zeroForOne {
		outputToken = p.Token1
	} else {
		outputToken = p.Token0
	}

	newPoolState := p._updatePoolData(baseL, reinvestL, sqrtP, currentTick, nextTick)

	return entities.FromRawAmount(outputToken, new(big.Int).Mul(returnedAmount, constants.NegativeOne)), newPoolState, nil
}

/**
 * Given a desired output amount of a token, return the computed input amount and a pool with state updated after the trade
 * @param outputAmount the output amount for which to quote the input amount
 * @param sqrtPriceLimitX96 The Q64.96 sqrt price limit. If zero for one, the price cannot be less than this value after the swap. If one for zero, the price cannot be greater than this value after the swap
 * @returns The input amount and the pool with updated state
 */
func (p *Pool) GetInputAmount(
	outputAmount *entities.CurrencyAmount, limitSqrtP *big.Int,
) (*entities.CurrencyAmount, *Pool, error) {
	if !(outputAmount.Currency.IsToken() && p.InvolvesToken(outputAmount.Currency.Wrapped())) {
		return nil, nil, ErrTokenNotInvolved
	}
	zeroForOne := outputAmount.Currency.Equal(p.Token1)
	returnedAmount, baseL, reinvestL, sqrtP, currentTick, nextTick, err := p.swap(
		zeroForOne,
		new(big.Int).Mul(outputAmount.Quotient(), constants.NegativeOne),
		limitSqrtP,
	)
	if err != nil {
		return nil, nil, err
	}

	var inputToken *entities.Token
	if zeroForOne {
		inputToken = p.Token0
	} else {
		inputToken = p.Token1
	}

	newPoolState := p._updatePoolData(baseL, reinvestL, sqrtP, currentTick, nextTick)

	return entities.FromRawAmount(inputToken, returnedAmount), newPoolState, nil
}

// Source: https://github.com/KyberNetwork/ks-elastic-sc-v2/blob/3ba84353cbd88f30f222bb9c673e242a2e46fd12/contracts/PoolTicksState.sol#L121-L147C4
func (p *Pool) _getInitialSwapData(willUpTick bool) (
	baseL *big.Int,
	reinvestL *big.Int,
	sqrtP *big.Int,
	currentTick int,
	nextTick int,
) {
	baseL = p.BaseL
	reinvestL = p.ReinvestL
	sqrtP = p.SqrtP
	currentTick = p.CurrentTick
	nextTick = p.NearestCurrentTick
	if willUpTick {
		nextTick = p.InitializedTicks[nextTick].Next
	}

	return
}

/**
 * Executes a swap
 * @param zeroForOne Whether the amount in is token0 or token1
 * @param amountSpecified The amount of the swap, which implicitly configures the swap as exact input (positive), or exact output (negative)
 * @param sqrtPriceLimitX96 The Q64.96 sqrt price limit. If zero for one, the price cannot be less than this value after the swap. If one for zero, the price cannot be greater than this value after the swap
 * @returns returnedAmount
 * @returns sqrtRatioX96
 * @returns liquidity
 * @returns tickCurrent
 */
func (p *Pool) swap(isToken0 bool, swapQty *big.Int, limitSqrtP *big.Int) (
	*big.Int, *big.Int, *big.Int, *big.Int, int, int, error,
) {
	var swapData SwapData
	swapData.specifiedAmount = swapQty
	swapData.isToken0 = isToken0
	swapData.isExactInput = swapData.specifiedAmount.Cmp(constants.Zero) > 0
	willUpTick := swapData.isExactInput != isToken0

	// keep track of swap state
	swapData.baseL,
		swapData.reinvestL,
		swapData.sqrtP,
		swapData.currentTick,
		swapData.nextTick = p._getInitialSwapData(willUpTick)
	swapData.returnedAmount = constants.Zero

	// ad-hoc logic in the SDK in case of misconfiguration in the SDK's consumer code
	if limitSqrtP == nil {
		if willUpTick {
			limitSqrtP = new(big.Int).Sub(utils.MaxSqrtRatio, constants.One)
		} else {
			limitSqrtP = new(big.Int).Add(utils.MinSqrtRatio, constants.One)
		}
	}

	if willUpTick {
		if limitSqrtP.Cmp(p.SqrtP) < 0 || limitSqrtP.Cmp(utils.MaxSqrtRatio) > 0 {
			return nil, nil, nil, nil, 0, 0, ErrBadLimitSqrtP
		}
	} else {
		if limitSqrtP.Cmp(p.SqrtP) > 0 || limitSqrtP.Cmp(utils.MinSqrtRatio) < 0 {
			return nil, nil, nil, nil, 0, 0, ErrBadLimitSqrtP
		}
	}

	var err error

	// continue swapping while specified input/output isn't satisfied or price limit not reached
	for swapData.specifiedAmount.Cmp(constants.Zero) != 0 && swapData.sqrtP.Cmp(limitSqrtP) != 0 {
		// math calculations work with the assumption that the price diff is capped to 5%
		// since tick distance is uncapped between currentTick and nextTick
		// we use tempNextTick to satisfy our assumption with MAX_TICK_DISTANCE is set to be matched this condition

		tempNextTick := swapData.nextTick
		if willUpTick && tempNextTick > MaxTickDistance+swapData.currentTick {
			tempNextTick = swapData.currentTick + MaxTickDistance
		} else if !willUpTick && tempNextTick < swapData.currentTick-MaxTickDistance {
			tempNextTick = swapData.currentTick - MaxTickDistance
		}

		swapData.startSqrtP = swapData.sqrtP
		swapData.nextSqrtP, err = utils.GetSqrtRatioAtTick(tempNextTick)
		if err != nil {
			return nil, nil, nil, nil, 0, 0, err
		}

		targetSqrtP := swapData.nextSqrtP
		// ensure next sqrtP (and its corresponding tick) does not exceed price limit
		if willUpTick == (swapData.nextSqrtP.Cmp(limitSqrtP) > 0) {
			targetSqrtP = limitSqrtP
		}

		var usedAmount, returnedAmount, deltaL *big.Int
		usedAmount, returnedAmount, deltaL, swapData.sqrtP, err = utils.ComputeSwapStep(
			new(big.Int).Add(swapData.baseL, swapData.reinvestL),
			swapData.sqrtP,
			targetSqrtP,
			p.Fee,
			swapData.specifiedAmount,
			swapData.isExactInput,
			isToken0,
		)
		if err != nil {
			return nil, nil, nil, nil, 0, 0, err
		}

		swapData.specifiedAmount = new(big.Int).Sub(swapData.specifiedAmount, usedAmount)
		swapData.returnedAmount = new(big.Int).Add(swapData.returnedAmount, returnedAmount)
		swapData.reinvestL = new(big.Int).Add(swapData.reinvestL, deltaL)

		if swapData.sqrtP.Cmp(swapData.nextSqrtP) != 0 {
			if swapData.sqrtP != swapData.startSqrtP {
				swapData.currentTick, err = utils.GetTickAtSqrtRatio(swapData.sqrtP)
				if err != nil {
					return nil, nil, nil, nil, 0, 0, err
				}
			}
			break
		}

		if willUpTick {
			swapData.currentTick = tempNextTick

		} else {
			swapData.currentTick = tempNextTick - 1
		}

		// if tempNextTick is not next initialized tick
		if tempNextTick != swapData.nextTick {
			continue
		}

		swapData.baseL, swapData.nextTick = p._updateLiquidityAndCrossTick(
			swapData.nextTick,
			swapData.baseL,
			willUpTick,
		)
	}

	return swapData.returnedAmount, swapData.baseL, swapData.reinvestL, swapData.sqrtP, swapData.currentTick, swapData.nextTick, nil
}

// Source: https://github.com/KyberNetwork/ks-elastic-sc-v2/blob/3ba84353cbd88f30f222bb9c673e242a2e46fd12/contracts/PoolTicksState.sol#L78-L103
func (p *Pool) _updateLiquidityAndCrossTick(
	nextTick int,
	currentLiquidity *big.Int,
	willUpTick bool,
) (newLiquidity *big.Int, newNextTick int) {
	liquidityNet := p.Ticks[nextTick].LiquidityNet

	if willUpTick {
		newNextTick = p.InitializedTicks[nextTick].Next
	} else {
		newNextTick = p.InitializedTicks[nextTick].Previous
		liquidityNet = new(big.Int).Mul(liquidityNet, constants.NegativeOne)
	}

	var liquidityDelta *big.Int
	if liquidityNet.Cmp(constants.Zero) >= 0 {
		liquidityDelta = liquidityNet
	} else {
		liquidityDelta = new(big.Int).Mul(liquidityNet, constants.NegativeOne)
	}

	newLiquidity = utils.ApplyLiquidityDelta(currentLiquidity, liquidityDelta, liquidityNet.Cmp(constants.Zero) >= 0)

	return newLiquidity, newNextTick
}

// In the contract, this function will mutate the pool state directly
// but in this SDK, we will return a new pool state instead
// because CalcAmountOut is not allowed to mutate state
// instead, we will return the new state to use in the UpdateBalance function
// Source: https://github.com/KyberNetwork/ks-elastic-sc-v2/blob/3ba84353cbd88f30f222bb9c673e242a2e46fd12/contracts/PoolTicksState.sol#L105-L119
func (p *Pool) _updatePoolData(
	baseL *big.Int,
	reinvestL *big.Int,
	sqrtP *big.Int,
	currentTick int,
	nextTick int,
) *Pool {
	var newPoolState Pool

	newPoolState.BaseL = baseL
	newPoolState.ReinvestL = reinvestL
	newPoolState.SqrtP = sqrtP
	newPoolState.CurrentTick = currentTick
	if nextTick > currentTick {
		newPoolState.NearestCurrentTick = p.InitializedTicks[nextTick].Previous
	} else {
		newPoolState.NearestCurrentTick = nextTick
	}

	return &newPoolState
}

func (p *Pool) tickSpacing() int {
	return constants.TickSpacings[p.Fee]
}
