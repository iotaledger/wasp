package isc

import (
	"testing"

	"github.com/stretchr/testify/require"

	iotago "github.com/iotaledger/iota.go/v3"
)

func TestAddNFTs(t *testing.T) {
	nftSet1 := []iotago.NFTID{
		{1},
		{2},
		{3},
	}

	nftSet2 := []iotago.NFTID{
		{3},
		{4},
		{5},
	}
	a := NewAllowance(0, nil, nftSet1)
	b := NewAllowance(0, nil, nftSet2)
	a.Add(b)
	require.Len(t, a.NFTs, 5)
	require.Contains(t, a.NFTs, iotago.NFTID{1})
	require.Contains(t, a.NFTs, iotago.NFTID{2})
	require.Contains(t, a.NFTs, iotago.NFTID{3})
	require.Contains(t, a.NFTs, iotago.NFTID{4})
	require.Contains(t, a.NFTs, iotago.NFTID{5})
}
