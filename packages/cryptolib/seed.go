package cryptolib

import (
	"crypto/ed25519"
	"encoding/binary"
	"fmt"

	"github.com/minio/blake2b-simd"
	"github.com/wollac/iota-crypto-demo/pkg/bip32path"
	"github.com/wollac/iota-crypto-demo/pkg/slip10"
	"github.com/wollac/iota-crypto-demo/pkg/slip10/eddsa"

	hivecrypto "github.com/iotaledger/hive.go/crypto/ed25519"
	"github.com/iotaledger/wasp/packages/cryptolib/byteutils"
	"github.com/iotaledger/wasp/packages/parameters"
)

// testnet/alphanet uses COIN_TYPE = 1
const TestnetCoinType = uint32(1)

// / IOTA coin type <https://github.com/satoshilabs/slips/blob/master/slip-0044.md>
const IotaCoinType = uint32(4218)

// / Shimmer coin type <https://github.com/satoshilabs/slips/blob/master/slip-0044.md>
const ShimmerCoinType = uint32(4219)

// SubSeed returns a Seed (ed25519 Seed) from a master seed (that has arbitrary length)
// note that the accountIndex is actually an uint31
func SubSeed(walletSeed []byte, accountIndex uint32, useLegacyDerivation ...bool) Seed {
	if len(useLegacyDerivation) > 0 && useLegacyDerivation[0] {
		seed := SeedFromBytes(walletSeed)
		return legacyDerivation(&seed, accountIndex)
	}

	coinType := TestnetCoinType // default to the testnet
	switch parameters.L1().Protocol.Bech32HRP {
	case "iota":
		coinType = IotaCoinType
	case "smr":
		coinType = ShimmerCoinType
	}

	bip32Path := fmt.Sprintf("m/44'/%d'/%d'/0'/0'", coinType, accountIndex) // this is the same as FF does it (only the account index changes, the ADDRESS_INDEX stays 0)
	path, err := bip32path.ParsePath(bip32Path)
	if err != nil {
		panic(err)
	}
	key, err := slip10.DeriveKeyFromPath(walletSeed, eddsa.Ed25519(), path)
	if err != nil {
		panic(err)
	}
	_, prvKey := key.Key.(eddsa.Seed).Ed25519Key()
	return SeedFromBytes(prvKey)
}

// ---
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

func legacyDerivation(seed *Seed, index uint32) Seed {
	subSeed := make([]byte, SeedSize)
	indexBytes := make([]byte, 8)
	binary.LittleEndian.PutUint32(indexBytes[:4], index)
	hashOfIndexBytes := blake2b.Sum256(indexBytes)
	byteutils.XORBytes(subSeed, seed[:], hashOfIndexBytes[:])
	return SeedFromBytes(subSeed)
}
