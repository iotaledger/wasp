package util

import (
	"math/big"

	"github.com/iotaledger/wasp/packages/hashing"
)

const WindowsOS = "windows"

var (
	Big0       = big.NewInt(0)
	Big1       = big.NewInt(1)
	MaxUint256 = new(big.Int).Sub(new(big.Int).Lsh(Big1, 256), Big1)
)

func ExecuteIfNotNil(function func()) {
	if function != nil {
		function()
	}
}

// MakeRange returns slice with a range of elements starting from to up to-1, inclusive
func MakeRange(from, to int) []int {
	a := make([]int, to-from)
	for i := range a {
		a[i] = from + i
	}
	return a
}

func IsZeroBigInt(bi *big.Int) bool {
	// see https://stackoverflow.com/questions/64257065/is-there-another-way-of-testing-if-a-big-int-is-0
	return len(bi.Bits()) == 0
}

func IsPositiveBigInt(n *big.Int) bool {
	return n.Sign() > 0
}

func GetHashValue(obj interface{ Bytes() []byte }) hashing.HashValue {
	return hashing.HashData(obj.Bytes())
}
