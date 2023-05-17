package entities

import (
	"math/big"
	"testing"

	"github.com/daoleno/uniswap-sdk-core/entities"
	"github.com/ethereum/go-ethereum/common"
	"github.com/stretchr/testify/assert"

	"github.com/KyberNetwork/elastic-go-sdk/v2/constants"
	"github.com/KyberNetwork/elastic-go-sdk/v2/utils"
)

var (
	Ether  = entities.EtherOnChain(1)
	token0 = entities.NewToken(1, common.HexToAddress("0x0000000000000000000000000000000000000001"), 18, "t0", "token0")
	token1 = entities.NewToken(1, common.HexToAddress("0x0000000000000000000000000000000000000002"), 18, "t1", "token1")
	token2 = entities.NewToken(1, common.HexToAddress("0x0000000000000000000000000000000000000003"), 18, "t2", "token2")
	token3 = entities.NewToken(1, common.HexToAddress("0x0000000000000000000000000000000000000004"), 18, "t3", "token3")

	pool_0_1 = v2StylePool(
		token0,
		token1,
		entities.FromRawAmount(token0, big.NewInt(100000)),
		entities.FromRawAmount(token1, big.NewInt(100000)),
		constants.Fee004,
	)
	pool_0_2 = v2StylePool(
		token0,
		token2,
		entities.FromRawAmount(token0, big.NewInt(100000)),
		entities.FromRawAmount(token2, big.NewInt(110000)),
		constants.Fee004,
	)
	pool_0_3 = v2StylePool(
		token0,
		token3,
		entities.FromRawAmount(token0, big.NewInt(100000)),
		entities.FromRawAmount(token3, big.NewInt(90000)),
		constants.Fee004,
	)
	pool_1_2 = v2StylePool(
		token1,
		token2,
		entities.FromRawAmount(token1, big.NewInt(120000)),
		entities.FromRawAmount(token2, big.NewInt(100000)),
		constants.Fee004,
	)
	pool_1_3 = v2StylePool(
		token1,
		token3,
		entities.FromRawAmount(token1, big.NewInt(120000)),
		entities.FromRawAmount(token3, big.NewInt(130000)),
		constants.Fee004,
	)
	pool_weth_0 = v2StylePool(
		entities.WETH9[1],
		token0,
		entities.FromRawAmount(entities.WETH9[1], big.NewInt(100000)),
		entities.FromRawAmount(token0, big.NewInt(100000)),
		constants.Fee004,
	)
	pool_weth_1 = v2StylePool(
		entities.WETH9[1],
		token1,
		entities.FromRawAmount(entities.WETH9[1], big.NewInt(100000)),
		entities.FromRawAmount(token1, big.NewInt(100000)),
		constants.Fee004,
	)
	pool_weth_2 = v2StylePool(
		entities.WETH9[1],
		token2,
		entities.FromRawAmount(entities.WETH9[1], big.NewInt(100000)),
		entities.FromRawAmount(token2, big.NewInt(100000)),
		constants.Fee004,
	)
)

func v2StylePool(token0, token1 *entities.Token, reserve0, reserve1 *entities.CurrencyAmount, feeAmount constants.FeeAmount) *Pool {
	sqrtRatioX96 := utils.EncodeSqrtRatioX96(reserve1.Quotient(), reserve0.Quotient())
	liquidity := new(big.Int).Sqrt(new(big.Int).Mul(reserve0.Quotient(), reserve1.Quotient()))
	ticks := []Tick{
		{
			Index:          NearestUsableTick(utils.MinTick, constants.TickSpacings[feeAmount]),
			LiquidityNet:   liquidity,
			LiquidityGross: liquidity,
		},
		{
			Index:          NearestUsableTick(utils.MaxTick, constants.TickSpacings[feeAmount]),
			LiquidityNet:   new(big.Int).Mul(liquidity, big.NewInt(-1)),
			LiquidityGross: liquidity,
		},
	}
	s, err := utils.GetTickAtSqrtRatio(sqrtRatioX96)
	if err != nil {
		panic(err)
	}
	p, err := NewTickListDataProvider(ticks, constants.TickSpacings[feeAmount])
	if err != nil {
		panic(err)
	}
	pool, err := NewPool(token0, token1, feeAmount, sqrtRatioX96, liquidity, big.NewInt(0), s, p)
	if err != nil {
		panic(err)
	}
	return pool
}

func TestFromRoute(t *testing.T) {
	// can be constructed with ETHER as input'
	r, _ := NewRoute([]*Pool{pool_weth_0}, Ether, token0)
	trade, err := FromRoute(r, entities.FromRawAmount(Ether, big.NewInt(10000)), entities.ExactInput)
	if err != nil {
		t.Fatal(err)
	}
	assert.Equal(t, trade.InputAmount().Currency, Ether)
	assert.Equal(t, trade.OutputAmount().Currency, token0)

	// can be constructed with ETHER as input for exact output
	r, _ = NewRoute([]*Pool{pool_weth_0}, Ether, token0)
	trade, _ = FromRoute(r, entities.FromRawAmount(token0, big.NewInt(10000)), entities.ExactOutput)
	assert.Equal(t, trade.InputAmount().Currency, Ether)
	assert.Equal(t, trade.OutputAmount().Currency, token0)

	// can be constructed with ETHER as output
	r, _ = NewRoute([]*Pool{pool_weth_0}, token0, Ether)
	trade, err = FromRoute(r, entities.FromRawAmount(Ether, big.NewInt(10000)), entities.ExactOutput)
	if err != nil {
		t.Fatal(err)
	}
	assert.Equal(t, trade.InputAmount().Currency, token0)
	assert.Equal(t, trade.OutputAmount().Currency, Ether)

	// can be constructed with ETHER as output for exact input
	r, _ = NewRoute([]*Pool{pool_weth_0}, token0, Ether)
	trade, err = FromRoute(r, entities.FromRawAmount(token0, big.NewInt(10000)), entities.ExactInput)
	if err != nil {
		t.Fatal(err)
	}
	assert.Equal(t, trade.InputAmount().Currency, token0)
	assert.Equal(t, trade.OutputAmount().Currency, Ether)
}

