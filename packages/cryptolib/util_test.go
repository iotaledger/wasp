package cryptolib

import (
	"crypto/rand"
	"testing"

	"github.com/stretchr/testify/require"
)

// TODO: remove
/*func TestVariantKeyPairValidationNil(t *testing.T) {
	var vkp VariantKeyPair

	require.False(t, IsVariantKeyPairValid(vkp))
}

func TestVariantKeyPairValidationNilPtr(t *testing.T) {
	var kp *KeyPair
	var vkp VariantKeyPair = kp

	require.False(t, IsVariantKeyPairValid(vkp))
}

func TestVariantKeyPairValidation(t *testing.T) {
	kp := NewKeyPair()
	var vkp VariantKeyPair = kp

	require.NotNil(t, kp)
	require.True(t, IsVariantKeyPairValid(vkp))
}

func TestVariantKeyPairValidationCastPtr(t *testing.T) {
	kp := KeyPair{}
	var vkp VariantKeyPair = &kp

	require.True(t, IsVariantKeyPairValid(vkp))
}
*/

func TestEncodeDecodeHex(t *testing.T) {
	testEncodeDecodeHex(t, 0)
	testEncodeDecodeHex(t, 1)
	testEncodeDecodeHex(t, 2)
	testEncodeDecodeHex(t, 32)
	testEncodeDecodeHex(t, 64)
}

func testEncodeDecodeHex(t *testing.T, size int) {
	dataIn := make([]byte, size)
	rand.Read(dataIn)

	hex := EncodeHex(dataIn)
	dataOut, err := DecodeHex(hex)
	require.NoError(t, err)
	require.Equal(t, dataIn, dataOut)
}
