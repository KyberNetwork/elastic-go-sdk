package utils

import (
	"math/big"

	"github.com/KyberNetwork/elastic-go-sdk/v2/constants"
)

// Source: https://github.com/KyberNetwork/ks-elastic-sc-v2/blob/3ba84353cbd88f30f222bb9c673e242a2e46fd12/contracts/libraries/SwapMath.sol

var FeeUnits = new(big.Int).Exp(big.NewInt(10), big.NewInt(5), nil)
var TwoFeeUnits = new(big.Int).Mul(FeeUnits, big.NewInt(2))

// ComputeSwapStep computes the actual swap input / output amounts to be deducted or added,
// the swap fee to be collected and the resulting sqrtP
func ComputeSwapStep(
	liquidity *big.Int,
	currentSqrtP *big.Int,
	targetSqrtP *big.Int,
	feeInFeeUnits constants.FeeAmount,
	specifiedAmount *big.Int,
	isExactInput bool,
	isToken0 bool,
) (
	usedAmount *big.Int,
	returnedAmount *big.Int,
	deltaL *big.Int,
	nextSqrtP *big.Int,
	err error,
) {
	// in the event currentSqrtP == targetSqrtP because of tick movements, return
	// e.g. swapped up tick where specified price limit is on an initialised tick
	// then swapping down tick will cause next tick to be the same as the current tick
	if currentSqrtP.Cmp(targetSqrtP) == 0 {
		return currentSqrtP, constants.Zero, constants.Zero, constants.Zero, nil
	}

	nextSqrtP = big.NewInt(0)

	usedAmount = calcReachAmount(
		liquidity,
		currentSqrtP,
		targetSqrtP,
		feeInFeeUnits,
		isExactInput,
		isToken0,
	)

	if isExactInput && usedAmount.Cmp(specifiedAmount) > 0 || (!isExactInput && usedAmount.Cmp(specifiedAmount) <= 0) {
		usedAmount = specifiedAmount
	} else {
		nextSqrtP = targetSqrtP
	}

	var absDelta *big.Int
	if usedAmount.Cmp(constants.Zero) >= 0 {
		absDelta = usedAmount
	} else {
		absDelta = new(big.Int).Mul(usedAmount, constants.NegativeOne)
	}

	if nextSqrtP.Cmp(constants.Zero) == 0 {
		deltaL = estimateIncrementalLiquidity(
			absDelta,
			liquidity,
			currentSqrtP,
			feeInFeeUnits,
			isExactInput,
			isToken0,
		)

		nextSqrtP = calcFinalPrice(
			absDelta,
			liquidity,
			deltaL,
			currentSqrtP,
			isExactInput,
			isToken0,
		)
	} else {
		deltaL = calcIncrementalLiquidity(
			absDelta,
			liquidity,
			currentSqrtP,
			nextSqrtP,
			isExactInput,
			isToken0,
		)
	}

	returnedAmount = calcReturnedAmount(
		liquidity,
		currentSqrtP,
		nextSqrtP,
		deltaL,
		isExactInput,
		isToken0,
	)

	return
}

