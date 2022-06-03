package constants

import (
	"math/big"

	"github.com/KyberNetwork/uniswap-sdk-core/entities"
	"github.com/ethereum/go-ethereum/common"
)

const PoolInitCodeHash = "0xe34f199b19b2b4f47f68442619d555527d244f78a3297ea89325f843f87b8b54"

var (
	FactoryAddress = common.HexToAddress("0xdEd9a1b7C954f0B2A431e9E0C1DaB3C24605A4e9")
	AddressZero    = common.HexToAddress("0x0000000000000000000000000000000000000000")
)

// The default factory enabled fee amounts, denominated in hundredths of bips.
type FeeAmount uint64

const (
	FeeLowest  FeeAmount = 8
	FeeLow     FeeAmount = 10
	FeeMedium  FeeAmount = 40
	FeeHigh    FeeAmount = 300
	FeeHighest FeeAmount = 1000

	FeeMax FeeAmount = 10000
)

// The default factory tick spacings by fee amount.
var TickSpacings = map[FeeAmount]int{
	FeeLowest:  1,
	FeeLow:     1,
	FeeMedium:  8,
	FeeHigh:    60,
	FeeHighest: 200,
}

var (
	NegativeOne = big.NewInt(-1)
	Zero        = big.NewInt(0)
	One         = big.NewInt(1)

	// used in liquidity amount math
	Q96  = new(big.Int).Exp(big.NewInt(2), big.NewInt(96), nil)
	Q192 = new(big.Int).Exp(Q96, big.NewInt(2), nil)

	PercentZero = entities.NewFraction(big.NewInt(0), big.NewInt(1))
)