func TestFromRoutes(t *testing.T) {
	// can be constructed with ETHER as input with multiple routes
	r, _ := NewRoute([]*Pool{pool_weth_0}, Ether, token0)
	trade, err := FromRoutes([]*WrappedRoute{{Amount: entities.FromRawAmount(Ether, big.NewInt(10000)), Route: r}}, entities.ExactInput)
	if err != nil {
		t.Fatal(err)
	}
	assert.Equal(t, trade.InputAmount().Currency, Ether)
	assert.Equal(t, trade.OutputAmount().Currency, token0)

	// can be constructed with ETHER as input for exact output with multiple routes
	r0, _ := NewRoute([]*Pool{pool_weth_0}, Ether, token0)
	r1, _ := NewRoute([]*Pool{pool_weth_1, pool_0_1}, Ether, token0)
	trade, err = FromRoutes([]*WrappedRoute{
		{Amount: entities.FromRawAmount(token0, big.NewInt(3000)), Route: r0},
		{Amount: entities.FromRawAmount(token0, big.NewInt(7000)), Route: r1},
	}, entities.ExactOutput)
	if err != nil {
		t.Fatal(err)
	}
	assert.Equal(t, trade.InputAmount().Currency, Ether)
	assert.Equal(t, trade.OutputAmount().Currency, token0)

	// can be constructed with ETHER as output with multiple routes
	r0, _ = NewRoute([]*Pool{pool_weth_0}, token0, Ether)
	r1, _ = NewRoute([]*Pool{pool_0_1, pool_weth_1}, token0, Ether)
	trade, err = FromRoutes([]*WrappedRoute{
		{Amount: entities.FromRawAmount(Ether, big.NewInt(4000)), Route: r0},
		{Amount: entities.FromRawAmount(Ether, big.NewInt(6000)), Route: r1},
	}, entities.ExactOutput)
	if err != nil {
		t.Fatal(err)
	}
	assert.Equal(t, trade.InputAmount().Currency, token0)
	assert.Equal(t, trade.OutputAmount().Currency, Ether)

	// can be constructed with ETHER as output for exact input with multiple routes
	r0, _ = NewRoute([]*Pool{pool_weth_0}, token0, Ether)
	r1, _ = NewRoute([]*Pool{pool_0_1, pool_weth_1}, token0, Ether)
	trade, err = FromRoutes([]*WrappedRoute{
		{Amount: entities.FromRawAmount(token0, big.NewInt(3000)), Route: r0},
		{Amount: entities.FromRawAmount(token0, big.NewInt(7000)), Route: r1},
	}, entities.ExactInput)
	if err != nil {
		t.Fatal(err)
	}
	assert.Equal(t, trade.InputAmount().Currency, token0)
	assert.Equal(t, trade.OutputAmount().Currency, Ether)

	// errors if pools are re-used between routes
	r0, _ = NewRoute([]*Pool{pool_0_1, pool_weth_1}, token0, Ether)
	r1, _ = NewRoute([]*Pool{pool_0_1, pool_1_2, pool_weth_2}, token0, Ether)
	_, err = FromRoutes([]*WrappedRoute{
		{Amount: entities.FromRawAmount(token0, big.NewInt(4500)), Route: r0},
		{Amount: entities.FromRawAmount(token0, big.NewInt(5500)), Route: r1},
	}, entities.ExactInput)
	assert.ErrorIs(t, err, ErrDuplicatePools)
}

func TestCreateUncheckedTrade(t *testing.T) {
	r, _ := NewRoute([]*Pool{pool_0_1}, token0, token1)
	_, err := CreateUncheckedTrade(r, entities.FromRawAmount(token2, big.NewInt(10000)), entities.FromRawAmount(token1, big.NewInt(10000)), entities.ExactInput)
	assert.ErrorIs(t, err, ErrInputCurrencyMismatch, "if input currency does not match route")

	_, err = CreateUncheckedTrade(r, entities.FromRawAmount(token0, big.NewInt(10000)), entities.FromRawAmount(token2, big.NewInt(10000)), entities.ExactOutput)
	assert.ErrorIs(t, err, ErrOutputCurrencyMismatch, "if output currency does not match route")

	_, err = CreateUncheckedTrade(r, entities.FromRawAmount(token0, big.NewInt(10000)), entities.FromRawAmount(token1, big.NewInt(100000)), entities.ExactInput)
	assert.NoError(t, err, "can create an exact input trade without simulating")

	_, err = CreateUncheckedTrade(r, entities.FromRawAmount(token0, big.NewInt(10000)), entities.FromRawAmount(token1, big.NewInt(100000)), entities.ExactOutput)
	assert.NoError(t, err, "can create an exact output trade without simulating")
}

