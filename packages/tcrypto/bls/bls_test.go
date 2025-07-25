package bls_test

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/iotaledger/wasp/v2/packages/tcrypto/bls"
)

var dataToSign = []byte("Hello BLS Test!")

func TestAggregateSignatures(t *testing.T) {
	signatureCount := 20

	signatures := make([]bls.SignatureWithPublicKey, signatureCount)
	privateKeys := make([]bls.PrivateKey, signatureCount)

	for i := range signatures {
		privateKeys[i] = bls.PrivateKeyFromRandomness()
		signature, err := privateKeys[i].Sign(dataToSign)
		require.NoError(t, err)
		signatures[i] = signature
	}

	// aggregate 2 signatures
	a01, err := bls.AggregateSignatures(signatures[0], signatures[1])
	require.NoError(t, err)
	assert.True(t, a01.IsValid(dataToSign))

	// aggregate N signatures
	aN, err := bls.AggregateSignatures(signatures...)
	require.NoError(t, err)
	assert.True(t, aN.IsValid(dataToSign))
}

func TestSingleSignature(t *testing.T) {
	privateKey := bls.PrivateKeyFromRandomness()

	signature, err := privateKey.Sign(dataToSign)
	require.NoError(t, err)
	assert.True(t, signature.IsValid(dataToSign))
	assert.False(t, signature.IsValid([]byte("some other data")))
}

func TestMarshalPublicKey(t *testing.T) {
	privateKey := bls.PrivateKeyFromRandomness()
	pubKey := privateKey.PublicKey()

	pubKeyBytes := pubKey.Bytes()
	pubKeyBack, err := bls.PublicKeyFromBytes(pubKeyBytes)
	require.NoError(t, err)
	require.EqualValues(t, pubKeyBytes, pubKeyBack.Bytes())
}
