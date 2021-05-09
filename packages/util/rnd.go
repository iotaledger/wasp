package util

import "math/rand"

func SelectRandomUint16(selectFrom []uint16, seed int64) uint16 {
	rnd := rand.New(rand.NewSource(seed))
	return selectFrom[rnd.Intn(len(selectFrom))]
}
