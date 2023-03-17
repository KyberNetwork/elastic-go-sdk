package constants

import (
	"math/big"

	"github.com/daoleno/uniswap-sdk-core/entities"
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
	Fee0008 FeeAmount = 8
	Fee001  FeeAmount = 10
	Fee002  FeeAmount = 20
	Fee004  FeeAmount = 40
	Fee01   FeeAmount = 100
	Fee025  FeeAmount = 250
	Fee03   FeeAmount = 300
	Fee1    FeeAmount = 1000
	Fee2    FeeAmount = 2000
	Fee5    FeeAmount = 5000

	FeeMax FeeAmount = 100000
)

// The default factory tick spacings by fee amount.
var TickSpacings = map[FeeAmount]int{
	Fee0008: 1,
	Fee001:  1,
	Fee002:  2,
	Fee004:  8,
	Fee01:   10,
	Fee025:  25,
	Fee03:   60,
	Fee1:    200,
	Fee2:    100,
	Fee5:    100,
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
