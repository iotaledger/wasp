// Package bigint provides utilities for working with large integers.
package bigint

import "math/big"

func Equal(a *big.Int, b *big.Int) bool {
	return a.Cmp(b) == 0
}

func Less(a *big.Int, b *big.Int) bool {
	return a.Cmp(b) < 0
}

func LessEqual(a *big.Int, b *big.Int) bool {
	return a.Cmp(b) <= 0
}

func Larger(a *big.Int, b *big.Int) bool {
	return a.Cmp(b) > 0
}

func LargerEqual(a *big.Int, b *big.Int) bool {
	return a.Cmp(b) >= 0
}

func Add(a *big.Int, b *big.Int) *big.Int {
	c := new(big.Int)
	return c.Add(a, b)
}

func Sub(a *big.Int, b *big.Int) *big.Int {
	c := new(big.Int)
	return c.Sub(a, b)
}

func Mul(a *big.Int, b *big.Int) *big.Int {
	c := new(big.Int)
	return c.Mul(a, b)
}

func Div(a *big.Int, b *big.Int) *big.Int {
	c := new(big.Int)
	return c.Div(a, b)
}

func Inc(a *big.Int) *big.Int {
	c := new(big.Int)
	return c.Add(a, big.NewInt(1))
}

func IsPositive(n *big.Int) bool {
	return n.Sign() > 0
}

func IsZero(bi *big.Int) bool {
	// see https://stackoverflow.com/questions/64257065/is-there-another-way-of-testing-if-a-big-int-is-0
	return len(bi.Bits()) == 0
}