func TestCreateUncheckedTradeWithMultipleRoutes(t *testing.T) {
	r0, _ := NewRoute([]*Pool{pool_1_2}, token2, token1)
	s0 := &Swap{
		Route:        r0,
		InputAmount:  entities.FromRawAmount(token2, big.NewInt(2000)),
		OutputAmount: entities.FromRawAmount(token1, big.NewInt(2000)),
	}
	r1, _ := NewRoute([]*Pool{pool_0_1}, token0, token1)
	s1 := &Swap{
		Route:        r1,
		InputAmount:  entities.FromRawAmount(token2, big.NewInt(8000)),
		OutputAmount: entities.FromRawAmount(token1, big.NewInt(8000)),
	}
	_, err := CreateUncheckedTradeWithMultipleRoutes([]*Swap{s0, s1}, entities.ExactInput)
	assert.ErrorIs(t, err, ErrInputCurrencyMismatch, "if input currency does not match route with multiple routes")

	r0, _ = NewRoute([]*Pool{pool_0_2}, token0, token2)
	s0 = &Swap{
		Route:        r0,
		InputAmount:  entities.FromRawAmount(token0, big.NewInt(10000)),
		OutputAmount: entities.FromRawAmount(token2, big.NewInt(10000)),
	}
	r1, _ = NewRoute([]*Pool{pool_0_1}, token0, token1)
	s1 = &Swap{
		Route:        r1,
		InputAmount:  entities.FromRawAmount(token0, big.NewInt(10000)),
		OutputAmount: entities.FromRawAmount(token2, big.NewInt(10000)),
	}
	_, err = CreateUncheckedTradeWithMultipleRoutes([]*Swap{s0, s1}, entities.ExactInput)
	assert.ErrorIs(t, err, ErrOutputCurrencyMismatch, "if output currency does not match route with multiple routes")

	r0, _ = NewRoute([]*Pool{pool_0_1}, token0, token1)
	s0 = &Swap{
		Route:        r0,
		InputAmount:  entities.FromRawAmount(token0, big.NewInt(5000)),
		OutputAmount: entities.FromRawAmount(token1, big.NewInt(50000)),
	}
	r1, _ = NewRoute([]*Pool{pool_0_2, pool_1_2}, token0, token1)
	s1 = &Swap{
		Route:        r1,
		InputAmount:  entities.FromRawAmount(token0, big.NewInt(5000)),
		OutputAmount: entities.FromRawAmount(token1, big.NewInt(50000)),
	}
	_, err = CreateUncheckedTradeWithMultipleRoutes([]*Swap{s0, s1}, entities.ExactInput)
	assert.NoError(t, err, "can create an exact input trade without simulating with multiple routes")

	r0, _ = NewRoute([]*Pool{pool_0_1}, token0, token1)
	s0 = &Swap{
		Route:        r0,
		InputAmount:  entities.FromRawAmount(token0, big.NewInt(5001)),
		OutputAmount: entities.FromRawAmount(token1, big.NewInt(50000)),
	}
	r1, _ = NewRoute([]*Pool{pool_0_2, pool_1_2}, token0, token1)
	s1 = &Swap{
		Route:        r1,
		InputAmount:  entities.FromRawAmount(token0, big.NewInt(4999)),
		OutputAmount: entities.FromRawAmount(token1, big.NewInt(50000)),
	}
	_, err = CreateUncheckedTradeWithMultipleRoutes([]*Swap{s0, s1}, entities.ExactOutput)
	assert.NoError(t, err, "can create an exact output trade without simulating with multiple routes")
}

func TestRouteSwaps(t *testing.T) {
	r, _ := NewRoute([]*Pool{pool_0_1, pool_1_2}, token0, token2)
	singleRoute, _ := CreateUncheckedTrade(r, entities.FromRawAmount(token0, big.NewInt(100)), entities.FromRawAmount(token2, big.NewInt(69)), entities.ExactInput)

	r0, _ := NewRoute([]*Pool{pool_0_1, pool_1_2}, token0, token2)
	s1 := &Swap{
		Route:        r0,
		InputAmount:  entities.FromRawAmount(token0, big.NewInt(50)),
		OutputAmount: entities.FromRawAmount(token2, big.NewInt(35)),
	}
	r1, _ := NewRoute([]*Pool{pool_0_2}, token0, token2)
	s2 := &Swap{
		Route:        r1,
		InputAmount:  entities.FromRawAmount(token0, big.NewInt(50)),
		OutputAmount: entities.FromRawAmount(token2, big.NewInt(34)),
	}
	multiRoute, _ := CreateUncheckedTradeWithMultipleRoutes([]*Swap{s1, s2}, entities.ExactInput)

	// can access route for single route trade if less than 0
	_, err := singleRoute.Route()
	assert.NoError(t, err, "can access route for single route trade if less than 0")

	// can access routes for both single and multi route trades
	assert.Equal(t, len(singleRoute.Swaps), 1, "can access routes for single route trades")
	assert.Equal(t, len(multiRoute.Swaps), 2, "can access routes for multi route trades")

	_, err = multiRoute.Route()
	assert.ErrorIs(t, err, ErrTradeHasMultipleRoutes, "if access route on multi route trade")
}

