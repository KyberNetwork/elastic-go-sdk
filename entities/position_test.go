package entities

import (
	"math/big"
	"testing"

	"github.com/daoleno/uniswap-sdk-core/entities"
	"github.com/ethereum/go-ethereum/common"
	"github.com/shopspring/decimal"
	"github.com/stretchr/testify/assert"

	"github.com/KyberNetwork/promm-sdk-go/constants"
	"github.com/KyberNetwork/promm-sdk-go/utils"
)

var (
	B100e6  = big.NewInt(100e6)
	B100e12 = big.NewInt(100e12)
	B100e18 = decimal.NewFromBigInt(big.NewInt(100), 18).BigInt()
)

func initPool() (*Pool, int, int) {
	USDC := entities.NewToken(1, common.HexToAddress("0xA0b86991c6218b36c1d19D4a2e9Eb0cE3606eB48"), 6, "USDC", "USD Coin")
	DAI := entities.NewToken(1, common.HexToAddress("0x6B175474E89094C44Da98b954EedeAC495271d0F"), 18, "DAI", "DAI Stablecoin")
	poolSqrtRatioStart := utils.EncodeSqrtRatioX96(B100e6, B100e18)
	poolTickCurrent, err := utils.GetTickAtSqrtRatio(poolSqrtRatioStart)
	if err != nil {
		panic(err)
	}

	tickSpacing := constants.TickSpacings[constants.FeeMedium]
	p, err := NewPool(DAI, USDC, constants.FeeMedium, poolSqrtRatioStart, big.NewInt(0), big.NewInt(0), poolTickCurrent, nil)
	if err != nil {
		panic(err)
	}

	return p, poolTickCurrent, tickSpacing
}

func TestPosition(t *testing.T) {
	DAIUSDCPool, _, tickSpacing := initPool()

	// can be constructed around 0 tick
	p, err := NewPosition(DAIUSDCPool, big.NewInt(1), -8, 8)
	assert.NoError(t, err)
	assert.Equal(t, big.NewInt(1), p.Liquidity)

	// can use min and max ticks
	p, err = NewPosition(DAIUSDCPool, big.NewInt(1), NearestUsableTick(utils.MinTick, tickSpacing), 8)
	assert.NoError(t, err)
	assert.Equal(t, big.NewInt(1), p.Liquidity)

	// tick lower must be less than tick upper
	_, err = NewPosition(DAIUSDCPool, big.NewInt(1), 10, -10)
	assert.ErrorIs(t, err, ErrTickOrder)

	// tick lower cannot equal tick upper
	_, err = NewPosition(DAIUSDCPool, big.NewInt(1), -10, -10)
	assert.ErrorIs(t, err, ErrTickOrder)

	// tick lower must be multiple of tick spacing
	_, err = NewPosition(DAIUSDCPool, big.NewInt(1), -5, 10)
	assert.ErrorIs(t, err, ErrTickLower)

	// tick lower must be greater than MIN_TICK
	_, err = NewPosition(DAIUSDCPool, big.NewInt(1), NearestUsableTick(utils.MinTick, tickSpacing)-tickSpacing, 10)
	assert.ErrorIs(t, err, ErrTickLower)

	// tick upper must be multiple of tick spacing
	_, err = NewPosition(DAIUSDCPool, big.NewInt(1), -8, 17)
	assert.ErrorIs(t, err, ErrTickUpper)

	// tick upper must be less than MAX_TICK
	_, err = NewPosition(DAIUSDCPool, big.NewInt(1), -8, NearestUsableTick(utils.MaxTick, tickSpacing)+tickSpacing)
	assert.ErrorIs(t, err, ErrTickUpper)
}

func TestAmount0(t *testing.T) {
	DAIUSDCPool, poolTickCurrent, tickSpacing := initPool()

	// is correct for price above
	p, err := NewPosition(DAIUSDCPool, B100e12, NearestUsableTick(poolTickCurrent, tickSpacing)+tickSpacing, NearestUsableTick(poolTickCurrent, tickSpacing)+tickSpacing*2)
	assert.NoError(t, err)
	amount0, err := p.Amount0()
	assert.NoError(t, err)
	assert.Equal(t, "39981952346063638", amount0.Quotient().String())

	// is correct for price below
	p, err = NewPosition(DAIUSDCPool, B100e18, NearestUsableTick(poolTickCurrent, tickSpacing)-tickSpacing*2, NearestUsableTick(poolTickCurrent, tickSpacing)-tickSpacing)
	assert.NoError(t, err)
	amount0, err = p.Amount0()
	assert.NoError(t, err)
	assert.Equal(t, "0", amount0.Quotient().String())

	// is correct for in-range position
	p, err = NewPosition(DAIUSDCPool, B100e18, NearestUsableTick(poolTickCurrent, tickSpacing)-tickSpacing*2, NearestUsableTick(poolTickCurrent, tickSpacing)+tickSpacing*2)
	assert.NoError(t, err)
	amount0, err = p.Amount0()
	assert.NoError(t, err)
	assert.Equal(t, "60111117597136795831849", amount0.Quotient().String())
	// ! 120054069145287995769396 in v3-sdk(typescript)
}

