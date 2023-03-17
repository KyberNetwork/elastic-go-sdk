package entities

import (
	"math/big"
	"testing"

	"github.com/daoleno/uniswap-sdk-core/entities"
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
		USDC, entities.WETH9[3], constants.Fee004, utils.EncodeSqrtRatioX96(constants.One, constants.One),
		big.NewInt(0), big.NewInt(0), 0, nil,
	)
	assert.ErrorIs(t, err, entities.ErrDifferentChain, "cannot be used for tokens on different chains")

	_, err = NewPool(
		USDC, entities.WETH9[1], 1e6, utils.EncodeSqrtRatioX96(constants.One, constants.One), big.NewInt(0),
		big.NewInt(0), 0, nil,
	)
	assert.ErrorIs(t, err, ErrFeeTooHigh, "fee cannot be more than 1e6'")

	_, err = NewPool(
		USDC, USDC, constants.Fee004, utils.EncodeSqrtRatioX96(constants.One, constants.One), big.NewInt(0),
		big.NewInt(0), 0, nil,
	)
	assert.ErrorIs(t, err, entities.ErrSameAddress, "cannot be used for the same token")

	_, err = NewPool(
		USDC, entities.WETH9[1], constants.Fee004, utils.EncodeSqrtRatioX96(constants.One, constants.One),
		big.NewInt(0), big.NewInt(0), 1, nil,
	)
	assert.ErrorIs(t, err, ErrInvalidSqrtRatioX96, "price must be within tick price bounds")

	_, err = NewPool(
		USDC, entities.WETH9[1], constants.Fee004,
		new(big.Int).Add(utils.EncodeSqrtRatioX96(constants.One, constants.One), big.NewInt(1)), big.NewInt(0),
		big.NewInt(0), -1, nil,
	)
	assert.ErrorIs(t, err, ErrInvalidSqrtRatioX96, "price must be within tick price bounds")

	_, err = NewPool(
		USDC, entities.WETH9[1], constants.Fee0008, utils.EncodeSqrtRatioX96(constants.One, constants.One),
		big.NewInt(0), big.NewInt(0), 0, nil,
	)
	assert.NoError(t, err, "works with valid arguments for empty pool 0.008%")

	_, err = NewPool(
		USDC, entities.WETH9[1], constants.Fee001, utils.EncodeSqrtRatioX96(constants.One, constants.One),
		big.NewInt(0), big.NewInt(0), 0, nil,
	)
	assert.NoError(t, err, "works with valid arguments for empty pool 0.01%")

	_, err = NewPool(
		USDC, entities.WETH9[1], constants.Fee002, utils.EncodeSqrtRatioX96(constants.One, constants.One),
		big.NewInt(0), big.NewInt(0), 0, nil,
	)
	assert.NoError(t, err, "works with valid arguments for empty pool 0.02%")

	_, err = NewPool(
		USDC, entities.WETH9[1], constants.Fee004, utils.EncodeSqrtRatioX96(constants.One, constants.One),
		big.NewInt(0), big.NewInt(0), 0, nil,
	)
	assert.NoError(t, err, "works with valid arguments for empty pool 0.04%")

	_, err = NewPool(
		USDC, entities.WETH9[1], constants.Fee01, utils.EncodeSqrtRatioX96(constants.One, constants.One),
		big.NewInt(0), big.NewInt(0), 0, nil,
	)
	assert.NoError(t, err, "works with valid arguments for empty pool 0.1%")

	_, err = NewPool(
		USDC, entities.WETH9[1], constants.Fee025, utils.EncodeSqrtRatioX96(constants.One, constants.One),
		big.NewInt(0), big.NewInt(0), 0, nil,
	)
	assert.NoError(t, err, "works with valid arguments for empty pool 0.25%")

	_, err = NewPool(
		USDC, entities.WETH9[1], constants.Fee03, utils.EncodeSqrtRatioX96(constants.One, constants.One),
		big.NewInt(0), big.NewInt(0), 0, nil,
	)
	assert.NoError(t, err, "works with valid arguments for empty pool 0.3%")

	_, err = NewPool(
		USDC, entities.WETH9[1], constants.Fee1, utils.EncodeSqrtRatioX96(constants.One, constants.One),
		big.NewInt(0), big.NewInt(0), 0, nil,
	)
	assert.NoError(t, err, "works with valid arguments for empty pool 1%")

	_, err = NewPool(
		USDC, entities.WETH9[1], constants.Fee2, utils.EncodeSqrtRatioX96(constants.One, constants.One),
		big.NewInt(0), big.NewInt(0), 0, nil,
	)
	assert.NoError(t, err, "works with valid arguments for empty pool 2%")

	_, err = NewPool(
		USDC, entities.WETH9[1], constants.Fee5, utils.EncodeSqrtRatioX96(constants.One, constants.One),
		big.NewInt(0), big.NewInt(0), 0, nil,
	)
	assert.NoError(t, err, "works with valid arguments for empty pool 5%")
}

