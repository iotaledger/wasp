package cryptolib

import (
	"crypto/rand"
	"testing"

	"github.com/stretchr/testify/require"
)

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