// calcReachAmount calculates the amount needed to reach targetSqrtP from currentSqrtP
// we cast currentSqrtP and targetSqrtP to uint256 as they are multiplied by TWO_FEE_UNITS or feeInFeeUnits
func calcReachAmount(
	liquidity *big.Int,
	currentSqrtP *big.Int,
	targetSqrtP *big.Int,
	feeInFeeUnits constants.FeeAmount,
	isExactInput bool,
	isToken0 bool,
) (reachAmount *big.Int) {
	var absPriceDiff *big.Int

	if currentSqrtP.Cmp(targetSqrtP) >= 0 {
		absPriceDiff = new(big.Int).Sub(currentSqrtP, targetSqrtP)
	} else {
		absPriceDiff = new(big.Int).Sub(targetSqrtP, currentSqrtP)
	}

	if isExactInput {
		// we round down so that we avoid taking giving away too much for the specified input
		// i.e. require less input qty to move ticks
		if isToken0 {
			// exactInput + swap 0 -> 1
			// numerator = 2 * liquidity * absPriceDiff
			// denominator = currentSqrtP * (2 * targetSqrtP - currentSqrtP * feeInFeeUnits / FEE_UNITS)
			// overflow should not happen because the absPriceDiff is capped to ~5%
			denominator := new(big.Int).Sub(
				new(big.Int).Mul(TwoFeeUnits, targetSqrtP),
				new(big.Int).Mul(big.NewInt(int64(feeInFeeUnits)), currentSqrtP),
			)
			numerator := MulDiv(liquidity, new(big.Int).Mul(TwoFeeUnits, absPriceDiff), denominator)

			reachAmount = MulDiv(numerator, constants.Q96, currentSqrtP)
		} else {
			// exactInput + swap 1 -> 0
			// numerator: liquidity * absPriceDiff * (TWO_FEE_UNITS * targetSqrtP - feeInFeeUnits * (targetSqrtP + currentSqrtP))
			// denominator: (TWO_FEE_UNITS * targetSqrtP - feeInFeeUnits * currentSqrtP)
			// overflow should not happen because the absPriceDiff is capped to ~5%
			denominator := new(big.Int).Sub(
				new(big.Int).Mul(TwoFeeUnits, currentSqrtP),
				new(big.Int).Mul(big.NewInt(int64(feeInFeeUnits)), targetSqrtP),
			)
			numerator := MulDiv(liquidity, new(big.Int).Mul(TwoFeeUnits, absPriceDiff), denominator)

			reachAmount = MulDiv(numerator, currentSqrtP, constants.Q96)
		}
	} else {
		// we will perform negation as the last step
		// we round down so that we require less output qty to move ticks
		if isToken0 {
			// exactOut + swap 0 -> 1
			// numerator: (liquidity)(absPriceDiff)(2 * currentSqrtP - deltaL * (currentSqrtP + targetSqrtP))
			// denominator: (currentSqrtP * targetSqrtP) * (2 * currentSqrtP - deltaL * targetSqrtP)
			// overflow should not happen because the absPriceDiff is capped to ~5%
			denominator := new(big.Int).Sub(
				new(big.Int).Mul(TwoFeeUnits, currentSqrtP),
				new(big.Int).Mul(big.NewInt(int64(feeInFeeUnits)), targetSqrtP),
			)
			numerator := new(big.Int).Sub(
				denominator, new(big.Int).Mul(big.NewInt(int64(feeInFeeUnits)), currentSqrtP),
			)
			numerator = MulDiv(new(big.Int).Lsh(liquidity, 96), numerator, denominator)

			reachAmount = new(big.Int).Div(MulDiv(numerator, absPriceDiff, currentSqrtP), targetSqrtP)
			reachAmount = new(big.Int).Mul(reachAmount, constants.NegativeOne)
		} else {
			// exactOut + swap 1 -> 0
			// numerator: liquidity * absPriceDiff * (TWO_FEE_UNITS * targetSqrtP - feeInFeeUnits * (targetSqrtP + currentSqrtP))
			// denominator: (TWO_FEE_UNITS * targetSqrtP - feeInFeeUnits * currentSqrtP)
			// overflow should not happen because the absPriceDiff is capped to ~5%
			denominator := new(big.Int).Sub(
				new(big.Int).Mul(TwoFeeUnits, targetSqrtP),
				new(big.Int).Mul(big.NewInt(int64(feeInFeeUnits)), currentSqrtP),
			)
			numerator := new(big.Int).Sub(
				denominator, new(big.Int).Mul(big.NewInt(int64(feeInFeeUnits)), targetSqrtP),
			)
			numerator = MulDiv(liquidity, numerator, denominator)

			reachAmount = MulDiv(numerator, absPriceDiff, constants.Q96)
			reachAmount = new(big.Int).Mul(reachAmount, constants.NegativeOne)
		}
	}

	return reachAmount
}

