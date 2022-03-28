package misc

import (
	"strconv"
	"testing"

	"github.com/iotaledger/wasp/packages/cryptolib"
)

func BenchmarkED25519Sign(b *testing.B) {
	keyPair := cryptolib.NewKeyPair()
	for i := 0; i < b.N; i++ {
		d := []byte("DataToSign" + strconv.Itoa(i))
		_ = keyPair.GetPrivateKey().Sign(d)
	}
	//
	//assert.True(t, publicKey.VerifySignature(data, sig))
}

func BenchmarkED25519SignVerify(b *testing.B) {
	keyPair := cryptolib.NewKeyPair()

	for i := 0; i < b.N; i++ {
		d := []byte("DataToSign" + strconv.Itoa(i))
		sig := keyPair.GetPrivateKey().Sign(d)
		if !keyPair.Verify(d, sig) {
			panic("very bad")
		}
	}
}
