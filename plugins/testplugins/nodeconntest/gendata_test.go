package nodeconntest

import (
	"github.com/iotaledger/goshimmer/dapps/valuetransfers/packages/address/signaturescheme"
	"github.com/iotaledger/hive.go/crypto/ed25519"
	"testing"
)

func TestGenOwnerAddress(t *testing.T) {
	keyPair := ed25519.GenerateKeyPair()
	t.Logf("Private key = %s", keyPair.PrivateKey.String())
	t.Logf("Public key = %s", keyPair.PublicKey.String())
	sigscheme := signaturescheme.ED25519(keyPair)
	t.Logf("Address = %s", sigscheme.Address().String())
}
