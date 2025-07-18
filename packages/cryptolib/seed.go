package cryptolib

import (
	"crypto/ed25519"

	"github.com/wollac/iota-crypto-demo/pkg/bip32path"
	"github.com/wollac/iota-crypto-demo/pkg/slip10"
	"github.com/wollac/iota-crypto-demo/pkg/slip10/eddsa"

	hivecrypto "github.com/iotaledger/hive.go/crypto/ed25519"

	"github.com/iotaledger/wasp/v2/clients/iota-go/iotasigner"
)

// TestnetCoinType is used by testnet/alphanet with COIN_TYPE = 1
const TestnetCoinType = iotasigner.TestnetCoinType

// IotaCoinType is the IOTA coin type <https://github.com/satoshilabs/slips/blob/master/slip-0044.md>
const IotaCoinType = iotasigner.IotaCoinType

// SubSeed returns a Seed (ed25519 Seed) from a master seed (that has arbitrary length)
// note that the accountIndex is actually an uint31
func SubSeed(walletSeed []byte, accountIndex uint32) Seed {
	bip32Path, err := iotasigner.BuildBip32Path(iotasigner.SignatureFlagEd25519, iotasigner.IotaCoinType, accountIndex)
	if err != nil {
		panic(err)
	}

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

var (
	TestSeed    = SeedFromBytes([]byte{0x12, 0x34, 0x56, 0x78, 0x90, 0xab, 0xcd, 0xef})
	TestKeyPair = KeyPairFromSeed(TestSeed)
)