func TestWorstExecutionPrice(t *testing.T) {
	r0, _ := NewRoute([]*Pool{pool_0_1, pool_1_2}, token0, token2)
	r1, _ := NewRoute([]*Pool{pool_0_2}, token0, token2)

	// tradeType = EXACT_INPUT
	exactIn, _ := CreateUncheckedTrade(r0, entities.FromRawAmount(token0, big.NewInt(100)), entities.FromRawAmount(token2, big.NewInt(69)), entities.ExactInput)
	exactInMultiRoute, _ := CreateUncheckedTradeWithMultipleRoutes([]*Swap{
		{
			Route:        r0,
			InputAmount:  entities.FromRawAmount(token0, big.NewInt(50)),
			OutputAmount: entities.FromRawAmount(token2, big.NewInt(35)),
		},
		{
			Route:        r1,
			InputAmount:  entities.FromRawAmount(token0, big.NewInt(50)),
			OutputAmount: entities.FromRawAmount(token2, big.NewInt(34)),
		},
	}, entities.ExactInput)

	_, err := exactIn.MinimumAmountOut(entities.NewPercent(big.NewInt(-1), big.NewInt(100)), nil)
	assert.ErrorIs(t, err, ErrInvalidSlippageTolerance, "throws if less than 0")

	price, _ := exactIn.WorstExecutionPrice(entities.NewPercent(big.NewInt(0), big.NewInt(100)))
	assert.Equal(t, price, exactIn.ExecutionPrice(), "returns exact if 0")

	price, _ = exactIn.WorstExecutionPrice(entities.NewPercent(big.NewInt(0), big.NewInt(100)))
	assert.Equal(t, price, entities.NewPrice(token0, token2, big.NewInt(100), big.NewInt(69)), "returns exact if nonzero")
	price, _ = exactIn.WorstExecutionPrice(entities.NewPercent(big.NewInt(5), big.NewInt(100)))
	assert.Equal(t, price, entities.NewPrice(token0, token2, big.NewInt(100), big.NewInt(65)), "returns exact if nonzero")
	price, _ = exactIn.WorstExecutionPrice(entities.NewPercent(big.NewInt(200), big.NewInt(100)))
	assert.Equal(t, price, entities.NewPrice(token0, token2, big.NewInt(100), big.NewInt(23)), "returns exact if nonzero")

	price, _ = exactInMultiRoute.WorstExecutionPrice(entities.NewPercent(big.NewInt(0), big.NewInt(100)))
	assert.Equal(t, price, entities.NewPrice(token0, token2, big.NewInt(100), big.NewInt(69)), "returns exact if nonzero with multiple routes")
	price, _ = exactInMultiRoute.WorstExecutionPrice(entities.NewPercent(big.NewInt(5), big.NewInt(100)))
	assert.Equal(t, price, entities.NewPrice(token0, token2, big.NewInt(100), big.NewInt(65)), "returns exact if nonzero with multiple routes")
	price, _ = exactInMultiRoute.WorstExecutionPrice(entities.NewPercent(big.NewInt(200), big.NewInt(100)))
	assert.Equal(t, price, entities.NewPrice(token0, token2, big.NewInt(100), big.NewInt(23)), "returns exact if nonzero with multiple routes")

	// tradeType = EXACT_OUTPUT
	exactOut, _ := CreateUncheckedTrade(r0, entities.FromRawAmount(token0, big.NewInt(156)), entities.FromRawAmount(token2, big.NewInt(100)), entities.ExactOutput)
	exactOutMultiRoute, _ := CreateUncheckedTradeWithMultipleRoutes([]*Swap{
		{
			Route:        r0,
			InputAmount:  entities.FromRawAmount(token0, big.NewInt(78)),
			OutputAmount: entities.FromRawAmount(token2, big.NewInt(50)),
		},
		{
			Route:        r1,
			InputAmount:  entities.FromRawAmount(token0, big.NewInt(78)),
			OutputAmount: entities.FromRawAmount(token2, big.NewInt(50)),
		},
	}, entities.ExactOutput)

	_, err = exactOut.WorstExecutionPrice(entities.NewPercent(big.NewInt(-1), big.NewInt(100)))
	assert.ErrorIs(t, err, ErrInvalidSlippageTolerance, "throws if less than 0")

	price, _ = exactOut.WorstExecutionPrice(entities.NewPercent(big.NewInt(0), big.NewInt(100)))
	assert.Equal(t, price, exactOut.ExecutionPrice(), "returns exact if 0")

	price, _ = exactOut.WorstExecutionPrice(entities.NewPercent(big.NewInt(0), big.NewInt(100)))
	assert.True(t, price.EqualTo(entities.NewPrice(token0, token2, big.NewInt(156), big.NewInt(100)).Fraction), "returns slippage amount if nonzero")
	price, _ = exactOut.WorstExecutionPrice(entities.NewPercent(big.NewInt(5), big.NewInt(100)))
	assert.True(t, price.EqualTo(entities.NewPrice(token0, token2, big.NewInt(163), big.NewInt(100)).Fraction), "returns slippage amount if nonzero")
	price, _ = exactOut.WorstExecutionPrice(entities.NewPercent(big.NewInt(200), big.NewInt(100)))
	assert.True(t, price.EqualTo(entities.NewPrice(token0, token2, big.NewInt(468), big.NewInt(100)).Fraction), "returns slippage amount if nonzero")

	price, _ = exactOutMultiRoute.WorstExecutionPrice(entities.NewPercent(big.NewInt(0), big.NewInt(100)))
	assert.True(t, price.EqualTo(entities.NewPrice(token0, token2, big.NewInt(156), big.NewInt(100)).Fraction), "returns exact if nonzero with multiple routes")
	price, _ = exactOutMultiRoute.WorstExecutionPrice(entities.NewPercent(big.NewInt(5), big.NewInt(100)))
	assert.True(t, price.EqualTo(entities.NewPrice(token0, token2, big.NewInt(163), big.NewInt(100)).Fraction), "returns exact if nonzero with multiple routes")
	price, _ = exactOutMultiRoute.WorstExecutionPrice(entities.NewPercent(big.NewInt(200), big.NewInt(100)))
	assert.True(t, price.EqualTo(entities.NewPrice(token0, token2, big.NewInt(468), big.NewInt(100)).Fraction), "returns exact if nonzero with multiple routes")
}

