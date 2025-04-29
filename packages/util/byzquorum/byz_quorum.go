// Package byzquorum implements Byzantine quorum algorithms for distributed consensus.
package byzquorum

func MaxF(n int) int {
	return (n - 1) / 3
}

func MinQuorum(n int) int {
	return n - MaxF(n)
}
