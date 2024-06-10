package isc

import (
	"testing"

	iotago "github.com/iotaledger/iota.go/v3"
	"github.com/iotaledger/wasp/packages/cryptolib"
	"github.com/iotaledger/wasp/packages/util/rwutil"
)

func TestNFTSerialization(t *testing.T) {
	nft := &NFT{
		ID:       iotago.NFTID{123},
		Issuer:   cryptolib.NewRandomAddress(),
		Metadata: []byte("foobar"),
	}
	rwutil.ReadWriteTest(t, nft, new(NFT))
	rwutil.BytesTest(t, nft, NFTFromBytes)
}
