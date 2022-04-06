package utils

import (
	"math/big"

	"github.com/KyberNetwork/promm-sdk-go/constants"
)

var Bps = new(big.Int).Exp(big.NewInt(10), big.NewInt(6), nil)
var TwoBps = new(big.Int).Mul(Bps, big.NewInt(2))

func ComputeSwapStep(sqrtRatioCurrentX96, sqrtRatioTargetX96, liquidity, amountRemaining *big.Int, feePips constants.FeeAmount, exactIn, zeroForOne bool) (sqrtRatioNextX96, amountIn, amountOut, deltaL *big.Int, err error) {
	if sqrtRatioCurrentX96.Cmp(sqrtRatioTargetX96) == 0 {
		return sqrtRatioCurrentX96, constants.Zero, constants.Zero, constants.Zero, nil
	}

	usedAmount := calcReachAmount(sqrtRatioCurrentX96, sqrtRatioTargetX96, liquidity, feePips, exactIn, zeroForOne)

	if exactIn && usedAmount.Cmp(amountRemaining) >= 0 || (!exactIn && usedAmount.Cmp(amountRemaining) <= 0) {
		usedAmount = amountRemaining
	} else {
		sqrtRatioNextX96 = sqrtRatioTargetX96
	}

	amountIn = usedAmount

	var absUsedAmount *big.Int

	if usedAmount.Cmp(constants.Zero) >= 0 {
		absUsedAmount = usedAmount
	} else {
		absUsedAmount = new(big.Int).Mul(usedAmount, constants.NegativeOne)
	}

	if sqrtRatioNextX96.Cmp(constants.Zero) == 0 {
		deltaL = estimateIncrementalLiquidity(absUsedAmount, liquidity, sqrtRatioCurrentX96, feePips, exactIn, zeroForOne)

		sqrtRatioNextX96 = calcFinalPrice(absUsedAmount, liquidity, deltaL, sqrtRatioCurrentX96, exactIn, zeroForOne)
	} else {
		deltaL = calcIncrementalLiquidity(sqrtRatioCurrentX96, sqrtRatioTargetX96, liquidity, absUsedAmount, exactIn, zeroForOne)
	}

	return
}

func calcReachAmount(sqrtRatioCurrentX96, sqrtRatioTargetX96, liquidity *big.Int, feePips constants.FeeAmount, exactIn, zeroForOne bool) (reachAmount *big.Int) {
	var absPriceDiff *big.Int

	if sqrtRatioCurrentX96.Cmp(sqrtRatioTargetX96) >= 0 {
		absPriceDiff = new(big.Int).Sub(sqrtRatioCurrentX96, sqrtRatioTargetX96)
	} else {
		absPriceDiff = new(big.Int).Sub(sqrtRatioTargetX96, sqrtRatioCurrentX96)
	}

	if exactIn {
		if zeroForOne {
			//exactInput + swap 0 -> 1
			denominator := new(big.Int).Sub(new(big.Int).Mul(TwoBps, sqrtRatioTargetX96), new(big.Int).Mul(big.NewInt(int64(feePips)), sqrtRatioCurrentX96))
			numerator := MulDiv(liquidity, new(big.Int).Mul(TwoBps, absPriceDiff), denominator)

			reachAmount = MulDiv(numerator, constants.Q96, sqrtRatioCurrentX96)
		} else {
			//exactInput + swap 1 -> 0
			denominator := new(big.Int).Sub(new(big.Int).Mul(TwoBps, sqrtRatioCurrentX96), new(big.Int).Mul(big.NewInt(int64(feePips)), sqrtRatioTargetX96))
			numerator := MulDiv(liquidity, new(big.Int).Mul(TwoBps, absPriceDiff), denominator)

			reachAmount = MulDiv(numerator, sqrtRatioCurrentX96, constants.Q96)
		}
	} else {
		if zeroForOne {
			//exactOut + swap 0 -> 1
			denominator := new(big.Int).Sub(new(big.Int).Mul(TwoBps, sqrtRatioCurrentX96), new(big.Int).Mul(big.NewInt(int64(feePips)), sqrtRatioTargetX96))
			numerator := new(big.Int).Sub(denominator, new(big.Int).Mul(big.NewInt(int64(feePips)), sqrtRatioCurrentX96))
			numerator = MulDiv(new(big.Int).Lsh(liquidity, 96), numerator, denominator)

			reachAmount = new(big.Int).Div(MulDiv(numerator, absPriceDiff, sqrtRatioCurrentX96), sqrtRatioTargetX96)
			reachAmount = new(big.Int).Mul(reachAmount, constants.NegativeOne)
		} else {
			//exactOut + swap 1 -> 0
			denominator := new(big.Int).Sub(new(big.Int).Mul(TwoBps, sqrtRatioTargetX96), new(big.Int).Mul(big.NewInt(int64(feePips)), sqrtRatioCurrentX96))
			numerator := new(big.Int).Sub(denominator, new(big.Int).Mul(big.NewInt(int64(feePips)), sqrtRatioTargetX96))
			numerator = MulDiv(liquidity, numerator, denominator)

			reachAmount = MulDiv(numerator, absPriceDiff, constants.Q96)
			reachAmount = new(big.Int).Mul(reachAmount, constants.NegativeOne)
		}
	}

	return reachAmount
}

