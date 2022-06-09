package entities

import (
	"math/big"
	"testing"

	"github.com/KyberNetwork/uniswap-sdk-core/entities"
	"github.com/ethereum/go-ethereum/common"
	"github.com/stretchr/testify/assert"

	"github.com/KyberNetwork/promm-sdk-go/constants"
	"github.com/KyberNetwork/promm-sdk-go/utils"
)

var (
	USDC = entities.NewToken(
		1, common.HexToAddress("0xA0b86991c6218b36c1d19D4a2e9Eb0cE3606eB48"), 6, "USDC", "USD Coin",
	)
	DAI = entities.NewToken(
		1, common.HexToAddress("0x6B175474E89094C44Da98b954EedeAC495271d0F"), 18, "DAI", "Dai Stablecoin",
	)
	ETHRinkeby = entities.NewToken(
		4, common.HexToAddress("0xc778417E063141139Fce010982780140Aa0cD5Ab"), 18, "ETH", "Ether",
	)
	DAIRinkeby = entities.NewToken(
		4, common.HexToAddress("0xc7AD46e0b8a400Bb3C915120d284AafbA8fc4735"), 18, "DAI", "Dai Stablecoin",
	)
	OneEther = big.NewInt(1e18)
)

func TestNewPool(t *testing.T) {
	_, err := NewPool(
		USDC, entities.WETH9[3], constants.FeeMedium, utils.EncodeSqrtRatioX96(constants.One, constants.One),
		big.NewInt(0), big.NewInt(0), 0, nil,
	)
	assert.ErrorIs(t, err, entities.ErrDifferentChain, "cannot be used for tokens on different chains")

	_, err = NewPool(
		USDC, entities.WETH9[1], 1e6, utils.EncodeSqrtRatioX96(constants.One, constants.One), big.NewInt(0),
		big.NewInt(0), 0, nil,
	)
	assert.ErrorIs(t, err, ErrFeeTooHigh, "fee cannot be more than 1e6'")

	_, err = NewPool(
		USDC, USDC, constants.FeeMedium, utils.EncodeSqrtRatioX96(constants.One, constants.One), big.NewInt(0),
		big.NewInt(0), 0, nil,
	)
	assert.ErrorIs(t, err, entities.ErrSameAddress, "cannot be used for the same token")

	_, err = NewPool(
		USDC, entities.WETH9[1], constants.FeeMedium, utils.EncodeSqrtRatioX96(constants.One, constants.One),
		big.NewInt(0), big.NewInt(0), 1, nil,
	)
	assert.ErrorIs(t, err, ErrInvalidSqrtRatioX96, "price must be within tick price bounds")

	_, err = NewPool(
		USDC, entities.WETH9[1], constants.FeeMedium,
		new(big.Int).Add(utils.EncodeSqrtRatioX96(constants.One, constants.One), big.NewInt(1)), big.NewInt(0),
		big.NewInt(0), -1, nil,
	)
	assert.ErrorIs(t, err, ErrInvalidSqrtRatioX96, "price must be within tick price bounds")

	_, err = NewPool(
		USDC, entities.WETH9[1], constants.FeeMedium, utils.EncodeSqrtRatioX96(constants.One, constants.One),
		big.NewInt(0), big.NewInt(0), 0, nil,
	)
	assert.NoError(t, err, "works with valid arguments for empty pool medium fee")

	_, err = NewPool(
		USDC, entities.WETH9[1], constants.FeeLow, utils.EncodeSqrtRatioX96(constants.One, constants.One),
		big.NewInt(0), big.NewInt(0), 0, nil,
	)
	assert.NoError(t, err, "works with valid arguments for empty pool low fee")

	_, err = NewPool(
		USDC, entities.WETH9[1], constants.FeeHigh, utils.EncodeSqrtRatioX96(constants.One, constants.One),
		big.NewInt(0), big.NewInt(0), 0, nil,
	)
	assert.NoError(t, err, "works with valid arguments for empty pool high fee")
}

func TestGetAddress(t *testing.T) {
	addr, _ := GetAddress(USDC, DAI, constants.FeeLow, "")
	assert.Equal(t, addr, common.HexToAddress("0x6c6Bc977E13Df9b0de53b251522280BB72383700"), "matches an example")
}

func TestToken0(t *testing.T) {
	pool, _ := NewPool(
		USDC, DAI, constants.FeeLow, utils.EncodeSqrtRatioX96(constants.One, constants.One), big.NewInt(0),
		big.NewInt(0), 0, nil,
	)
	assert.Equal(t, pool.Token0, DAI, "always is the token that sorts before")

	pool, _ = NewPool(
		DAI, USDC, constants.FeeLow, utils.EncodeSqrtRatioX96(constants.One, constants.One), big.NewInt(0),
		big.NewInt(0), 0, nil,
	)
	assert.Equal(t, pool.Token0, DAI, "always is the token that sorts before")
}