func TestPriceImpact(t *testing.T) {
	r0, _ := NewRoute([]*Pool{pool_0_1, pool_1_2}, token0, token2)
	r1, _ := NewRoute([]*Pool{pool_0_2}, token0, token2)

	// tradeType = EXACT_INPUT
	exactIn, _ := CreateUncheckedTradeWithMultipleRoutes([]*Swap{
		{
			Route:        r0,
			InputAmount:  entities.FromRawAmount(token0, big.NewInt(100)),
			OutputAmount: entities.FromRawAmount(token2, big.NewInt(69)),
		}}, entities.ExactInput)
	exactInMultiRoute, _ := CreateUncheckedTradeWithMultipleRoutes([]*Swap{
		{
			Route:        r0,
			InputAmount:  entities.FromRawAmount(token0, big.NewInt(90)),
			OutputAmount: entities.FromRawAmount(token2, big.NewInt(62)),
		},
		{
			Route:        r1,
			InputAmount:  entities.FromRawAmount(token0, big.NewInt(10)),
			OutputAmount: entities.FromRawAmount(token2, big.NewInt(7)),
		},
	}, entities.ExactInput)

	impact, _ := exactIn.PriceImpact()
	assert.Equal(t, impact.ToSignificant(3), "17.2", "is correct")
	mimpact, _ := exactInMultiRoute.PriceImpact()
	assert.Equal(t, mimpact.ToSignificant(3), "19.8", "is correct")

	// tradeType = EXACT_OUTPUT
	exactOut, _ := CreateUncheckedTradeWithMultipleRoutes([]*Swap{
		{
			Route:        r0,
			InputAmount:  entities.FromRawAmount(token0, big.NewInt(156)),
			OutputAmount: entities.FromRawAmount(token2, big.NewInt(100)),
		}}, entities.ExactOutput)
	exactOutMultiRoute, _ := CreateUncheckedTradeWithMultipleRoutes([]*Swap{
		{
			Route:        r0,
			InputAmount:  entities.FromRawAmount(token0, big.NewInt(140)),
			OutputAmount: entities.FromRawAmount(token2, big.NewInt(90)),
		},
		{
			Route:        r1,
			InputAmount:  entities.FromRawAmount(token0, big.NewInt(16)),
			OutputAmount: entities.FromRawAmount(token2, big.NewInt(10)),
		},
	}, entities.ExactOutput)

	impact, _ = exactOut.PriceImpact()
	assert.Equal(t, impact.ToSignificant(3), "23.1", "is correct")
	mimpact, _ = exactOutMultiRoute.PriceImpact()
	assert.Equal(t, mimpact.ToSignificant(3), "25.5", "is correct")
}