// estimateIncrementalLiquidity estimates deltaL, the swap fee to be collected based on amount specified
// for the final swap step to be performed,
// where the next (temporary) tick will not be crossed
func estimateIncrementalLiquidity(
	absDelta *big.Int,
	liquidity *big.Int,
	currentSqrtP *big.Int,
	feeInFeeUnits constants.FeeAmount,
	isExactInput bool,
	isToken0 bool,
) (deltaL *big.Int) {
	// this is when we didn't reach the target (last step before loop end), then we have to recalculate the target_X96, deltaL ...
	fee := big.NewInt(int64(feeInFeeUnits))

	if isExactInput {
		if isToken0 {
			// deltaL = feeInFeeUnits * absDelta * currentSqrtP / 2
			deltaL = MulDiv(currentSqrtP, new(big.Int).Mul(absDelta, fee), new(big.Int).Lsh(TwoFeeUnits, 96))
		} else {
			// deltaL = feeInFeeUnits * absDelta * / (currentSqrtP * 2)
			// Because nextSqrtP = (liquidity + absDelta / currentSqrtP) * currentSqrtP / (liquidity + deltaL)
			// so we round down deltaL, to round up nextSqrtP
			deltaL = MulDivRoundingDown(
				constants.Q96, new(big.Int).Mul(absDelta, fee), new(big.Int).Mul(TwoFeeUnits, currentSqrtP),
			)
		}
	} else {
		// obtain the smaller root of the quadratic equation
		// ax^2 - 2bx + c = 0 such that b > 0, and x denotes deltaL
		a := fee
		b := new(big.Int).Mul(new(big.Int).Sub(FeeUnits, fee), liquidity)
		c := new(big.Int).Mul(new(big.Int).Mul(fee, liquidity), absDelta)

		if isToken0 {
			// a = feeInFeeUnits
			// b = (FEE_UNITS - feeInFeeUnits) * liquidity - FEE_UNITS * absDelta * currentSqrtP
			// c = feeInFeeUnits * liquidity * absDelta * currentSqrtP
			b = new(big.Int).Sub(b, MulDiv(new(big.Int).Mul(FeeUnits, absDelta), currentSqrtP, constants.Q96))
			c = MulDiv(c, currentSqrtP, constants.Q96)
		} else {
			// a = feeInFeeUnits
			// b = (FEE_UNITS - feeInFeeUnits) * liquidity - FEE_UNITS * absDelta / currentSqrtP
			// c = liquidity * feeInFeeUnits * absDelta / currentSqrtP
			b = new(big.Int).Sub(b, MulDiv(new(big.Int).Mul(FeeUnits, absDelta), constants.Q96, currentSqrtP))
			c = MulDiv(c, constants.Q96, currentSqrtP)
		}

		deltaL = GetSmallerRootOfQuadEqn(a, b, c)
	}

	return deltaL
}

// calcIncrementalLiquidity calculates deltaL, the swap fee to be collected for an intermediate swap step,
// where the next (temporary) tick will be crossed
func calcIncrementalLiquidity(
	absDelta *big.Int,
	liquidity *big.Int,
	currentSqrtP *big.Int,
	nextSqrtP *big.Int,
	isExactInput bool,
	isToken0 bool,
) (deltaL *big.Int) {
	var tmp1, tmp2, tmp3 *big.Int

	// this is when we reach the target, then we have target_X96
	if isToken0 {
		// deltaL = nextSqrtP * (liquidity / currentSqrtP +/- absDelta)) - liquidity
		// needs to be minimum
		tmp1 = MulDiv(liquidity, constants.Q96, currentSqrtP)
		if isExactInput {
			tmp2 = new(big.Int).Add(tmp1, absDelta)
		} else {
			tmp2 = new(big.Int).Sub(tmp1, absDelta)
		}
		tmp3 = MulDiv(nextSqrtP, tmp2, constants.Q96)

		// in edge cases where liquidity or absDelta is small
		// liquidity might be greater than nextSqrtP * ((liquidity / currentSqrtP) +/- absDelta))
		// due to rounding
		if tmp3.Cmp(liquidity) > 0 {
			deltaL = new(big.Int).Sub(tmp3, liquidity)
		} else {
			deltaL = constants.Zero
		}
	} else {
		// deltaL = (liquidity * currentSqrtP +/- absDelta) / nextSqrtP - liquidity
		// needs to be minimum
		tmp1 = MulDiv(liquidity, currentSqrtP, constants.Q96)
		if isExactInput {
			tmp2 = new(big.Int).Add(tmp1, absDelta)
		} else {
			tmp2 = new(big.Int).Sub(tmp1, absDelta)
		}
		tmp3 = MulDiv(tmp2, constants.Q96, nextSqrtP)

		// in edge cases where liquidity or absDelta is small
		// liquidity might be greater than nextSqrtP * ((liquidity / currentSqrtP) +/- absDelta))
		// due to rounding
		if tmp3.Cmp(liquidity) > 0 {
			deltaL = new(big.Int).Sub(tmp3, liquidity)
		} else {
			deltaL = constants.Zero
		}
	}

	return deltaL
}