func TestToken1(t *testing.T) {
	pool, _ := NewPool(
		USDC, DAI, constants.FeeLow, utils.EncodeSqrtRatioX96(constants.One, constants.One), big.NewInt(0),
		big.NewInt(0), 0, nil,
	)
	assert.Equal(t, pool.Token1, USDC, "always is the token that sorts after")

	pool, _ = NewPool(
		DAI, USDC, constants.FeeLow, utils.EncodeSqrtRatioX96(constants.One, constants.One), big.NewInt(0),
		big.NewInt(0), 0, nil,
	)
	assert.Equal(t, pool.Token1, USDC, "always is the token that sorts after")
}

func TestToken0Price(t *testing.T) {
	a1 := new(big.Int).Mul(big.NewInt(101), big.NewInt(1e6))
	a2 := new(big.Int).Mul(big.NewInt(100), big.NewInt(1e18))
	r, _ := utils.GetTickAtSqrtRatio(utils.EncodeSqrtRatioX96(a1, a2))
	pool0, _ := NewPool(
		USDC, DAI, constants.FeeLow, utils.EncodeSqrtRatioX96(a1, a2), big.NewInt(0), big.NewInt(0), r, nil,
	)
	assert.Equal(t, pool0.Token0Price().ToSignificant(5), "1.01", "returns price of token0 in terms of token1")

	pool1, _ := NewPool(
		DAI, USDC, constants.FeeLow, utils.EncodeSqrtRatioX96(a1, a2), big.NewInt(0), big.NewInt(0), r, nil,
	)
	assert.Equal(t, pool1.Token0Price().ToSignificant(5), "1.01", "returns price of token0 in terms of token1")
}

func TestToken1Price(t *testing.T) {
	a1 := new(big.Int).Mul(big.NewInt(101), big.NewInt(1e6))
	a2 := new(big.Int).Mul(big.NewInt(100), big.NewInt(1e18))
	r, _ := utils.GetTickAtSqrtRatio(utils.EncodeSqrtRatioX96(a1, a2))
	pool0, _ := NewPool(
		USDC, DAI, constants.FeeLow, utils.EncodeSqrtRatioX96(a1, a2), big.NewInt(0), big.NewInt(0), r, nil,
	)
	assert.Equal(t, pool0.Token1Price().ToSignificant(5), "0.9901", "returns price of token1 in terms of token0")

	pool1, _ := NewPool(
		DAI, USDC, constants.FeeLow, utils.EncodeSqrtRatioX96(a1, a2), big.NewInt(0), big.NewInt(0), r, nil,
	)
	assert.Equal(t, pool1.Token1Price().ToSignificant(5), "0.9901", "returns price of token1 in terms of token0")
}

func TestPriceOf(t *testing.T) {
	pool, _ := NewPool(
		USDC, DAI, constants.FeeLow, utils.EncodeSqrtRatioX96(constants.One, constants.One), big.NewInt(0),
		big.NewInt(0), 0, nil,
	)
	price0, _ := pool.PriceOf(DAI)
	assert.Equal(t, price0, pool.Token0Price(), "returns price of token in terms of other token")
	price1, _ := pool.PriceOf(USDC)
	assert.Equal(t, price1, pool.Token1Price(), "returns price of token in terms of other token")

	_, err := pool.PriceOf(entities.WETH9[1])
	assert.Error(t, err, "invalid token")
}

func TestChainID(t *testing.T) {
	pool0, _ := NewPool(
		USDC, DAI, constants.FeeLow, utils.EncodeSqrtRatioX96(constants.One, constants.One), big.NewInt(0),
		big.NewInt(0), 0, nil,
	)
	assert.Equal(t, pool0.ChainID(), uint(1), "returns the token0 chainId")

	pool1, _ := NewPool(
		DAI, USDC, constants.FeeLow, utils.EncodeSqrtRatioX96(constants.One, constants.One), big.NewInt(0),
		big.NewInt(0), 0, nil,
	)
	assert.Equal(t, pool1.ChainID(), uint(1), "returns the token0 chainId")
}

func TestInvolvesToken(t *testing.T) {
	pool, _ := NewPool(
		USDC, DAI, 50, utils.EncodeSqrtRatioX96(constants.One, constants.One), big.NewInt(0),
		big.NewInt(0), 0, nil,
	)
	assert.True(t, pool.InvolvesToken(USDC), "involves USDC")
	assert.True(t, pool.InvolvesToken(DAI), "involves DAI")
	assert.False(t, pool.InvolvesToken(entities.WETH9[1]), "does not involve WETH9")
}