func TestAmount1(t *testing.T) {
	DAIUSDCPool, poolTickCurrent, tickSpacing := initPool()

	// is correct for price above
	p, err := NewPosition(DAIUSDCPool, B100e18, NearestUsableTick(poolTickCurrent, tickSpacing)+tickSpacing, NearestUsableTick(poolTickCurrent, tickSpacing)+tickSpacing*2)
	assert.NoError(t, err)
	amount1, err := p.Amount1()
	assert.NoError(t, err)
	assert.Equal(t, "0", amount1.Quotient().String())

	// is correct for price below
	p, err = NewPosition(DAIUSDCPool, B100e18, NearestUsableTick(poolTickCurrent, tickSpacing)-tickSpacing*2, NearestUsableTick(poolTickCurrent, tickSpacing)-tickSpacing)
	assert.NoError(t, err)
	amount1, err = p.Amount1()
	assert.NoError(t, err)
	assert.Equal(t, "39966069225", amount1.Quotient().String())

	// is correct for in-range position
	p, err = NewPosition(DAIUSDCPool, B100e18, NearestUsableTick(poolTickCurrent, tickSpacing)-tickSpacing*2, NearestUsableTick(poolTickCurrent, tickSpacing)+tickSpacing*2)
	assert.NoError(t, err)
	amount1, err = p.Amount1()
	assert.NoError(t, err)
	assert.Equal(t, "99812962651", amount1.Quotient().String())
}

func TestMintAmountsWithSlippage(t *testing.T) {
	DAIUSDCPool, poolTickCurrent, tickSpacing := initPool()
	// 0 slippage
	slippageTolerance := entities.NewPercent(big.NewInt(0), big.NewInt(1))

	// is correct for positions below
	p, err := NewPosition(DAIUSDCPool, B100e18,
		NearestUsableTick(poolTickCurrent, tickSpacing)+tickSpacing,
		NearestUsableTick(poolTickCurrent, tickSpacing)+tickSpacing*2)
	assert.NoError(t, err)
	amount0, amount1, err := p.MintAmountsWithSlippage(slippageTolerance)
	assert.NoError(t, err)
	assert.Equal(t, "39981952346063638713107", amount0.String())
	assert.Equal(t, "0", amount1.String())

	// is correct for positions above
	p, err = NewPosition(DAIUSDCPool, B100e18,
		NearestUsableTick(poolTickCurrent, tickSpacing)-tickSpacing*2,
		NearestUsableTick(poolTickCurrent, tickSpacing)-tickSpacing)
	assert.NoError(t, err)
	amount0, amount1, err = p.MintAmountsWithSlippage(slippageTolerance)
	assert.NoError(t, err)
	assert.Equal(t, "0", amount0.String())
	assert.Equal(t, "39966069226", amount1.String())

	// is correct for positions within
	p, err = NewPosition(DAIUSDCPool, B100e18,
		NearestUsableTick(poolTickCurrent, tickSpacing)-tickSpacing*2,
		NearestUsableTick(poolTickCurrent, tickSpacing)+tickSpacing*2)
	assert.NoError(t, err)
	amount0, amount1, err = p.MintAmountsWithSlippage(slippageTolerance)
	assert.NoError(t, err)
	assert.Equal(t, "60111117597136795513260", amount0.String())
	assert.Equal(t, "99812962652", amount1.String())

	// 0.05% slippage
	slippageTolerance = entities.NewPercent(big.NewInt(5), big.NewInt(10000))

	// is correct for positions below
	p, err = NewPosition(DAIUSDCPool, B100e18,
		NearestUsableTick(poolTickCurrent, tickSpacing)+tickSpacing,
		NearestUsableTick(poolTickCurrent, tickSpacing)+tickSpacing*2)
	assert.NoError(t, err)
	amount0, amount1, err = p.MintAmountsWithSlippage(slippageTolerance)
	assert.NoError(t, err)
	assert.Equal(t, "35120488692595011288991", amount0.String())
	assert.Equal(t, "0", amount1.String())

	// is correct for positions above
	p, err = NewPosition(DAIUSDCPool, B100e18,
		NearestUsableTick(poolTickCurrent, tickSpacing)-tickSpacing*2,
		NearestUsableTick(poolTickCurrent, tickSpacing)-tickSpacing)
	assert.NoError(t, err)
	amount0, amount1, err = p.MintAmountsWithSlippage(slippageTolerance)
	assert.NoError(t, err)
	assert.Equal(t, "0", amount0.String())
	assert.Equal(t, "39966069226", amount1.String())

	// is correct for positions within
	p, err = NewPosition(DAIUSDCPool, B100e18,
		NearestUsableTick(poolTickCurrent, tickSpacing)-tickSpacing*2,
		NearestUsableTick(poolTickCurrent, tickSpacing)+tickSpacing*2)
	assert.NoError(t, err)
	amount0, amount1, err = p.MintAmountsWithSlippage(slippageTolerance)
	assert.NoError(t, err)
	assert.Equal(t, "35120488692595011331136", amount0.String())
	assert.Equal(t, "74809836870", amount1.String())

	// 5% slippage tolerance
	slippageTolerance = entities.NewPercent(big.NewInt(5), big.NewInt(100))

	// is correct for pool at min price
	minPricePool, err := NewPool(DAI, USDC, constants.FeeLow, utils.MinSqrtRatio, big.NewInt(0), big.NewInt(0), utils.MinTick, nil)
	assert.NoError(t, err)
	p, err = NewPosition(minPricePool, B100e18,
		NearestUsableTick(poolTickCurrent, tickSpacing)+tickSpacing,
		NearestUsableTick(poolTickCurrent, tickSpacing)+tickSpacing*2)
	assert.NoError(t, err)
	amount0, amount1, err = p.BurnAmountsWithSlippage(slippageTolerance)
	assert.NoError(t, err)
	assert.Equal(t, "39981952346063638972989", amount0.String())
	assert.Equal(t, "0", amount1.String())

	// is correct for pool at max price
	maxPricePool, err := NewPool(DAI, USDC, constants.FeeLow, new(big.Int).Sub(utils.MaxSqrtRatio, big.NewInt(1)), big.NewInt(0), big.NewInt(0), utils.MaxTick-1, nil)
	assert.NoError(t, err)
	p, err = NewPosition(maxPricePool, B100e18,
		NearestUsableTick(poolTickCurrent, tickSpacing)+tickSpacing,
		NearestUsableTick(poolTickCurrent, tickSpacing)+tickSpacing*2)
	assert.NoError(t, err)
	amount0, amount1, err = p.BurnAmountsWithSlippage(slippageTolerance)
	assert.NoError(t, err)
	assert.Equal(t, "0", amount0.String())
	assert.Equal(t, "40014054895", amount1.String())
}

