package util

import (
	crypto_rand "crypto/rand"
	"math/rand"
)

// Permutation16 deterministic permutation of integers from 0 to size-1
type Permutation16 struct {
	size        uint16
	permutation []uint16
	curSeqIndex uint16
	random      *rand.Rand
}

// Seed should be provided in tests only to obtain the predicted test results.
// In production use the seed is generated using cryptographically secure random
// number generator.
// This function allways returnes permutation; error should be considered as
// warning that permutation was seeded incorrectly.
func NewPermutation16(size uint16, seedOptional ...int64) (*Permutation16, error) {
	var seed int64
	if len(seedOptional) == 0 {
		seedArray := make([]byte, 8)
		_, err := crypto_rand.Read(seedArray)
		if err != nil {
			result, _ := NewPermutation16(size, int64(0))
			return result, err
		}
		seed = 0
		for i:=0; i<8; i++ {
			seed = seed<<8 | int64(seedArray[i])
		}
	} else {
		seed = seedOptional[0]
	}
	ret := &Permutation16{
		size:   size,
		permutation: make([]uint16, size),
		random: rand.New(rand.NewSource(seed)),
	}
	for i := range ret.permutation {
		ret.permutation[i]= uint16(i)
	}
	return ret.Shuffle(), nil
}

func (perm *Permutation16) Shuffle() *Permutation16 {
	rand.Shuffle(len(perm.permutation), func(i, j int){
		perm.permutation[i], perm.permutation[j] = perm.permutation[j],  perm.permutation[i]
	})
	perm.curSeqIndex = 0
	return perm
}

func (perm *Permutation16) Current() uint16 {
	return perm.permutation[perm.curSeqIndex]
}

func (perm *Permutation16) Next() uint16 {
	ret := perm.permutation[perm.curSeqIndex]
	perm.curSeqIndex = (perm.curSeqIndex + 1) % perm.size
	return ret
}

// If the whole permutation is obtained, reshuffles it to avoid cycles
func (perm *Permutation16) NextNoCycles() uint16 {
	ret := perm.Next()
	if perm.curSeqIndex == 0 {
		perm.Shuffle()
	}
	return ret
}

func (perm *Permutation16) GetArray() []uint16 {
	ret := make([]uint16, len(perm.permutation))
	copy(ret, perm.permutation)
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

	// no need to check if all the values are different:
	// if all the numbers from 0 to n-1 exist in n length array, then every element is deffinitelly different
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

func (perm *Permutation16) ForEach(f func(i uint16) bool) {
	for _, v := range perm.permutation {
		if !f(v) {
			return
		}
	}
}
