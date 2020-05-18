package nodeconntest

import (
	"github.com/iotaledger/goshimmer/dapps/valuetransfers/packages/address/signaturescheme"
	"github.com/iotaledger/hive.go/crypto/ed25519"
	"github.com/iotaledger/wasp/packages/sctransaction"
	"testing"
)

func TestGenOwnerAddress(t *testing.T) {
	keyPair := ed25519.GenerateKeyPair()
	t.Logf("Private key = %s", keyPair.PrivateKey.String())
	t.Logf("Public key = %s", keyPair.PublicKey.String())
	sigscheme := signaturescheme.ED25519(keyPair)
	t.Logf("Address = %s", sigscheme.Address().String())

	txId := sctransaction.RandomTransactionID()
	t.Logf("random tx id = %s", txId.String())
}