func TestBestTradeExactIn(t *testing.T) {
	_, err := BestTradeExactIn(nil, entities.FromRawAmount(token0, big.NewInt(10000)), nil, nil, nil, nil, nil)
	assert.ErrorIs(t, err, ErrNoPools, "throws with empty pools")

	_, err = BestTradeExactIn([]*Pool{pool_0_2}, entities.FromRawAmount(token0, big.NewInt(10000)), token2, &BestTradeOptions{MaxHops: 0}, nil, nil, nil)
	assert.ErrorIs(t, err, ErrInvalidMaxHops, "throws with max hops of 0")

	// provides best route
	result, err := BestTradeExactIn([]*Pool{pool_0_1, pool_0_2, pool_1_2}, entities.FromRawAmount(token0, big.NewInt(10000)), token2, nil, nil, nil, nil)
	if err != nil {
		t.Fatal(err)
	}
	assert.Equal(t, len(result), 2)
	assert.Equal(t, len(result[0].Swaps[0].Route.Pools), 1)
	assert.Equal(t, result[0].Swaps[0].Route.TokenPath, []*entities.Token{token0, token2})
	assert.True(t, result[0].InputAmount().EqualTo(entities.FromRawAmount(token0, big.NewInt(10000)).Fraction))
	assert.True(t, result[0].OutputAmount().EqualTo(entities.FromRawAmount(token2, big.NewInt(9999)).Fraction))
	assert.Equal(t, len(result[1].Swaps[0].Route.Pools), 2)
	assert.Equal(t, result[1].Swaps[0].Route.TokenPath, []*entities.Token{token0, token1, token2})
	assert.True(t, result[1].InputAmount().EqualTo(entities.FromRawAmount(token0, big.NewInt(10000)).Fraction))
	assert.True(t, result[1].OutputAmount().EqualTo(entities.FromRawAmount(token2, big.NewInt(7039)).Fraction))

	// respects maxHops
	result, err = BestTradeExactIn([]*Pool{pool_0_1, pool_0_2, pool_1_2}, entities.FromRawAmount(token0, big.NewInt(10)), token2, &BestTradeOptions{MaxNumResults: 3, MaxHops: 1}, nil, nil, nil)
	if err != nil {
		t.Fatal(err)
	}
	assert.Equal(t, len(result), 1)
	assert.Equal(t, len(result[0].Swaps[0].Route.Pools), 1)

	// insufficient input for one pool
	result, err = BestTradeExactIn([]*Pool{pool_0_1, pool_0_2, pool_1_2}, entities.FromRawAmount(token0, big.NewInt(1)), token2, nil, nil, nil, nil)
	if err != nil {
		t.Fatal(err)
	}
	assert.Equal(t, len(result), 2)
	assert.Equal(t, len(result[0].Swaps[0].Route.Pools), 1)
	assert.Equal(t, result[0].Swaps[0].Route.TokenPath, []*entities.Token{token0, token2})
	assert.True(t, result[0].OutputAmount().EqualTo(entities.FromRawAmount(token2, big.NewInt(1)).Fraction))

	// respects n
	result, err = BestTradeExactIn([]*Pool{pool_0_1, pool_0_2, pool_1_2}, entities.FromRawAmount(token0, big.NewInt(10)), token2, &BestTradeOptions{MaxNumResults: 1, MaxHops: 3}, nil, nil, nil)
	if err != nil {
		t.Fatal(err)
	}
	assert.Equal(t, len(result), 1)

	// no path
	result, err = BestTradeExactIn([]*Pool{pool_0_1, pool_0_3, pool_1_3}, entities.FromRawAmount(token0, big.NewInt(10)), token2, nil, nil, nil, nil)
	if err != nil {
		t.Fatal(err)
	}
	assert.Equal(t, len(result), 0)

	// works for ETHER currency input
	result, err = BestTradeExactIn([]*Pool{pool_weth_0, pool_0_1, pool_0_3, pool_1_3}, entities.FromRawAmount(Ether, big.NewInt(100)), token3, nil, nil, nil, nil)
	if err != nil {
		t.Fatal(err)
	}
	assert.Equal(t, len(result), 2)
	assert.Equal(t, result[0].InputAmount().Currency, Ether)
	assert.Equal(t, result[0].Swaps[0].Route.TokenPath, []*entities.Token{entities.WETH9[1], token0, token1, token3})
	assert.Equal(t, result[0].OutputAmount().Currency, token3)
	assert.Equal(t, result[1].InputAmount().Currency, Ether)
	assert.Equal(t, result[1].Swaps[0].Route.TokenPath, []*entities.Token{entities.WETH9[1], token0, token3})
	assert.Equal(t, result[1].OutputAmount().Currency, token3)

	// works for ETHER currency output
	result, err = BestTradeExactIn([]*Pool{pool_weth_0, pool_0_1, pool_0_3, pool_1_3}, entities.FromRawAmount(token3, big.NewInt(100)), Ether, nil, nil, nil, nil)
	if err != nil {
		t.Fatal(err)
	}
	assert.Equal(t, len(result), 2)
	assert.Equal(t, result[0].InputAmount().Currency, token3)
	assert.Equal(t, result[0].Swaps[0].Route.TokenPath, []*entities.Token{token3, token0, entities.WETH9[1]})
	assert.Equal(t, result[0].OutputAmount().Currency, Ether)
	assert.Equal(t, result[1].InputAmount().Currency, token3)
	assert.Equal(t, result[1].Swaps[0].Route.TokenPath, []*entities.Token{token3, token1, token0, entities.WETH9[1]})
	assert.Equal(t, result[1].OutputAmount().Currency, Ether)
}