func TestBurnAmountsWithSlippage(t *testing.T) {
	DAIUSDCPool, poolTickCurrent, tickSpacing := initPool()

	// 0 slippage tolerance
	slippageTolerance := entities.NewPercent(big.NewInt(0), big.NewInt(100))

	// is correct for positions below
	p, err := NewPosition(DAIUSDCPool, B100e18,
		NearestUsableTick(poolTickCurrent, tickSpacing)+tickSpacing,
		NearestUsableTick(poolTickCurrent, tickSpacing)+tickSpacing*2)
	assert.NoError(t, err)
	amount0, amount1, err := p.BurnAmountsWithSlippage(slippageTolerance)
	assert.NoError(t, err)
	assert.Equal(t, "39981952346063638972989", amount0.String())
	assert.Equal(t, "0", amount1.String())

	// is correct for positions above
	p, err = NewPosition(DAIUSDCPool, B100e18,
		NearestUsableTick(poolTickCurrent, tickSpacing)-tickSpacing*2,
		NearestUsableTick(poolTickCurrent, tickSpacing)-tickSpacing)
	assert.NoError(t, err)
	amount0, amount1, err = p.BurnAmountsWithSlippage(slippageTolerance)
	assert.NoError(t, err)
	assert.Equal(t, "0", amount0.String())
	assert.Equal(t, "39966069225", amount1.String())

	// is correct for positions within
	p, err = NewPosition(DAIUSDCPool, B100e18,
		NearestUsableTick(poolTickCurrent, tickSpacing)-tickSpacing*2,
		NearestUsableTick(poolTickCurrent, tickSpacing)+tickSpacing*2)
	assert.NoError(t, err)
	amount0, amount1, err = p.BurnAmountsWithSlippage(slippageTolerance)
	assert.NoError(t, err)
	//! 120054069145287995769396 in v3-sdk
	assert.Equal(t, "60111117597136795831849", amount0.String())
	assert.Equal(t, "99812962651", amount1.String())

	// 0.05% slippage
	slippageTolerance = entities.NewPercent(big.NewInt(5), big.NewInt(10000))

	// is correct for positions below
	p, err = NewPosition(DAIUSDCPool, B100e18,
		NearestUsableTick(poolTickCurrent, tickSpacing)+tickSpacing,
		NearestUsableTick(poolTickCurrent, tickSpacing)+tickSpacing*2)
	assert.NoError(t, err)
	amount0, amount1, err = p.BurnAmountsWithSlippage(slippageTolerance)
	assert.NoError(t, err)
	assert.Equal(t, "35120488692595011517274", amount0.String())
	assert.Equal(t, "0", amount1.String())

	// is correct for positions above
	p, err = NewPosition(DAIUSDCPool, B100e18,
		NearestUsableTick(poolTickCurrent, tickSpacing)-tickSpacing*2,
		NearestUsableTick(poolTickCurrent, tickSpacing)-tickSpacing)
	assert.NoError(t, err)
	amount0, amount1, err = p.BurnAmountsWithSlippage(slippageTolerance)
	assert.NoError(t, err)
	assert.Equal(t, "0", amount0.String())
	assert.Equal(t, "39966069225", amount1.String())

	// is correct for positions within
	p, err = NewPosition(DAIUSDCPool, B100e18,
		NearestUsableTick(poolTickCurrent, tickSpacing)-tickSpacing*2,
		NearestUsableTick(poolTickCurrent, tickSpacing)+tickSpacing*2)
	assert.NoError(t, err)
	amount0, amount1, err = p.BurnAmountsWithSlippage(slippageTolerance)
	assert.NoError(t, err)
	// ! 95063440240746211454822 in v3-sdk
	assert.Equal(t, "35120488692595011517274", amount0.String())
	assert.Equal(t, "74809836869", amount1.String())

	// 5% slippage tolerance
	slippageTolerance = entities.NewPercent(big.NewInt(5), big.NewInt(100))

	// is correct for pool at min price
	minPricePool, err := NewPool(DAI, USDC, constants.FeeLow, utils.MinSqrtRatio, big.NewInt(0), big.NewInt(0), utils.MinTick, nil)
	assert.NoError(t, err)
	p, err = NewPosition(minPricePool, B100e18,
		NearestUsableTick(poolTickCurrent, tickSpacing)+tickSpacing,
		NearestUsableTick(poolTickCurrent, tickSpacing)+tickSpacing*2)
	assert.NoError(t, err)
	amount0, amount1, err = p.BurnAmountsWithSlippage(slippageTolerance)
	assert.NoError(t, err)
	// ! 49949961958869841738198 in v3-sdk
	assert.Equal(t, "39981952346063638972989", amount0.String())
	assert.Equal(t, "0", amount1.String())

	// is correct for pool at max price
	maxPricePool, err := NewPool(DAI, USDC, constants.FeeLow, new(big.Int).Sub(utils.MaxSqrtRatio, big.NewInt(1)), big.NewInt(0), big.NewInt(0), utils.MaxTick-1, nil)
	assert.NoError(t, err)
	p, err = NewPosition(maxPricePool, B100e18,
		NearestUsableTick(poolTickCurrent, tickSpacing)+tickSpacing,
		NearestUsableTick(poolTickCurrent, tickSpacing)+tickSpacing*2)
	assert.NoError(t, err)
	amount0, amount1, err = p.BurnAmountsWithSlippage(slippageTolerance)
	assert.NoError(t, err)
	assert.Equal(t, "0", amount0.String())
	// ! 50045084660 in v3-sdk
	assert.Equal(t, "40014054895", amount1.String())
}

