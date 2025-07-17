package testutil

import (
	"encoding/hex"

	"github.com/samber/lo"
)

func TestBytes(n int, seed ...int) []byte {
	if n <= 0 {
		return nil
	}
	ret := make([]byte, n)
	seedVal := lo.FirstOrEmpty(seed)
	for i := 0; i < n; i++ {
		ret[i] = byte((i + seedVal) % 256)
	}
	return ret
}

func TestHex(n int, seed ...int) string {
	b := TestBytes(n, seed...)
	return hex.EncodeToString(b)
}