func TestMaximumAmountIn(t *testing.T) {
	r, _ := NewRoute([]*Pool{pool_0_1, pool_1_2}, token0, token2)

	// tradeType = EXACT_INPUT
	exactIn, _ := FromRoute(r, entities.FromRawAmount(token0, big.NewInt(100)), entities.ExactInput)

	_, err := exactIn.MaximumAmountIn(entities.NewPercent(big.NewInt(-1), big.NewInt(100)), nil)
	assert.ErrorIs(t, err, ErrInvalidSlippageTolerance, "throws if less than 0")

	amountIn, _ := exactIn.MaximumAmountIn(entities.NewPercent(big.NewInt(0), big.NewInt(100)), nil)
	assert.Equal(t, amountIn, exactIn.InputAmount(), "returns exact if 0")

	// returns exact if nonzero
	amountIn, _ = exactIn.MaximumAmountIn(entities.NewPercent(big.NewInt(0), big.NewInt(100)), nil)
	assert.True(t, amountIn.EqualTo(entities.FromRawAmount(token0, big.NewInt(100)).Fraction))
	amountIn, _ = exactIn.MaximumAmountIn(entities.NewPercent(big.NewInt(5), big.NewInt(100)), nil)
	assert.True(t, amountIn.EqualTo(entities.FromRawAmount(token0, big.NewInt(100)).Fraction))
	amountIn, _ = exactIn.MaximumAmountIn(entities.NewPercent(big.NewInt(200), big.NewInt(100)), nil)
	assert.True(t, amountIn.EqualTo(entities.FromRawAmount(token0, big.NewInt(100)).Fraction))

	// tradeType = EXACT_OUTPUT
	exactOut, _ := FromRoute(r, entities.FromRawAmount(token2, big.NewInt(10000)), entities.ExactOutput)

	_, err = exactOut.MaximumAmountIn(entities.NewPercent(big.NewInt(-1), big.NewInt(10000)), nil)
	assert.ErrorIs(t, err, ErrInvalidSlippageTolerance, "throws if less than 0")

	amountIn, _ = exactOut.MaximumAmountIn(entities.NewPercent(big.NewInt(0), big.NewInt(10000)), nil)
	assert.Equal(t, amountIn, exactOut.InputAmount(), "returns exact if 0")

	// returns slippage amount if nonzero
	amountIn, _ = exactOut.MaximumAmountIn(entities.NewPercent(big.NewInt(0), big.NewInt(100)), nil)
	assert.True(t, amountIn.EqualTo(entities.FromRawAmount(token1, big.NewInt(10014)).Fraction))
	amountIn, _ = exactOut.MaximumAmountIn(entities.NewPercent(big.NewInt(5), big.NewInt(100)), nil)
	assert.True(t, amountIn.EqualTo(entities.FromRawAmount(token0, big.NewInt(10514)).Fraction))
	amountIn, _ = exactOut.MaximumAmountIn(entities.NewPercent(big.NewInt(200), big.NewInt(100)), nil)
	assert.True(t, amountIn.EqualTo(entities.FromRawAmount(token0, big.NewInt(30042)).Fraction))
}

func TestMinimumAmountOut(t *testing.T) {
	r, _ := NewRoute([]*Pool{pool_0_1, pool_1_2}, token0, token2)

	// tradeType = EXACT_INPUT
	exactIn, _ := FromRoute(r, entities.FromRawAmount(token0, big.NewInt(10000)), entities.ExactInput)

	_, err := exactIn.MinimumAmountOut(entities.NewPercent(big.NewInt(-1), big.NewInt(100)), nil)
	assert.ErrorIs(t, err, ErrInvalidSlippageTolerance, "throws if less than 0")

	amountOut, _ := exactIn.MinimumAmountOut(entities.NewPercent(big.NewInt(0), big.NewInt(10000)), nil)
	assert.Equal(t, amountOut, exactIn.OutputAmount(), "returns exact if 0")

	// returns exact if nonzero
	amountOut, _ = exactIn.MinimumAmountOut(entities.NewPercent(big.NewInt(0), big.NewInt(100)), nil)
	assert.True(t, amountOut.EqualTo(entities.FromRawAmount(token2, big.NewInt(7039)).Fraction))
	amountOut, _ = exactIn.MinimumAmountOut(entities.NewPercent(big.NewInt(5), big.NewInt(100)), nil)
	assert.True(t, amountOut.EqualTo(entities.FromRawAmount(token2, big.NewInt(6703)).Fraction))
	amountOut, _ = exactIn.MinimumAmountOut(entities.NewPercent(big.NewInt(200), big.NewInt(100)), nil)
	assert.True(t, amountOut.EqualTo(entities.FromRawAmount(token2, big.NewInt(2346)).Fraction))

	// tradeType = EXACT_OUTPUT
	exactOut, _ := FromRoute(r, entities.FromRawAmount(token2, big.NewInt(100)), entities.ExactOutput)

	_, err = exactOut.MinimumAmountOut(entities.NewPercent(big.NewInt(-1), big.NewInt(100)), nil)
	assert.ErrorIs(t, err, ErrInvalidSlippageTolerance, "throws if less than 0")

	amountOut, _ = exactOut.MinimumAmountOut(entities.NewPercent(big.NewInt(0), big.NewInt(100)), nil)
	assert.Equal(t, amountOut, exactOut.OutputAmount(), "returns exact if 0")

	// returns slippage amount if nonzero
	amountOut, _ = exactOut.MinimumAmountOut(entities.NewPercent(big.NewInt(0), big.NewInt(100)), nil)
	assert.True(t, amountOut.EqualTo(entities.FromRawAmount(token2, big.NewInt(100)).Fraction))
	amountOut, _ = exactOut.MinimumAmountOut(entities.NewPercent(big.NewInt(5), big.NewInt(100)), nil)
	assert.True(t, amountOut.EqualTo(entities.FromRawAmount(token2, big.NewInt(100)).Fraction))
	amountOut, _ = exactOut.MinimumAmountOut(entities.NewPercent(big.NewInt(200), big.NewInt(100)), nil)
	assert.True(t, amountOut.EqualTo(entities.FromRawAmount(token2, big.NewInt(100)).Fraction))
}