func calcReturnedAmount(sqrtRatioCurrentX96, sqrtRatioTargetX96, liquidity, deltaL *big.Int, exactIn, zeroForOne bool) (returnedAmount *big.Int) {
	if zeroForOne {
		if exactIn {
			returnedAmount = new(big.Int).Add(
				MulDivRoundingUp(deltaL, sqrtRatioTargetX96, constants.Q96),
				new(big.Int).Mul(MulDiv(liquidity, new(big.Int).Sub(sqrtRatioCurrentX96, sqrtRatioTargetX96), constants.Q96), constants.NegativeOne),
			)
		} else {
			returnedAmount = new(big.Int).Add(
				MulDivRoundingUp(deltaL, sqrtRatioTargetX96, constants.Q96),
				MulDivRoundingUp(liquidity, new(big.Int).Sub(sqrtRatioTargetX96, sqrtRatioCurrentX96), constants.Q96),
			)
		}
	} else {
		returnedAmount = new(big.Int).Add(
			MulDivRoundingUp(new(big.Int).Add(liquidity, deltaL), constants.Q96, sqrtRatioTargetX96),
			new(big.Int).Mul(MulDivRoundingUp(liquidity, constants.Q96, sqrtRatioCurrentX96), constants.NegativeOne),
		)
	}

	if exactIn && returnedAmount.Cmp(constants.One) == 0 {
		returnedAmount = constants.Zero
	}

	return returnedAmount
}

func calcIncrementalLiquidity(sqrtRatioCurrentX96, sqrtRatioTargetX96, liquidity, absAmount *big.Int, exactIn, zeroForOne bool) (deltaL *big.Int) {
	var tmp1, tmp2, tmp3 *big.Int

	// this is when we reach the target, then we have target_X96
	if zeroForOne {
		tmp1 = MulDiv(liquidity, constants.Q96, sqrtRatioCurrentX96)
		if exactIn {
			tmp2 = new(big.Int).Add(tmp1, absAmount)
		} else {
			tmp2 = new(big.Int).Sub(tmp1, absAmount)
		}
		tmp3 = MulDiv(sqrtRatioTargetX96, tmp2, constants.Q96)

		if tmp3.Cmp(liquidity) > 0 {
			deltaL = new(big.Int).Sub(tmp3, liquidity)
		} else {
			deltaL = constants.Zero
		}
	} else {
		tmp1 = MulDiv(liquidity, sqrtRatioCurrentX96, constants.Q96)
		if exactIn {
			tmp2 = new(big.Int).Add(tmp1, absAmount)
		} else {
			tmp2 = new(big.Int).Sub(tmp1, absAmount)
		}
		tmp3 = MulDiv(tmp2, constants.Q96, sqrtRatioTargetX96)

		if tmp3.Cmp(liquidity) > 0 {
			deltaL = new(big.Int).Sub(tmp3, liquidity)
		} else {
			deltaL = constants.Zero
		}
	}

	return deltaL
}

func estimateIncrementalLiquidity(absAmount, liquidity, sqrtRatioCurrentX96 *big.Int, feePips constants.FeeAmount, exactIn, zeroForOne bool) (deltaL *big.Int) {
	// this is when we didn't reach the target (last step before loop end), then we have to recalculate the target_X96, deltaL ...
	fee := big.NewInt(int64(feePips))

	if exactIn {
		if zeroForOne {
			deltaL = MulDiv(sqrtRatioCurrentX96, new(big.Int).Mul(absAmount, fee), new(big.Int).Lsh(TwoBps, 96))
		} else {
			// deltaL = feeInBps * absDelta * / (currentSqrtP * 2)
			// Because nextSqrtP = (liquidity + absDelta / currentSqrtP) * currentSqrtP / (liquidity + deltaL)
			// so we round up deltaL, to round down nextSqrtP
			deltaL = MulDivRoundingUp(constants.Q96, new(big.Int).Mul(absAmount, fee), new(big.Int).Mul(TwoBps, sqrtRatioCurrentX96))
		}
	} else {
		// obtain the smaller root of the quadratic equation
		// ax^2 - 2bx + c = 0 such that b > 0, and x denotes deltaL
		a := fee
		b := new(big.Int).Sub(Bps, fee)
		c := new(big.Int).Mul(new(big.Int).Mul(fee, liquidity), absAmount)

		if zeroForOne {
			b = new(big.Int).Sub(b, MulDiv(new(big.Int).Mul(Bps, absAmount), sqrtRatioCurrentX96, constants.Q96))
			c = MulDiv(c, sqrtRatioCurrentX96, constants.Q96)
		} else {
			b = new(big.Int).Sub(b, MulDiv(new(big.Int).Mul(Bps, absAmount), constants.Q96, sqrtRatioCurrentX96))
			c = MulDiv(c, constants.Q96, sqrtRatioCurrentX96)
		}

		deltaL = GetSmallerRootOfQuadEqn(a, b, c)
	}

	return deltaL
}

func calcFinalPrice(absAmount, liquidity, deltaL, sqrtRatioCurrentX96 *big.Int, exactIn, zeroForOne bool) *big.Int {
	if zeroForOne {
		tmp := MulDiv(absAmount, sqrtRatioCurrentX96, constants.Q96)

		if exactIn {
			return MulDivRoundingUp(new(big.Int).Add(liquidity, deltaL), sqrtRatioCurrentX96, new(big.Int).Add(liquidity, tmp))
		} else {
			return MulDiv(new(big.Int).Add(liquidity, deltaL), sqrtRatioCurrentX96, new(big.Int).Sub(liquidity, tmp))
		}
	} else {
		tmp := MulDiv(absAmount, constants.Q96, sqrtRatioCurrentX96)

		if exactIn {
			return MulDiv(new(big.Int).Add(liquidity, tmp), sqrtRatioCurrentX96, new(big.Int).Add(liquidity, deltaL))
		} else {
			return MulDivRoundingUp(new(big.Int).Sub(liquidity, tmp), sqrtRatioCurrentX96, new(big.Int).Add(liquidity, deltaL))
		}
	}
}
