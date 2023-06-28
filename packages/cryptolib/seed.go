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

func NewSeed() (ret Seed) {
	copy(ret[:], hivecrypto.NewSeed().Bytes())
	return ret
}

func SeedFromBytes(data []byte) (ret Seed) {
	copy(ret[:], data)
	return ret
}

func (seed *Seed) SubSeed(n uint64) Seed {
	subSeed := make([]byte, SeedSize)

	indexBytes := make([]byte, 8)
	binary.LittleEndian.PutUint64(indexBytes, n)
	hashOfIndexBytes := blake2b.Sum256(indexBytes)

	byteutils.XORBytes(subSeed, seed[:], hashOfIndexBytes[:])

	return SeedFromBytes(subSeed)
}