func TestGetAddress(t *testing.T) {
	addr, _ := GetAddress(USDC, DAI, constants.Fee001, "")
	assert.Equal(t, addr, common.HexToAddress("0xE5e30b9aDD54E8E6DDf05b76693ad690fEe56a25"), "matches an example")
}

func TestToken0(t *testing.T) {
	pool, _ := NewPool(
		USDC, DAI, constants.Fee001, utils.EncodeSqrtRatioX96(constants.One, constants.One), big.NewInt(0),
		big.NewInt(0), 0, nil,
	)
	assert.Equal(t, pool.Token0, DAI, "always is the token that sorts before")

	pool, _ = NewPool(
		DAI, USDC, constants.Fee001, utils.EncodeSqrtRatioX96(constants.One, constants.One), big.NewInt(0),
		big.NewInt(0), 0, nil,
	)
	assert.Equal(t, pool.Token0, DAI, "always is the token that sorts before")
}

func TestToken1(t *testing.T) {
	pool, _ := NewPool(
		USDC, DAI, constants.Fee001, utils.EncodeSqrtRatioX96(constants.One, constants.One), big.NewInt(0),
		big.NewInt(0), 0, nil,
	)
	assert.Equal(t, pool.Token1, USDC, "always is the token that sorts after")

	pool, _ = NewPool(
		DAI, USDC, constants.Fee001, utils.EncodeSqrtRatioX96(constants.One, constants.One), big.NewInt(0),
		big.NewInt(0), 0, nil,
	)
	assert.Equal(t, pool.Token1, USDC, "always is the token that sorts after")
}

func TestToken0Price(t *testing.T) {
	a1 := new(big.Int).Mul(big.NewInt(101), big.NewInt(1e6))
	a2 := new(big.Int).Mul(big.NewInt(100), big.NewInt(1e18))
	r, _ := utils.GetTickAtSqrtRatio(utils.EncodeSqrtRatioX96(a1, a2))
	pool0, _ := NewPool(
		USDC, DAI, constants.Fee001, utils.EncodeSqrtRatioX96(a1, a2), big.NewInt(0), big.NewInt(0), r, nil,
	)
	assert.Equal(t, pool0.Token0Price().ToSignificant(5), "1.01", "returns price of token0 in terms of token1")

	pool1, _ := NewPool(
		DAI, USDC, constants.Fee001, utils.EncodeSqrtRatioX96(a1, a2), big.NewInt(0), big.NewInt(0), r, nil,
	)
	assert.Equal(t, pool1.Token0Price().ToSignificant(5), "1.01", "returns price of token0 in terms of token1")
}

func TestToken1Price(t *testing.T) {
	a1 := new(big.Int).Mul(big.NewInt(101), big.NewInt(1e6))
	a2 := new(big.Int).Mul(big.NewInt(100), big.NewInt(1e18))
	r, _ := utils.GetTickAtSqrtRatio(utils.EncodeSqrtRatioX96(a1, a2))
	pool0, _ := NewPool(
		USDC, DAI, constants.Fee001, utils.EncodeSqrtRatioX96(a1, a2), big.NewInt(0), big.NewInt(0), r, nil,
	)
	assert.Equal(t, pool0.Token1Price().ToSignificant(5), "0.9901", "returns price of token1 in terms of token0")

	pool1, _ := NewPool(
		DAI, USDC, constants.Fee001, utils.EncodeSqrtRatioX96(a1, a2), big.NewInt(0), big.NewInt(0), r, nil,
	)
	assert.Equal(t, pool1.Token1Price().ToSignificant(5), "0.9901", "returns price of token1 in terms of token0")
}

