package isc

import (
	"testing"

	"github.com/stretchr/testify/require"

	iotago "github.com/iotaledger/iota.go/v3"
	"github.com/iotaledger/iota.go/v3/tpkg"
)

func TestSerialize(t *testing.T) {
	nft := NFT{
		ID:       iotago.NFTID{123},
		Issuer:   tpkg.RandEd25519Address(),
		Metadata: []byte("foobar"),
	}
	nftBytes := nft.Bytes()
	deserialized, err := NFTFromBytes(nftBytes)
	require.NoError(t, err)
	require.Equal(t, nft.ID, deserialized.ID)
	require.Equal(t, nft.Issuer, deserialized.Issuer)
	require.Equal(t, nft.Metadata, deserialized.Metadata)
}
