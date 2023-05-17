package entities

import "math/big"

type Tick struct {
	Index          int
	LiquidityGross *big.Int
	LiquidityNet   *big.Int
}

type TickData struct {
	LiquidityGross *big.Int
	LiquidityNet   *big.Int
}

type LinkedListData struct {
	Previous int
	Next     int
}

// Provides information about ticks
type TickDataProvider interface {
	GetNearestCurrentTick(currentTick int) (int, error)

	TransformToMap() (map[int]TickData, map[int]LinkedListData)
}
