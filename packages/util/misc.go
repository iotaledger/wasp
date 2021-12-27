package util

import "math/big"

func StringInList(s string, lst []string) bool {
	for _, l := range lst {
		if l == s {
			return true
		}
	}
	return false
}

func AllDifferentStrings(lst ...string) bool {
	for i := range lst {
		for j := range lst {
			if i >= j {
				continue
			}
			if lst[i] == lst[j] {
				return false
			}
		}
	}
	return true
}

func IsSubset(sub, super []string) bool {
	for _, s := range sub {
		if !StringInList(s, super) {
			return false
		}
	}
	return true
}

// MakeRange returns slice with a range of elements starting from to up to from-1, inclusive
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

func MinUint64(a, b uint64) uint64 {
	if a < b {
		return a
	}
	return b
}
