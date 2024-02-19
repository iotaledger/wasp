package cryptolib

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestVariantKeyPairValidationNil(t *testing.T) {
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
