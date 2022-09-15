package misc

import (
	"strconv"
	"testing"

	"github.com/iotaledger/hive.go/core/crypto/bls"
)

var dataToSign = "Hello BLS Benchmark!"

func BenchmarkSign(b *testing.B) {
	privateKey := bls.PrivateKeyFromRandomness()

	for i := 0; i < b.N; i++ {
		d := []byte(dataToSign + strconv.Itoa(i))
		_, _ = privateKey.Sign(d)
	}
}

func BenchmarkSignVerify(b *testing.B) {
	privateKey := bls.PrivateKeyFromRandomness()

	for i := 0; i < b.N; i++ {
		d := []byte(dataToSign + strconv.Itoa(i))
		signature, _ := privateKey.Sign(d)
		if !signature.IsValid(d) {
			panic("too bad")
		}
	}
}

const sigCount = 1000

var (
	signatures  []bls.SignatureWithPublicKey
	privateKeys []bls.PrivateKey
	data        [][]byte //nolint
)

func init() {
	signatures = make([]bls.SignatureWithPublicKey, sigCount)
	privateKeys = make([]bls.PrivateKey, sigCount)

	for i := range signatures {
		privateKeys[i] = bls.PrivateKeyFromRandomness()
		signatures[i], _ = privateKeys[i].Sign([]byte(dataToSign))
	}
}

func BenchmarkAggregate(b *testing.B) {
	b.Run("aggregate 4 signatures", func(b *testing.B) {
		for i := 0; i < b.N; i++ {
			aN, _ := bls.AggregateSignatures(signatures[:4]...)
			if !aN.IsValid([]byte(dataToSign)) {
				panic("too bad")
			}
		}
	})
	b.Run("aggregate 20 signatures", func(b *testing.B) {
		for i := 0; i < b.N; i++ {
			aN, _ := bls.AggregateSignatures(signatures[:20]...)
			if !aN.IsValid([]byte(dataToSign)) {
				panic("too bad")
			}
		}
	})
	b.Run("aggregate 100 signatures", func(b *testing.B) {
		for i := 0; i < b.N; i++ {
			aN, _ := bls.AggregateSignatures(signatures[:100]...)
			if !aN.IsValid([]byte(dataToSign)) {
				panic("too bad")
			}
		}
	})
	b.Run("aggregate 500 signatures", func(b *testing.B) {
		for i := 0; i < b.N; i++ {
			aN, _ := bls.AggregateSignatures(signatures[:500]...)
			if !aN.IsValid([]byte(dataToSign)) {
				panic("too bad")
			}
		}
	})
	b.Run("aggregate 1000 signatures", func(b *testing.B) {
		for i := 0; i < b.N; i++ {
			aN, _ := bls.AggregateSignatures(signatures...)
			if !aN.IsValid([]byte(dataToSign)) {
				panic("too bad")
			}
		}
	})
}
