package util

import (
	"encoding/binary"
	"math/rand"

	"github.com/iotaledger/hive.go/byteutils"
	iotago "github.com/iotaledger/iota.go/v3"
	"github.com/iotaledger/iota.go/v3/ed25519"
	"github.com/iotaledger/iota.go/v3/tpkg"
	"github.com/minio/blake2b-simd"
)

type Seed *[ed25519.SeedSize]byte

func NewSeed(optionalSeedBytes ...[]byte) Seed {
	seedBytes := [ed25519.SeedSize]byte{}

	if len(optionalSeedBytes) >= 1 {
		if len(optionalSeedBytes[0]) < ed25519.SeedSize {
			panic("seed is not long enough")
		}
		copy(seedBytes[:], optionalSeedBytes[0])
		return &seedBytes
	}

	_, err := rand.Read(seedBytes[:])
	if err != nil {
		panic(err)
	}
	return &seedBytes
}

// SubSeed generates the n'th sub seed of this Seed which is then used to generate the KeyPair.
func SubSeed(seed *[ed25519.SeedSize]byte, n uint64) []byte {
	subSeed := make([]byte, ed25519.SeedSize)

	indexBytes := make([]byte, 8)
	binary.LittleEndian.PutUint64(indexBytes, n)
	hashOfIndexBytes := blake2b.Sum256(indexBytes)

	byteutils.XORBytes(subSeed, seed[:], hashOfIndexBytes[:])

	return subSeed
}

func NewPrivateKey() ed25519.PrivateKey {
	seed := tpkg.RandEd25519Seed()
	key := ed25519.NewKeyFromSeed(seed[:])
	return key
}

func AddreessFromKey(key ed25519.PrivateKey) iotago.Address {
	addr := iotago.Ed25519AddressFromPubKey(key.Public().(ed25519.PublicKey))
	return &addr
}
