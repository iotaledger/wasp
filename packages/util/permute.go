package util

import (
	"bytes"
	"encoding/binary"
	"github.com/iotaledger/wasp/packages/hashing"
	"sort"
)

// sorting indices of operators by the request hash

type idxToPermute struct {
	idx  uint16
	hash *hashing.HashValue
}

type arrToSort []idxToPermute

func (s arrToSort) Len() int {
	return len(s)
}

func (s arrToSort) Less(i, j int) bool {
	return bytes.Compare((*s[i].hash)[:], (*s[j].hash)[:]) < 0
}

func (s arrToSort) Swap(i, j int) {
	s[i].idx, s[j].idx = s[j].idx, s[i].idx
	s[i].hash, s[j].hash = s[j].hash, s[i].hash
}

func GetPermutation(n uint16, seed []byte) []uint16 {
	if seed == nil {
		ret := make([]uint16, n)
		for i := range ret {
			ret[i] = uint16(i)
		}
		return ret
	}

	arr := make(arrToSort, n)
	var t [2]byte
	for i := range arr {
		binary.LittleEndian.PutUint16(t[:], uint16(i))
		arr[i] = idxToPermute{
			idx:  uint16(i),
			hash: hashing.HashData(seed, t[:]),
		}
	}
	sort.Sort(arr)
	ret := make([]uint16, n)
	for i := range ret {
		ret[i] = arr[i].idx
	}
	return ret
}

func ValidPermutation(perm []uint16) bool {
	n := uint16(len(perm))

	// check if every value exists
	for i := uint16(0); i < n; i++ {
		if _, found := findIndexOf(i, perm); !found {
			return false
		}
	}

	// check if all value are different
	for i, v1 := range perm {
		for j, v2 := range perm {
			if i != j && v1 == v2 {
				return false
			}
		}
	}
	return true
}

func findIndexOf(val uint16, sequence []uint16) (uint16, bool) {
	for i, s := range sequence {
		if s == val {
			return uint16(i), true
		}
	}
	return 0, false
}