// calcFinalPrice calculates returned output | input tokens in exchange for specified amount
// round down when calculating returned output (isExactInput) so we avoid sending too much
// round up when calculating returned input (!isExactInput) so we get desired output amount
func calcFinalPrice(
	absDelta *big.Int,
	liquidity *big.Int,
	deltaL *big.Int,
	currentSqrtP *big.Int,
	isExactInput bool,
	isToken0 bool,
) *big.Int {
	finalPrice := constants.Zero

	if isToken0 {
		tmp := MulDiv(absDelta, currentSqrtP, constants.Q96)

		if isExactInput {
			// minimise actual output (<0, make less negative) so we avoid sending too much
			// returnedAmount = deltaL * nextSqrtP - liquidity * (currentSqrtP - nextSqrtP)
			finalPrice = MulDivRoundingUp(
				new(big.Int).Add(liquidity, deltaL), currentSqrtP, new(big.Int).Add(liquidity, tmp),
			)
		} else {
			// maximise actual input (>0) so we get desired output amount
			// returnedAmount = deltaL * nextSqrtP + liquidity * (nextSqrtP - currentSqrtP)
			finalPrice = MulDiv(
				new(big.Int).Add(liquidity, deltaL), currentSqrtP, new(big.Int).Sub(liquidity, tmp),
			)
		}
	} else {
		// returnedAmount = (liquidity + deltaL)/nextSqrtP - (liquidity)/currentSqrtP
		// if exactInput, minimise actual output (<0, make less negative) so we avoid sending too much
		// if exactOutput, maximise actual input (>0) so we get desired output amount
		tmp := MulDiv(absDelta, constants.Q96, currentSqrtP)

		if isExactInput {
			finalPrice = MulDiv(
				new(big.Int).Add(liquidity, tmp), currentSqrtP, new(big.Int).Add(liquidity, deltaL),
			)
		} else {
			finalPrice = MulDivRoundingUp(
				new(big.Int).Sub(liquidity, tmp), currentSqrtP, new(big.Int).Add(liquidity, deltaL),
			)
		}
	}

	if isExactInput && finalPrice.Cmp(constants.One) == 0 {
		finalPrice = constants.Zero
	}

	return finalPrice
}

// calcReturnedAmount calculates returned output | input tokens in exchange for specified amount
// round down when calculating returned output (isExactInput) so we avoid sending too much
// round up when calculating returned input (!isExactInput) so we get desired output amount
func calcReturnedAmount(
	liquidity *big.Int,
	currentSqrtP *big.Int,
	nextSqrtP *big.Int,
	deltaL *big.Int,
	isExactInput bool,
	isToken0 bool,
) (returnedAmount *big.Int) {
	if isToken0 {
		if isExactInput {
			// minimise actual output (<0, make less negative) so we avoid sending too much
			// returnedAmount = deltaL * nextSqrtP - liquidity * (currentSqrtP - nextSqrtP)
			returnedAmount = new(big.Int).Add(
				MulDivRoundingUp(deltaL, nextSqrtP, constants.Q96),
				new(big.Int).Mul(
					MulDiv(
						liquidity, new(big.Int).Sub(currentSqrtP, nextSqrtP), constants.Q96,
					), constants.NegativeOne,
				),
			)
		} else {
			// maximise actual input (>0) so we get desired output amount
			// returnedAmount = deltaL * nextSqrtP + liquidity * (nextSqrtP - currentSqrtP)
			returnedAmount = new(big.Int).Add(
				MulDivRoundingUp(deltaL, nextSqrtP, constants.Q96),
				MulDivRoundingUp(liquidity, new(big.Int).Sub(nextSqrtP, currentSqrtP), constants.Q96),
			)
		}
	} else {
		// returnedAmount = (liquidity + deltaL)/nextSqrtP - (liquidity)/currentSqrtP
		// if exactInput, minimise actual output (<0, make less negative) so we avoid sending too much
		// if exactOutput, maximise actual input (>0) so we get desired output amount
		returnedAmount = new(big.Int).Add(
			MulDivRoundingUp(new(big.Int).Add(liquidity, deltaL), constants.Q96, nextSqrtP),
			new(big.Int).Mul(MulDivRoundingUp(liquidity, constants.Q96, currentSqrtP), constants.NegativeOne),
		)
	}

	if isExactInput && returnedAmount.Cmp(constants.One) == 0 {
		// rounding make returnedAmount == 1
		returnedAmount = constants.Zero
	}

	return returnedAmount
}
