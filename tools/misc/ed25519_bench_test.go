package misc

import (
	"github.com/iotaledger/hive.go/crypto/ed25519"
	"strconv"
	"testing"
)

func BenchmarkED25519Sign(b *testing.B) {
	_, privateKey, _ := ed25519.GenerateKey()
	for i := 0; i < b.N; i++ {
		d := []byte("DataToSign" + strconv.Itoa(i))
		_ = privateKey.Sign(d)

	}
	//
	//assert.True(t, publicKey.VerifySignature(data, sig))
}

func BenchmarkED25519SignVerify(b *testing.B) {
	publicKey, privateKey, _ := ed25519.GenerateKey()

	for i := 0; i < b.N; i++ {
		d := []byte("DataToSign" + strconv.Itoa(i))
		sig := privateKey.Sign(d)
		if !publicKey.VerifySignature(d, sig) {
			panic("very bad")
		}
	}
}