func TestMintAmounts(t *testing.T) {
	DAIUSDCPool, poolTickCurrent, tickSpacing := initPool()

	// is correct for price above
	p, err := NewPosition(DAIUSDCPool, B100e18,
		NearestUsableTick(poolTickCurrent, tickSpacing)+tickSpacing,
		NearestUsableTick(poolTickCurrent, tickSpacing)+tickSpacing*2)
	assert.NoError(t, err)
	amount0, amount1, err := p.MintAmounts()
	assert.NoError(t, err)
	assert.Equal(t, "39981952346063638972990", amount0.String())
	assert.Equal(t, "0", amount1.String())

	// is correct for price below
	p, err = NewPosition(DAIUSDCPool, B100e18,
		NearestUsableTick(poolTickCurrent, tickSpacing)-tickSpacing*2,
		NearestUsableTick(poolTickCurrent, tickSpacing)-tickSpacing)
	assert.NoError(t, err)
	amount0, amount1, err = p.MintAmounts()
	assert.NoError(t, err)
	assert.Equal(t, "0", amount0.String())
	assert.Equal(t, "39966069226", amount1.String())

	// is correct for in-range position
	p, err = NewPosition(DAIUSDCPool, B100e18,
		NearestUsableTick(poolTickCurrent, tickSpacing)-tickSpacing*2,
		NearestUsableTick(poolTickCurrent, tickSpacing)+tickSpacing*2)
	assert.NoError(t, err)
	amount0, amount1, err = p.MintAmounts()
	assert.NoError(t, err)
	// note these are rounded up
	assert.Equal(t, "60111117597136795831849", amount0.String())
	assert.Equal(t, "99812962652", amount1.String())
}
