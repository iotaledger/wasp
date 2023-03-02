package cryptolib

import (
	"crypto/ed25519"
	"encoding/binary"

	"golang.org/x/crypto/blake2b"

	hivecrypto "github.com/iotaledger/hive.go/crypto/ed25519"
	"github.com/iotaledger/wasp/packages/cryptolib/byteutils"
)

const (
	SeedSize = ed25519.SeedSize
)

type Seed [SeedSize]byte

func NewSeed() Seed {
	newSeedBytes := hivecrypto.NewSeed().Bytes()
	var newSeed Seed
	copy(newSeed[:], newSeedBytes)
	return newSeed
}

func NewSeedFromBytes(seedData []byte) Seed {
	var seed Seed

	copy(seed[:], seedData)

	return seed
}

func (seed *Seed) SubSeed(n uint64) Seed {
	subSeed := make([]byte, SeedSize)

	indexBytes := make([]byte, 8)
	binary.LittleEndian.PutUint64(indexBytes, n)
	hashOfIndexBytes := blake2b.Sum256(indexBytes)

	byteutils.XORBytes(subSeed, seed[:], hashOfIndexBytes[:])

	return NewSeedFromBytes(subSeed)
}
