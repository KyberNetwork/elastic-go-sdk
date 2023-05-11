package utils

import (
	"math/big"
)

func ApplyLiquidityDelta(
	liquidity *big.Int,
	liquidityDelta *big.Int,
	isAddLiquidity bool,
) *big.Int {
	if isAddLiquidity {
		return new(big.Int).Add(liquidity, liquidityDelta)
	} else {
		return new(big.Int).Sub(liquidity, liquidityDelta)
	}
}
