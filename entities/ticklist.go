package entities

import (
	"errors"
	"math/big"

	"github.com/KyberNetwork/elastic-go-sdk/v2/utils"
)

const (
	TickIndexZero = 0
)

var (
	ErrZeroTickSpacing    = errors.New("tick spacing must be greater than 0")
	ErrInvalidTickSpacing = errors.New("invalid tick spacing")
	ErrZeroNet            = errors.New("tick net delta must be zero")
	ErrSorted             = errors.New("ticks must be sorted")
	ErrEmptyTickList      = errors.New("empty tick list")
	ErrBelowSmallest      = errors.New("below smallest")
	ErrAtOrAboveLargest   = errors.New("at or above largest")
)

var (
	EmptyTick = Tick{}
)

func ValidateList(ticks []Tick, tickSpacing int) error {
	if tickSpacing <= 0 {
		return ErrZeroTickSpacing
	}

	// ensure ticks are spaced appropriately
	for _, t := range ticks {
		if t.Index%tickSpacing != 0 {
			return ErrInvalidTickSpacing
		}
	}

	// ensure tick liquidity deltas sum to 0
	sum := big.NewInt(0)
	for _, tick := range ticks {
		sum.Add(sum, tick.LiquidityNet)
	}
	if sum.Cmp(big.NewInt(0)) != 0 {
		return ErrZeroNet
	}

	if !isTicksSorted(ticks) {
		return ErrSorted
	}

	return nil
}

func IsBelowSmallest(ticks []Tick, tick int) (bool, error) {
	if len(ticks) == 0 {
		return true, ErrEmptyTickList
	}

	return tick < ticks[0].Index, nil
}

func IsAtOrAboveLargest(ticks []Tick, tick int) (bool, error) {
	if len(ticks) == 0 {
		return true, ErrEmptyTickList
	}

	return tick >= ticks[len(ticks)-1].Index, nil
}

func NextInitializedTick(ticks []Tick, tick int, lte bool) (Tick, error) {
	if lte {
		isBelowSmallest, err := IsBelowSmallest(ticks, tick)
		if err != nil {
			return EmptyTick, err
		}

		if isBelowSmallest {
			return EmptyTick, ErrBelowSmallest
		}

		isAtOrAboveLargest, err := IsAtOrAboveLargest(ticks, tick)
		if err != nil {
			return EmptyTick, err
		}

		if isAtOrAboveLargest {
			return ticks[len(ticks)-1], nil
		}

		index, err := binarySearch(ticks, tick)
		if err != nil {
			return EmptyTick, err
		}

		return ticks[index], nil
	} else {
		isAtOrAboveLargest, err := IsAtOrAboveLargest(ticks, tick)
		if err != nil {
			return EmptyTick, err
		}

		if isAtOrAboveLargest {
			return EmptyTick, ErrAtOrAboveLargest
		}

		isBelowSmallest, err := IsBelowSmallest(ticks, tick)

		if err != nil {
			return EmptyTick, err
		}

		if isBelowSmallest {
			return ticks[0], nil
		}

		index, err := binarySearch(ticks, tick)
		if err != nil {
			return EmptyTick, err
		}

		return ticks[index+1], nil
	}
}

func GetNearestCurrentTick(ticks []Tick, currentTick int) (int, error) {
	// https://github.com/KyberNetwork/ks-elastic-sc/blob/3ba84353cbd88f30f222bb9c673e242a2e46fd12/contracts/PoolStorage.sol#L114
	// NearestCurrentTick is initialized with MinTick at the beginning
	isBelowSmallest, err := IsBelowSmallest(ticks, currentTick)
	if err != nil {
		return utils.MinTick, err
	}

	if isBelowSmallest {
		return utils.MinTick, nil
	}

	tick, err := NextInitializedTick(ticks, currentTick, true)
	if err != nil {
		return TickIndexZero, err
	}

	return tick.Index, nil
}

func TransformToMap(ticks []Tick) (map[int]TickData, map[int]LinkedListData) {
	tickDataByIndex := make(map[int]TickData)
	initializedTicks := make(map[int]LinkedListData)

	// Init the initializedTicks
	initializedTicks[utils.MinTick] = LinkedListData{
		Previous: utils.MinTick,
		Next:     utils.MaxTick,
	}
	initializedTicks[utils.MaxTick] = LinkedListData{
		Previous: utils.MinTick,
		Next:     utils.MaxTick,
	}

	for i, t := range ticks {
		tickDataByIndex[t.Index] = TickData{
			LiquidityGross: t.LiquidityGross,
			LiquidityNet:   t.LiquidityNet,
		}

		if len(ticks) == 1 {
			initializedTicks[t.Index] = LinkedListData{
				Next:     utils.MaxTick,
				Previous: utils.MinTick,
			}
			initializedTicks[utils.MinTick] = LinkedListData{
				Previous: utils.MinTick,
				Next:     t.Index,
			}
			initializedTicks[utils.MaxTick] = LinkedListData{
				Previous: t.Index,
				Next:     utils.MaxTick,
			}
		} else if i == 0 {
			initializedTicks[t.Index] = LinkedListData{
				Next:     ticks[i+1].Index,
				Previous: utils.MinTick,
			}
			initializedTicks[utils.MinTick] = LinkedListData{
				Previous: utils.MinTick,
				Next:     t.Index,
			}
		} else if i == len(ticks)-1 {
			initializedTicks[t.Index] = LinkedListData{
				Next:     utils.MaxTick,
				Previous: ticks[i-1].Index,
			}
			initializedTicks[utils.MaxTick] = LinkedListData{
				Previous: t.Index,
				Next:     utils.MaxTick,
			}
		} else {
			initializedTicks[t.Index] = LinkedListData{
				Next:     ticks[i+1].Index,
				Previous: ticks[i-1].Index,
			}
		}
	}

	return tickDataByIndex, initializedTicks
}

// utils

func isTicksSorted(ticks []Tick) bool {
	for i := 0; i < len(ticks)-1; i++ {
		if ticks[i].Index > ticks[i+1].Index {
			return false
		}
	}
	return true
}

/**
 * Finds the largest tick in the list of ticks that is less than or equal to tick
 * @param ticks list of ticks
 * @param tick tick to find the largest tick that is less than or equal to tick
 * @private
 */
func binarySearch(ticks []Tick, tick int) (int, error) {
	isBelowSmallest, err := IsBelowSmallest(ticks, tick)
	if err != nil {
		return TickIndexZero, err
	}

	if isBelowSmallest {
		return TickIndexZero, ErrBelowSmallest
	}

	// binary search
	start := 0
	end := len(ticks) - 1
	for start <= end {
		mid := (start + end) / 2
		if ticks[mid].Index == tick {
			return mid, nil
		} else if ticks[mid].Index < tick {
			start = mid + 1
		} else {
			end = mid - 1
		}
	}

	// if we get here, we didn't find a tick that is less than or equal to tick
	// so we return the index of the tick that is closest to tick
	if ticks[start].Index < tick {
		return start, nil
	} else {
		return start - 1, nil
	}
}
