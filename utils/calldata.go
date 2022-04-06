package utils

import (
	"math/big"
)

type MethodParameters struct {
	Calldata []byte   // The hex encoded calldata to perform the given operation
	Value    *big.Int // The amount of ether (wei) to send in hex
}

/**
 * Converts a big int to a hex string
 * @param bigintIsh
 * @returns The hex encoded calldata
 */
func ToHex(i *big.Int) string {
	hex := i.String()
	if len(hex)%2 != 0 {
		hex = "0" + hex
	}
	return "0x" + hex
}
