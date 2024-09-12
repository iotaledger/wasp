package cryptolib_test

import (
	"crypto/rand"
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/iotaledger/wasp/packages/cryptolib"
	"github.com/iotaledger/wasp/packages/util/bcs"
)

func TestPrivateKeySerialization(t *testing.T) {
	seedBytes := make([]byte, cryptolib.SeedSize)
	rand.Read(seedBytes)
	pivkey1 := cryptolib.PrivateKeyFromSeed((cryptolib.SeedFromBytes(seedBytes)))
	pivkey2, err := cryptolib.PrivateKeyFromBytes(pivkey1.AsBytes())
	require.NoError(t, err)
	require.Equal(t, pivkey1, pivkey2)

	bcs.TestCodec(t, pivkey1)
}
