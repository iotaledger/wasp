package byz_quorum

func MaxF(n int) int {
	return (n - 1) / 3
}

func MinQuorum(n int) int {
	return n - MaxF(n)
}
