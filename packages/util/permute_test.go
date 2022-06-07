package util

import (
	"testing"

	"github.com/iotaledger/wasp/packages/hashing"
	"github.com/stretchr/testify/require"
)

func TestValidPermutation(t *testing.T) {
	require.True(t, ValidPermutation([]uint16{0}))
	require.True(t, ValidPermutation([]uint16{0, 1, 2, 3, 4}))
	require.True(t, ValidPermutation([]uint16{0, 3, 2, 1, 4}))
	require.True(t, ValidPermutation([]uint16{0, 4, 2, 1, 3}))
	require.True(t, ValidPermutation([]uint16{2, 4, 1, 0, 3}))

	require.False(t, ValidPermutation([]uint16{1, 2, 3, 4, 5})) // Permutation indexes should start at 0
	require.False(t, ValidPermutation([]uint16{0, 2, 2}))
	require.False(t, ValidPermutation([]uint16{1, 2, 1}))
}

func TestPermute(t *testing.T) {
	for n := uint16(1); n < 1000; n += 3 {
		for k := 0; k < 10; k++ {
			seed := hashing.RandomHash(nil)
			perm := NewPermutation16(n, seed[:])
			require.Truef(t, ValidPermutation(perm.GetArray()), "invalid permutation %+v", perm)
		}
	}
}

func TestNext(t *testing.T) {
	for n := uint16(1); n < 100; n += 3 {
		for k := 0; k < 10; k++ {
			seed := hashing.RandomHash(nil)
			perm := NewPermutation16(n, seed[:])
			arr := make([]uint16, n*5)
			for i := range arr {
				arr[i] = perm.Next()
			}
			for i := uint16(0); i < n; i++ {
				require.Equal(t, arr[i], arr[i+n])
				require.Equal(t, arr[i], arr[i+2*n])
				require.Equal(t, arr[i], arr[i+3*n])
				require.Equal(t, arr[i], arr[i+4*n])
			}
			permArr := arr[0:n]
			require.Truef(t, ValidPermutation(permArr), "invalid permutation %+v", permArr)
			// No need to check other cases (arr[n:n*2], arr[n*2, n*3], etc):
			// If arr[0:n] is a valid iteration and for every i in [0;n) arr[i] == arr[i+k*n], then arr[n*k:n*(k+1)] is also a valid iteration
		}
	}
}

func TestNextNoCycles(t *testing.T) {
	// There are 10! = 3628800 permutations of 10 length array.
	// There is p = 1 - 3628800!/(3628795!*3628800^5) ~ 0.00028% probability of obtaining at least two equal permutations while picking 5 random ones.
	// There is q = 1 - (1-p)^10 ~ 0.0028% probability of obtaining at least two equal permutations while picking 5 random ones if you have 10 tries to do that.
	// q is a probability of test returning false positive on initial permutation length of 10. ~ 1 false positive in 36k runs.
	// For 9 q ~ 0.028% ~ 1 false positive in 3.6k runs.
	// For 8 q ~ 0.25% ~ 1 false positive in 400 runs.
	// For 7 q ~ 2.0% ~ 1 false positive in 50 runs.
	// For 6 q ~ 13.0% ~ 1 false positive in 8 runs.
	// For 5 q ~ 57.0% ~  false positive every second run.
	// That is why `n` starts at 10.
	// NOTE: the exact values might be a little bit of due to accuracy of float arythmetics in LibreOffice Calc. They are provided just to get an impression.
	for n := uint16(10); n < 100; n += 3 {
		for k := 0; k < 10; k++ {
			seed := hashing.RandomHash(nil)
			perm := NewPermutation16(n, seed[:])
			arr := make([]uint16, n*5)
			for i := range arr {
				arr[i] = perm.NextNoCycles()
			}
			permArrs := [][]uint16{arr[0:n], arr[n : 2*n], arr[2*n : 3*n], arr[3*n : 4*n], arr[4*n : 5*n]}
			for i := 0; i < len(permArrs); i++ {
				require.Truef(t, ValidPermutation(permArrs[i]), "invalid permutation %v: %+v", i, permArrs[i])
				for j := 0; j < i; j++ {
					require.Truef(t, arePermutationsDifferent(t, permArrs[i], permArrs[j]), "permutations %v and %v match: %+v", j, i, permArrs[i])
				}
			}
		}
	}
}

func arePermutationsDifferent(t *testing.T, perm1, perm2 []uint16) bool {
	if len(perm1) != len(perm2) {
		t.Fatalf("different length permutations: %v != %v", len(perm1), len(perm2))
	}
	for i := range perm1 {
		if perm1[i] != perm2[i] {
			return true
		}
	}
	return false
}
