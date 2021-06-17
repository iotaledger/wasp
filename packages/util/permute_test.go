package util

import (
	"testing"

	"github.com/iotaledger/wasp/packages/hashing"
)

func TestPermute(t *testing.T) {
	for n := uint16(1); n < 1000; n = n + 3 {
		for k := 0; k < 10; k++ {
			seed := hashing.RandomHash(nil)
			perm := NewPermutation16(n, seed[:])
			if !ValidPermutation(perm.GetArray()) {
				t.Fatalf("invalid permutation %+v", perm)
			}
		}
	}
}
