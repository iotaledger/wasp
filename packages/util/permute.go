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
