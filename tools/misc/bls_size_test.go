package misc

import (
	"testing"

	"github.com/iotaledger/hive.go/core/crypto/bls"
)

func TestSizes(t *testing.T) {
	privateKey := bls.PrivateKeyFromRandomness()
	signature, _ := privateKey.Sign([]byte(dataToSign))
	t.Logf("private key len: %d", len(privateKey.Bytes()))
	t.Logf("public key len: %d", len(privateKey.PublicKey().Bytes()))
	t.Logf("signature len: %d", len(signature.Signature))
	t.Logf("signature with public key len: %d", len(signature.Bytes()))
}
