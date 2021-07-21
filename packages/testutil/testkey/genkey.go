package testkey

import (
	"github.com/iotaledger/goshimmer/packages/ledgerstate"
	"github.com/iotaledger/hive.go/crypto/ed25519"
)

func GenKeyAddr(seedOpt ...*ed25519.Seed) (*ed25519.KeyPair, ledgerstate.Address) {
	var keyPair ed25519.KeyPair
	if len(seedOpt) > 0 {
		keyPair = *seedOpt[0].KeyPair(1)
	} else {
		keyPair = ed25519.GenerateKeyPair()
	}
	addr := ledgerstate.NewED25519Address(keyPair.PublicKey)
	return &keyPair, addr
}