func TestBestTradeExactOut(t *testing.T) {
	_, err := BestTradeExactOut(nil, token0, entities.FromRawAmount(token2, big.NewInt(100)), nil, nil, nil, nil)
	assert.ErrorIs(t, err, ErrNoPools, "throws with empty pools")

	_, err = BestTradeExactOut([]*Pool{pool_0_2}, token0, entities.FromRawAmount(token2, big.NewInt(100)), &BestTradeOptions{MaxHops: 0}, nil, nil, nil)
	assert.ErrorIs(t, err, ErrInvalidMaxHops, "throws with max hops of 0")

	// provides best route
	result, err := BestTradeExactOut([]*Pool{pool_0_1, pool_0_2, pool_1_2}, token0, entities.FromRawAmount(token2, big.NewInt(10000)), nil, nil, nil, nil)
	if err != nil {
		t.Fatal(err)
	}
	assert.Equal(t, len(result), 2)
	assert.Equal(t, len(result[1].Swaps[0].Route.Pools), 1)
	assert.Equal(t, result[1].Swaps[0].Route.TokenPath, []*entities.Token{token0, token2})
	assert.True(t, result[1].InputAmount().EqualTo(entities.FromRawAmount(result[1].InputAmount().Currency, big.NewInt(12229)).Fraction))
	assert.True(t, result[1].OutputAmount().EqualTo(entities.FromRawAmount(token2, big.NewInt(10000)).Fraction))
	assert.Equal(t, len(result[0].Swaps[0].Route.Pools), 2)
	assert.Equal(t, result[0].Swaps[0].Route.TokenPath, []*entities.Token{token0, token1, token2})
	assert.True(t, result[0].InputAmount().EqualTo(entities.FromRawAmount(token0, big.NewInt(10014)).Fraction))
	assert.True(t, result[0].OutputAmount().EqualTo(entities.FromRawAmount(token2, big.NewInt(10000)).Fraction))

	// respects maxHops
	result, err = BestTradeExactOut([]*Pool{pool_0_1, pool_0_2, pool_1_2}, token0, entities.FromRawAmount(token2, big.NewInt(10)), &BestTradeOptions{MaxNumResults: 3, MaxHops: 1}, nil, nil, nil)
	if err != nil {
		t.Fatal(err)
	}
	assert.Equal(t, len(result), 1)
	assert.Equal(t, len(result[0].Swaps[0].Route.Pools), 1)
	assert.Equal(t, result[0].Swaps[0].Route.TokenPath, []*entities.Token{token0, token2})

	// // insufficient liquidity
	// result, _ = BestTradeExactOut([]*Pool{pool_0_1, pool_0_2, pool_1_2}, token0, entities.FromRawAmount(token2, big.NewInt(1200)), nil, nil, nil, nil)
	// assert.Equal(t, len(result), 0)

	// // insufficient liquidity in one pool but not the other
	// result, _ = BestTradeExactOut([]*Pool{pool_0_1, pool_0_2, pool_1_2}, token0, entities.FromRawAmount(token2, big.NewInt(1050)), nil, nil, nil, nil)
	// assert.Equal(t, len(result), 1)

	// respects n
	result, err = BestTradeExactOut([]*Pool{pool_0_1, pool_0_2, pool_1_2}, token0, entities.FromRawAmount(token2, big.NewInt(10)), &BestTradeOptions{MaxNumResults: 1, MaxHops: 3}, nil, nil, nil)
	if err != nil {
		t.Fatal(err)
	}
	assert.Equal(t, len(result), 1)

	// no path
	result, err = BestTradeExactOut([]*Pool{pool_0_1, pool_0_3, pool_1_3}, token0, entities.FromRawAmount(token2, big.NewInt(10)), nil, nil, nil, nil)
	if err != nil {
		t.Fatal(err)
	}
	assert.Equal(t, len(result), 0)

	// works for ETHER currency input
	result, err = BestTradeExactOut([]*Pool{pool_weth_0, pool_0_1, pool_0_3, pool_1_3}, Ether, entities.FromRawAmount(token3, big.NewInt(10000)), nil, nil, nil, nil)
	if err != nil {
		t.Fatal(err)
	}
	assert.Equal(t, len(result), 2)
	assert.Equal(t, result[0].InputAmount().Currency, Ether)
	assert.Equal(t, result[0].Swaps[0].Route.TokenPath, []*entities.Token{entities.WETH9[1], token0, token3})
	assert.Equal(t, result[0].OutputAmount().Currency, token3)
	assert.Equal(t, result[1].Swaps[0].Route.TokenPath, []*entities.Token{entities.WETH9[1], token0, token1, token3})
	assert.Equal(t, result[1].InputAmount().Currency, Ether)
	assert.Equal(t, result[1].OutputAmount().Currency, token3)

	// works for ETHER currency output
	result, err = BestTradeExactOut([]*Pool{pool_weth_0, pool_0_1, pool_0_3, pool_1_3}, token3, entities.FromRawAmount(Ether, big.NewInt(100)), nil, nil, nil, nil)
	if err != nil {
		t.Fatal(err)
	}
	assert.Equal(t, len(result), 2)
	assert.Equal(t, result[0].InputAmount().Currency, token3)
	assert.Equal(t, result[0].Swaps[0].Route.TokenPath, []*entities.Token{token3, token1, token0, entities.WETH9[1]})
	assert.Equal(t, result[0].OutputAmount().Currency, Ether)
	assert.Equal(t, result[1].InputAmount().Currency, token3)
	assert.Equal(t, result[1].Swaps[0].Route.TokenPath, []*entities.Token{token3, token0, entities.WETH9[1]})
	assert.Equal(t, result[1].OutputAmount().Currency, Ether)
}
