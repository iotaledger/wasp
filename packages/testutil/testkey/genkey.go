package testkey

import (
	"github.com/iotaledger/goshimmer/packages/ledgerstate"
	"github.com/iotaledger/hive.go/crypto/ed25519"
)

func GenKeyAddr() (*ed25519.KeyPair, ledgerstate.Address) {
	keyPair := ed25519.GenerateKeyPair()
	addr := ledgerstate.NewED25519Address(keyPair.PublicKey)
	return &keyPair, addr
}
