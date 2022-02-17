package testkey

import (
	iotago "github.com/iotaledger/iota.go/v3"
	"github.com/iotaledger/wasp/packages/cryptolib"
)

func GenKeyAddr(seedOpt ...*cryptolib.Seed) (*cryptolib.KeyPair, iotago.Address) {
	var keyPair *cryptolib.KeyPair
	if len(seedOpt) > 0 {
		keyPair = cryptolib.NewKeyPairFromSeed(*seedOpt[0])
	} else {
		keyPair = cryptolib.NewKeyPair()
	}
	addr := keyPair.GetPublicKey().AsEd25519Address()
	return keyPair, addr
}
