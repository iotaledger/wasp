package iotago_test

import (
	"bytes"
	"testing"

	bcs "github.com/iotaledger/bcs-go"
	"github.com/iotaledger/wasp/v2/clients/iota-go/iotago"
	"github.com/iotaledger/wasp/v2/clients/iota-go/iotago/iotatest"
	"github.com/stretchr/testify/require"
)

func TestDigestBcs(t *testing.T) {
	digest1 := iotatest.RandomDigest()
	b, err := bcs.Marshal(&digest1)
	require.NoError(t, err)
	require.Len(t, b, iotago.DigestSize+1)
	require.Equal(t, b[0], uint8(32))
	digest2, err := bcs.Unmarshal[iotago.Digest](b)
	require.NoError(t, err)
	bytes.Equal(digest1[:], digest2[:])
}
