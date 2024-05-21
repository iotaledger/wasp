package testkey

import (
	"github.com/iotaledger/wasp/packages/cryptolib"
)

func GenKeyAddr(seedOpt ...*cryptolib.Seed) (*cryptolib.KeyPair, *cryptolib.Address) {
	var keyPair *cryptolib.KeyPair
	if len(seedOpt) > 0 {
		keyPair = cryptolib.KeyPairFromSeed(*seedOpt[0])
	} else {
		keyPair = cryptolib.NewKeyPair()
	}
	addr := keyPair.GetPublicKey().AsAddress()
	return keyPair, addr
}