func TestPriceOf(t *testing.T) {
	pool, _ := NewPool(
		USDC, DAI, constants.Fee001, utils.EncodeSqrtRatioX96(constants.One, constants.One), big.NewInt(0),
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
		USDC, DAI, constants.Fee001, utils.EncodeSqrtRatioX96(constants.One, constants.One), big.NewInt(0),
		big.NewInt(0), 0, nil,
	)
	assert.Equal(t, pool0.ChainID(), uint(1), "returns the token0 chainId")

	pool1, _ := NewPool(
		DAI, USDC, constants.Fee001, utils.EncodeSqrtRatioX96(constants.One, constants.One), big.NewInt(0),
		big.NewInt(0), 0, nil,
	)
	assert.Equal(t, pool1.ChainID(), uint(1), "returns the token0 chainId")
}

func TestInvolvesToken(t *testing.T) {
	pool, _ := NewPool(
		USDC, DAI, constants.Fee01, utils.EncodeSqrtRatioX96(constants.One, constants.One), big.NewInt(0),
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

func newTestPoolFee0008() *Pool {
	ticks := []Tick{
		{
			Index:          NearestUsableTick(utils.MinTick, constants.TickSpacings[constants.Fee0008]),
			LiquidityNet:   OneEther,
			LiquidityGross: OneEther,
		},
		{
			Index:          NearestUsableTick(utils.MaxTick, constants.TickSpacings[constants.Fee0008]),
			LiquidityNet:   new(big.Int).Mul(OneEther, constants.NegativeOne),
			LiquidityGross: OneEther,
		},
	}

	p, err := NewTickListDataProvider(ticks, constants.TickSpacings[constants.Fee0008])
	if err != nil {
		panic(err)
	}

	pool, err := NewPool(
		USDC, DAI, constants.Fee0008, utils.EncodeSqrtRatioX96(constants.One, constants.One), OneEther, big.NewInt(0),
		0, p,
	)
	if err != nil {
		panic(err)
	}

	return pool
}

func newTestPoolFee001() *Pool {
	ticks := []Tick{
		{
			Index:          NearestUsableTick(utils.MinTick, constants.TickSpacings[constants.Fee001]),
			LiquidityNet:   OneEther,
			LiquidityGross: OneEther,
		},
		{
			Index:          NearestUsableTick(utils.MaxTick, constants.TickSpacings[constants.Fee001]),
			LiquidityNet:   new(big.Int).Mul(OneEther, constants.NegativeOne),
			LiquidityGross: OneEther,
		},
	}

	p, err := NewTickListDataProvider(ticks, constants.TickSpacings[constants.Fee001])
	if err != nil {
		panic(err)
	}

	pool, err := NewPool(
		USDC, DAI, constants.Fee001, utils.EncodeSqrtRatioX96(constants.One, constants.One), OneEther, big.NewInt(0),
		0, p,
	)
	if err != nil {
		panic(err)
	}

	return pool
}

func newTestPoolFee002() *Pool {
	ticks := []Tick{
		{
			Index:          NearestUsableTick(utils.MinTick, constants.TickSpacings[constants.Fee002]),
			LiquidityNet:   OneEther,
			LiquidityGross: OneEther,
		},
		{
			Index:          NearestUsableTick(utils.MaxTick, constants.TickSpacings[constants.Fee002]),
			LiquidityNet:   new(big.Int).Mul(OneEther, constants.NegativeOne),
			LiquidityGross: OneEther,
		},
	}

	p, err := NewTickListDataProvider(ticks, constants.TickSpacings[constants.Fee002])
	if err != nil {
		panic(err)
	}

	pool, err := NewPool(
		USDC, DAI, constants.Fee002, utils.EncodeSqrtRatioX96(constants.One, constants.One), OneEther, big.NewInt(0),
		0, p,
	)
	if err != nil {
		panic(err)
	}

	return pool
}

func newTestPoolFee004() *Pool {
	ticks := []Tick{
		{
			Index:          NearestUsableTick(utils.MinTick, constants.TickSpacings[constants.Fee004]),
			LiquidityNet:   OneEther,
			LiquidityGross: OneEther,
		},
		{
			Index:          NearestUsableTick(utils.MaxTick, constants.TickSpacings[constants.Fee004]),
			LiquidityNet:   new(big.Int).Mul(OneEther, constants.NegativeOne),
			LiquidityGross: OneEther,
		},
	}

	p, err := NewTickListDataProvider(ticks, constants.TickSpacings[constants.Fee004])
	if err != nil {
		panic(err)
	}

	pool, err := NewPool(
		USDC, DAI, constants.Fee004, utils.EncodeSqrtRatioX96(constants.One, constants.One), OneEther, big.NewInt(0),
		0, p,
	)
	if err != nil {
		panic(err)
	}

	return pool
}

func newTestPoolFee01() *Pool {
	ticks := []Tick{
		{
			Index:          NearestUsableTick(utils.MinTick, constants.TickSpacings[constants.Fee01]),
			LiquidityNet:   OneEther,
			LiquidityGross: OneEther,
		},
		{
			Index:          NearestUsableTick(utils.MaxTick, constants.TickSpacings[constants.Fee01]),
			LiquidityNet:   new(big.Int).Mul(OneEther, constants.NegativeOne),
			LiquidityGross: OneEther,
		},
	}

	p, err := NewTickListDataProvider(ticks, constants.TickSpacings[constants.Fee01])
	if err != nil {
		panic(err)
	}

	pool, err := NewPool(
		USDC, DAI, constants.Fee01, utils.EncodeSqrtRatioX96(constants.One, constants.One), OneEther, big.NewInt(0),
		0, p,
	)
	if err != nil {
		panic(err)
	}

	return pool
}

func newTestPoolFee025() *Pool {
	ticks := []Tick{
		{
			Index:          NearestUsableTick(utils.MinTick, constants.TickSpacings[constants.Fee025]),
			LiquidityNet:   OneEther,
			LiquidityGross: OneEther,
		},
		{
			Index:          NearestUsableTick(utils.MaxTick, constants.TickSpacings[constants.Fee025]),
			LiquidityNet:   new(big.Int).Mul(OneEther, constants.NegativeOne),
			LiquidityGross: OneEther,
		},
	}

	p, err := NewTickListDataProvider(ticks, constants.TickSpacings[constants.Fee025])
	if err != nil {
		panic(err)
	}

	pool, err := NewPool(
		USDC, DAI, constants.Fee025, utils.EncodeSqrtRatioX96(constants.One, constants.One), OneEther, big.NewInt(0),
		0, p,
	)
	if err != nil {
		panic(err)
	}

	return pool
}

func newTestPoolFee03() *Pool {
	ticks := []Tick{
		{
			Index:          NearestUsableTick(utils.MinTick, constants.TickSpacings[constants.Fee03]),
			LiquidityNet:   OneEther,
			LiquidityGross: OneEther,
		},
		{
			Index:          NearestUsableTick(utils.MaxTick, constants.TickSpacings[constants.Fee03]),
			LiquidityNet:   new(big.Int).Mul(OneEther, constants.NegativeOne),
			LiquidityGross: OneEther,
		},
	}

	p, err := NewTickListDataProvider(ticks, constants.TickSpacings[constants.Fee03])
	if err != nil {
		panic(err)
	}

	pool, err := NewPool(
		USDC, DAI, constants.Fee03, utils.EncodeSqrtRatioX96(constants.One, constants.One), OneEther, big.NewInt(0),
		0, p,
	)
	if err != nil {
		panic(err)
	}

	return pool
}

func newTestPoolFee1() *Pool {
	ticks := []Tick{
		{
			Index:          NearestUsableTick(utils.MinTick, constants.TickSpacings[constants.Fee1]),
			LiquidityNet:   OneEther,
			LiquidityGross: OneEther,
		},
		{
			Index:          NearestUsableTick(utils.MaxTick, constants.TickSpacings[constants.Fee1]),
			LiquidityNet:   new(big.Int).Mul(OneEther, constants.NegativeOne),
			LiquidityGross: OneEther,
		},
	}

	p, err := NewTickListDataProvider(ticks, constants.TickSpacings[constants.Fee1])
	if err != nil {
		panic(err)
	}

	pool, err := NewPool(
		USDC, DAI, constants.Fee1, utils.EncodeSqrtRatioX96(constants.One, constants.One), OneEther, big.NewInt(0),
		0, p,
	)
	if err != nil {
		panic(err)
	}

	return pool
}

func newTestPoolFee2() *Pool {
	ticks := []Tick{
		{
			Index:          NearestUsableTick(utils.MinTick, constants.TickSpacings[constants.Fee2]),
			LiquidityNet:   OneEther,
			LiquidityGross: OneEther,
		},
		{
			Index:          NearestUsableTick(utils.MaxTick, constants.TickSpacings[constants.Fee2]),
			LiquidityNet:   new(big.Int).Mul(OneEther, constants.NegativeOne),
			LiquidityGross: OneEther,
		},
	}

	p, err := NewTickListDataProvider(ticks, constants.TickSpacings[constants.Fee2])
	if err != nil {
		panic(err)
	}

	pool, err := NewPool(
		USDC, DAI, constants.Fee2, utils.EncodeSqrtRatioX96(constants.One, constants.One), OneEther, big.NewInt(0),
		0, p,
	)
	if err != nil {
		panic(err)
	}

	return pool
}

func newTestPoolFee5() *Pool {
	ticks := []Tick{
		{
			Index:          NearestUsableTick(utils.MinTick, constants.TickSpacings[constants.Fee5]),
			LiquidityNet:   OneEther,
			LiquidityGross: OneEther,
		},
		{
			Index:          NearestUsableTick(utils.MaxTick, constants.TickSpacings[constants.Fee5]),
			LiquidityNet:   new(big.Int).Mul(OneEther, constants.NegativeOne),
			LiquidityGross: OneEther,
		},
	}

	p, err := NewTickListDataProvider(ticks, constants.TickSpacings[constants.Fee5])
	if err != nil {
		panic(err)
	}

	pool, err := NewPool(
		USDC, DAI, constants.Fee5, utils.EncodeSqrtRatioX96(constants.One, constants.One), OneEther, big.NewInt(0),
		0, p,
	)
	if err != nil {
		panic(err)
	}

	return pool
}

func TestPool_GetOutputAmount(t *testing.T) {
	expectValueFee0008USDCDAI, _ := new(big.Int).SetString("999919", 10)
	expectValueFee0008DAIUSDC, _ := new(big.Int).SetString("23717151023641933", 10)
	expectValueFee001USDCDAI, _ := new(big.Int).SetString("999899", 10)
	expectValueFee001DAIUSDC, _ := new(big.Int).SetString("23716676643996285", 10)
	expectValueFee002USDCDAI, _ := new(big.Int).SetString("999799", 10)
	expectValueFee002DAIUSDC, _ := new(big.Int).SetString("23714304740582796", 10)
	expectValueFee004USDCDAI, _ := new(big.Int).SetString("999599", 10)
	expectValueFee004DAIUSDC, _ := new(big.Int).SetString("23709560907826659", 10)
	expectValueFee01USDCDAI, _ := new(big.Int).SetString("998999", 10)
	expectValueFee01DAIUSDC, _ := new(big.Int).SetString("23695329346163752", 10)
	expectValueFee025USDCDAI, _ := new(big.Int).SetString("997499", 10)
	expectValueFee025DAIUSDC, _ := new(big.Int).SetString("23659750017012196", 10)
	expectValueFee03USDCDAI, _ := new(big.Int).SetString("996999", 10)
	expectValueFee03DAIUSDC, _ := new(big.Int).SetString("23647890096562934", 10)
	expectValueFee1USDCDAI, _ := new(big.Int).SetString("989999", 10)
	expectValueFee1DAIUSDC, _ := new(big.Int).SetString("23481843646839215", 10)
	expectValueFee2USDCDAI, _ := new(big.Int).SetString("979999", 10)
	expectValueFee2DAIUSDC, _ := new(big.Int).SetString("23244609941828430", 10)
	expectValueFee5USDCDAI, _ := new(big.Int).SetString("949999", 10)
	expectValueFee5DAIUSDC, _ := new(big.Int).SetString("22532735948303642", 10)

	type args struct {
		inputAmount       *entities.CurrencyAmount
		sqrtPriceLimitX96 *big.Int
	}
	tests := []struct {
		name               string
		pool               *Pool
		args               args
		expectCurrencyOut  *entities.Token
		expectOutputAmount *big.Int
		expectErr          error
	}{
		{
			name: "it should return correct amount for pool with fee 0.008%",
			pool: newTestPoolFee0008(),
			args: args{
				inputAmount:       entities.FromRawAmount(USDC, big.NewInt(1000000)),
				sqrtPriceLimitX96: nil,
			},
			expectCurrencyOut:  DAI,
			expectOutputAmount: expectValueFee0008USDCDAI,
			expectErr:          nil,
		},
		{
			name: "it should return correct amount for pool with fee 0.008%",
			pool: newTestPoolFee0008(),
			args: args{
				inputAmount:       entities.FromRawAmount(DAI, big.NewInt(24295310180196433)),
				sqrtPriceLimitX96: nil,
			},
			expectCurrencyOut:  USDC,
			expectOutputAmount: expectValueFee0008DAIUSDC,
			expectErr:          nil,
		},
		{
			name: "it should return correct amount for pool with fee 0.01%",
			pool: newTestPoolFee001(),
			args: args{
				inputAmount:       entities.FromRawAmount(USDC, big.NewInt(1000000)),
				sqrtPriceLimitX96: nil,
			},
			expectCurrencyOut:  DAI,
			expectOutputAmount: expectValueFee001USDCDAI,
			expectErr:          nil,
		},
		{
			name: "it should return correct amount for pool with fee 0.01%",
			pool: newTestPoolFee001(),
			args: args{
				inputAmount:       entities.FromRawAmount(DAI, big.NewInt(24295310180196433)),
				sqrtPriceLimitX96: nil,
			},
			expectCurrencyOut:  USDC,
			expectOutputAmount: expectValueFee001DAIUSDC,
			expectErr:          nil,
		},
		{
			name: "it should return correct amount for pool with fee 0.02%",
			pool: newTestPoolFee002(),
			args: args{
				inputAmount:       entities.FromRawAmount(USDC, big.NewInt(1000000)),
				sqrtPriceLimitX96: nil,
			},
			expectCurrencyOut:  DAI,
			expectOutputAmount: expectValueFee002USDCDAI,
			expectErr:          nil,
		},
		{
			name: "it should return correct amount for pool with fee 0.02%",
			pool: newTestPoolFee002(),
			args: args{
				inputAmount:       entities.FromRawAmount(DAI, big.NewInt(24295310180196433)),
				sqrtPriceLimitX96: nil,
			},
			expectCurrencyOut:  USDC,
			expectOutputAmount: expectValueFee002DAIUSDC,
			expectErr:          nil,
		},
		{
			name: "it should return correct amount for pool with fee 0.04%",
			pool: newTestPoolFee004(),
			args: args{
				inputAmount:       entities.FromRawAmount(USDC, big.NewInt(1000000)),
				sqrtPriceLimitX96: nil,
			},
			expectCurrencyOut:  DAI,
			expectOutputAmount: expectValueFee004USDCDAI,
			expectErr:          nil,
		},
		{
			name: "it should return correct amount for pool with fee 0.04%",
			pool: newTestPoolFee004(),
			args: args{
				inputAmount:       entities.FromRawAmount(DAI, big.NewInt(24295310180196433)),
				sqrtPriceLimitX96: nil,
			},
			expectCurrencyOut:  USDC,
			expectOutputAmount: expectValueFee004DAIUSDC,
			expectErr:          nil,
		},
		{
			name: "it should return correct amount for pool with fee 0.1%",
			pool: newTestPoolFee01(),
			args: args{
				inputAmount:       entities.FromRawAmount(USDC, big.NewInt(1000000)),
				sqrtPriceLimitX96: nil,
			},
			expectCurrencyOut:  DAI,
			expectOutputAmount: expectValueFee01USDCDAI,
			expectErr:          nil,
		},
		{
			name: "it should return correct amount for pool with fee 0.1%",
			pool: newTestPoolFee01(),
			args: args{
				inputAmount:       entities.FromRawAmount(DAI, big.NewInt(24295310180196433)),
				sqrtPriceLimitX96: nil,
			},
			expectCurrencyOut:  USDC,
			expectOutputAmount: expectValueFee01DAIUSDC,
			expectErr:          nil,
		},
		{
			name: "it should return correct amount for pool with fee 0.25%",
			pool: newTestPoolFee025(),
			args: args{
				inputAmount:       entities.FromRawAmount(USDC, big.NewInt(1000000)),
				sqrtPriceLimitX96: nil,
			},
			expectCurrencyOut:  DAI,
			expectOutputAmount: expectValueFee025USDCDAI,
			expectErr:          nil,
		},
		{
			name: "it should return correct amount for pool with fee 0.25%",
			pool: newTestPoolFee025(),
			args: args{
				inputAmount:       entities.FromRawAmount(DAI, big.NewInt(24295310180196433)),
				sqrtPriceLimitX96: nil,
			},
			expectCurrencyOut:  USDC,
			expectOutputAmount: expectValueFee025DAIUSDC,
			expectErr:          nil,
		},
		{
			name: "it should return correct amount for pool with fee 0.3%",
			pool: newTestPoolFee03(),
			args: args{
				inputAmount:       entities.FromRawAmount(USDC, big.NewInt(1000000)),
				sqrtPriceLimitX96: nil,
			},
			expectCurrencyOut:  DAI,
			expectOutputAmount: expectValueFee03USDCDAI,
			expectErr:          nil,
		},
		{
			name: "it should return correct amount for pool with fee 0.3%",
			pool: newTestPoolFee03(),
			args: args{
				inputAmount:       entities.FromRawAmount(DAI, big.NewInt(24295310180196433)),
				sqrtPriceLimitX96: nil,
			},
			expectCurrencyOut:  USDC,
			expectOutputAmount: expectValueFee03DAIUSDC,
			expectErr:          nil,
		},
		{
			name: "it should return correct amount for pool with fee 1%",
			pool: newTestPoolFee1(),
			args: args{
				inputAmount:       entities.FromRawAmount(USDC, big.NewInt(1000000)),
				sqrtPriceLimitX96: nil,
			},
			expectCurrencyOut:  DAI,
			expectOutputAmount: expectValueFee1USDCDAI,
			expectErr:          nil,
		},
		{
			name: "it should return correct amount for pool with fee 1%",
			pool: newTestPoolFee1(),
			args: args{
				inputAmount:       entities.FromRawAmount(DAI, big.NewInt(24295310180196433)),
				sqrtPriceLimitX96: nil,
			},
			expectCurrencyOut:  USDC,
			expectOutputAmount: expectValueFee1DAIUSDC,
			expectErr:          nil,
		},
		{
			name: "it should return correct amount for pool with fee 2%",
			pool: newTestPoolFee2(),
			args: args{
				inputAmount:       entities.FromRawAmount(USDC, big.NewInt(1000000)),
				sqrtPriceLimitX96: nil,
			},
			expectCurrencyOut:  DAI,
			expectOutputAmount: expectValueFee2USDCDAI,
			expectErr:          nil,
		},
		{
			name: "it should return correct amount for pool with fee 2%",
			pool: newTestPoolFee2(),
			args: args{
				inputAmount:       entities.FromRawAmount(DAI, big.NewInt(24295310180196433)),
				sqrtPriceLimitX96: nil,
			},
			expectCurrencyOut:  USDC,
			expectOutputAmount: expectValueFee2DAIUSDC,
			expectErr:          nil,
		},
		{
			name: "it should return correct amount for pool with fee 5%",
			pool: newTestPoolFee5(),
			args: args{
				inputAmount:       entities.FromRawAmount(USDC, big.NewInt(1000000)),
				sqrtPriceLimitX96: nil,
			},
			expectCurrencyOut:  DAI,
			expectOutputAmount: expectValueFee5USDCDAI,
			expectErr:          nil,
		},
		{
			name: "it should return correct amount for pool with fee 5%",
			pool: newTestPoolFee5(),
			args: args{
				inputAmount:       entities.FromRawAmount(DAI, big.NewInt(24295310180196433)),
				sqrtPriceLimitX96: nil,
			},
			expectCurrencyOut:  USDC,
			expectOutputAmount: expectValueFee5DAIUSDC,
			expectErr:          nil,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			outputAmount, _, err := tt.pool.GetOutputAmount(tt.args.inputAmount, tt.args.sqrtPriceLimitX96)

			assert.ErrorIs(t, err, tt.expectErr)
			assert.True(t, outputAmount.Currency.Equal(tt.expectCurrencyOut))
			assert.Equal(t, tt.expectOutputAmount, outputAmount.Quotient())
		})
	}
}

func TestGetInputAmount(t *testing.T) {
	pool := newTestPool()

	// USDC -> DAI
	outputAmount := entities.FromRawAmount(DAI, big.NewInt(98))
	inputAmount, _, err := pool.GetInputAmount(outputAmount, nil)
	if err != nil {
		t.Fatal(err)
	}
	assert.True(t, inputAmount.Currency.Equal(USDC))
	assert.Equal(t, inputAmount.Quotient(), big.NewInt(98))

	// DAI -> USDC
	outputAmount = entities.FromRawAmount(USDC, big.NewInt(98))
	inputAmount, _, err = pool.GetInputAmount(outputAmount, nil)
	if err != nil {
		t.Fatal(err)
	}
	assert.True(t, inputAmount.Currency.Equal(DAI))
	assert.Equal(t, inputAmount.Quotient(), big.NewInt(98))
}
