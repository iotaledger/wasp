package testkey

import (
	iotago "github.com/iotaledger/iota.go/v3"
	"github.com/iotaledger/wasp/packages/cryptolib"
)

func GenKeyAddr(seedOpt ...*cryptolib.Seed) (*cryptolib.KeyPair, iotago.Address) {
	var keyPair cryptolib.KeyPair
	if len(seedOpt) > 0 {
		keyPair = cryptolib.NewKeyPairFromSeed(seedOpt[0])
	} else {
		keyPair = cryptolib.NewKeyPair()
	}
	addr := cryptolib.Ed25519AddressFromPubKey(keyPair.PublicKey)
	return &keyPair, addr
}
