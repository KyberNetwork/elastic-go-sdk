package utils

import (
	"math/big"

	"github.com/KyberNetwork/promm-sdk-go/constants"
)

func MulDivRoundingUp(a, b, denominator *big.Int) *big.Int {
	product := new(big.Int).Mul(a, b)
	result := new(big.Int).Div(product, denominator)
	if new(big.Int).Rem(product, denominator).Cmp(big.NewInt(0)) != 0 {
		result.Add(result, constants.One)
	}
	return result
}

func MulDivRoundingDown(a, b, denominator *big.Int) *big.Int {
	product := new(big.Int).Mul(a, b)
	result := new(big.Int).Quo(product, denominator)
	return result
}

func MulDiv(a, b, denominator *big.Int) *big.Int {
	product := new(big.Int).Mul(a, b)

	return new(big.Int).Div(product, denominator)
}

func GetSmallerRootOfQuadEqn(a, b, c *big.Int) *big.Int {
	// smallerRoot = (b - sqrt(b * b - a * c)) / a;
	tmp1 := new(big.Int).Mul(b, b)
	tmp2 := new(big.Int).Mul(a, c)
	tmp3 := new(big.Int).Sqrt(new(big.Int).Sub(tmp1, tmp2))
	tmp4 := new(big.Int).Sub(b, tmp3)

	return new(big.Int).Div(tmp4, a)
}
