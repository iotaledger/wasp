package util

import (
	"math/rand"
	"os"
	"time"
)

func NewPseudoRand(seed ...int64) *rand.Rand {
	if len(seed) > 0 {
		return rand.New(rand.NewSource(seed[0]))
	}

	return rand.New(rand.NewSource(time.Now().UnixNano() + int64(os.Getpid())))
}
