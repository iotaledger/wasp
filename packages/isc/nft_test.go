package isc

import (
	"testing"

	iotago "github.com/iotaledger/iota.go/v3"
	"github.com/iotaledger/iota.go/v3/tpkg"
	"github.com/iotaledger/wasp/packages/util/rwutil"
)

func TestNFTSerialization(t *testing.T) {
	nft := &NFT{
		ID:       iotago.NFTID{123},
		Issuer:   tpkg.RandEd25519Address(),
		Metadata: []byte("foobar"),
	}
	rwutil.ReadWriteTest(t, nft, new(NFT))
	rwutil.BytesTest(t, nft, NFTFromBytes)
}