func newTestPool() *Pool {
	ticks := []Tick{
		{
			Index:          NearestUsableTick(utils.MinTick, 8),
			LiquidityNet:   OneEther,
			LiquidityGross: OneEther,
		},
		{
			Index:          NearestUsableTick(utils.MaxTick, 8),
			LiquidityNet:   new(big.Int).Mul(OneEther, constants.NegativeOne),
			LiquidityGross: OneEther,
		},
	}

	p, err := NewTickListDataProvider(ticks, 8)
	if err != nil {
		panic(err)
	}

	pool, err := NewPool(
		USDC, DAI, 50, utils.EncodeSqrtRatioX96(constants.One, constants.One), OneEther, big.NewInt(0),
		0, p,
	)
	if err != nil {
		panic(err)
	}
	return pool
}

func newTestPool2() *Pool {
	liquidityNet1, _ := new(big.Int).SetString("120714648802705550448", 10)
	liquidityGross1, _ := new(big.Int).SetString("120714648802705550448", 10)
	liquidityNet2, _ := new(big.Int).SetString("-120714648802705550448", 10)
	liquidityGross2, _ := new(big.Int).SetString("120714648802705550448", 10)
	ticks := []Tick{
		{
			Index:          62160,
			LiquidityNet:   liquidityNet1,
			LiquidityGross: liquidityGross1,
		},
		{
			Index:          92160,
			LiquidityNet:   liquidityNet2,
			LiquidityGross: liquidityGross2,
		},
	}

	p, err := NewTickListDataProvider(ticks, constants.TickSpacings[constants.FeeMedium])
	if err != nil {
		panic(err)
	}
	price, _ := new(big.Int).SetString("4317840471017651404712833792646", 10)
	liquidity, _ := new(big.Int).SetString("120714648802705550448", 10)
	reinvestL, _ := new(big.Int).SetString("81785081063693", 10)

	pool, err := NewPool(ETHRinkeby, DAIRinkeby, constants.FeeMedium, price, liquidity, reinvestL, 79967, p)
	if err != nil {
		panic(err)
	}
	return pool
}

func TestGetOutputAmount(t *testing.T) {
	pool := newTestPool()

	// USDC -> DAI
	inputAmount := entities.FromRawAmount(USDC.Currency, big.NewInt(1000000))
	outputAmount, _, err := pool.GetOutputAmount(inputAmount, nil)
	if err != nil {
		t.Fatal(err)
	}
	assert.True(t, outputAmount.Currency.Equal(DAI.Currency))
	assert.Equal(t, big.NewInt(999499), outputAmount.Quotient())

	// DAI -> USDC
	inputAmount = entities.FromRawAmount(DAI.Currency, big.NewInt(24295310180196433))
	outputAmount, _, err = pool.GetOutputAmount(inputAmount, nil)
	if err != nil {
		t.Fatal(err)
	}
	assert.True(t, outputAmount.Currency.Equal(USDC.Currency))
	assert.Equal(t, big.NewInt(23707188978482194), outputAmount.Quotient())

	pool = newTestPool2()

	// ETH -> DAI
	inputAmount = entities.FromRawAmount(ETHRinkeby.Currency, big.NewInt(1000000000000000000))
	outputAmount, _, err = pool.GetOutputAmount(inputAmount, nil)
	if err != nil {
		t.Fatal(err)
	}
	assert.True(t, outputAmount.Currency.Equal(DAIRinkeby.Currency))

	expectValue, _ := new(big.Int).SetString("2045603787129768717773", 10)
	assert.Equal(t, expectValue, outputAmount.Quotient())
}

func TestGetInputAmount(t *testing.T) {
	pool := newTestPool()

	// USDC -> DAI
	outputAmount := entities.FromRawAmount(DAI.Currency, big.NewInt(98))
	inputAmount, _, err := pool.GetInputAmount(outputAmount, nil)
	if err != nil {
		t.Fatal(err)
	}
	assert.True(t, inputAmount.Currency.Equal(USDC.Currency))
	assert.Equal(t, inputAmount.Quotient(), big.NewInt(100))

	// DAI -> USDC
	outputAmount = entities.FromRawAmount(USDC.Currency, big.NewInt(98))
	inputAmount, _, err = pool.GetInputAmount(outputAmount, nil)
	if err != nil {
		t.Fatal(err)
	}
	assert.True(t, inputAmount.Currency.Equal(DAI.Currency))
	assert.Equal(t, inputAmount.Quotient(), big.NewInt(100))
}
