package cryptolib_test

import (
	"testing"

	"github.com/stretchr/testify/require"

	bcs "github.com/iotaledger/bcs-go"
	"github.com/iotaledger/wasp/v2/packages/cryptolib"
)

func TestSignatureSerialization(t *testing.T) {
	s0 := cryptolib.NewEmptySignature()
	b, err := bcs.Marshal(&s0)
	require.NoError(t, err)
	s1, err := bcs.Unmarshal[*cryptolib.Signature](b)
	require.NoError(t, err)
	require.Equal(t, s0, s1)

	s2 := cryptolib.NewDummySignature(cryptolib.NewEmptyPublicKey())
	b, err = bcs.Marshal(&s2)
	require.NoError(t, err)
	s3, err := bcs.Unmarshal[*cryptolib.Signature](b)
	require.NoError(t, err)
	require.Equal(t, s2, s3)
}
