package cryptolib

import (
	"crypto/ed25519"
	"encoding/binary"
	"fmt"
	"github.com/iotaledger/iota.go/v3/tpkg"
	"github.com/iotaledger/wasp/packages/cryptolib/byteutils"
	"golang.org/x/crypto/blake2b"
)

const (
	SeedSize      = ed25519.SeedSize
	SignatureSize = ed25519.SignatureSize
)

type Seed [SeedSize]byte

func (seed *Seed) SubSeed(n uint64) Seed {
	subSeed := make([]byte, SeedSize)

	indexBytes := make([]byte, 8)
	binary.LittleEndian.PutUint64(indexBytes, n)
	hashOfIndexBytes := blake2b.Sum256(indexBytes)

	byteutils.XORBytes(subSeed, seed[:], hashOfIndexBytes[:])

	return SeedFromByteArray(subSeed)
}

func SeedFromByteArray(seedData []byte) Seed {
	var seed Seed

	copy(seed[:], seedData)

	return seed
}

func NewSeed() Seed {
	return tpkg.RandEd25519Seed()
}

func SignatureFromBytes(bytes []byte) (result [ed25519.SignatureSize]byte, consumedBytes int, err error) {
	if len(bytes) < SignatureSize {
		err = fmt.Errorf("bytes too short")
		return
	}

	copy(result[:SignatureSize], bytes)
	consumedBytes = SignatureSize

	return
}
