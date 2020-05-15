package util

import (
	"github.com/iotaledger/wasp/packages/hashing"
	"testing"
)

func TestPermute(t *testing.T) {

	for n := uint16(1); n < 1000; n = n + 3 {
		for k := 0; k < 10; k++ {
			perm := GetPermutation(n, hashing.RandomHash(nil).Bytes())
			if !ValidPermutation(perm) {
				t.Fatalf("invalid permutation %+v", perm)
			}
		}
	}
}
