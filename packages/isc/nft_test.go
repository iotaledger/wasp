package isc

import (
	"testing"

	"github.com/iotaledger/wasp/packages/cryptolib"
	"github.com/iotaledger/wasp/packages/util/rwutil"
	"github.com/iotaledger/wasp/sui-go/sui"
)

func TestNFTSerialization(t *testing.T) {
	nft := &NFT{
		ID:       sui.ObjectID{123},
		Issuer:   cryptolib.NewRandomAddress(),
		Metadata: []byte("foobar"),
	}
	rwutil.ReadWriteTest(t, nft, new(NFT))
	rwutil.BytesTest(t, nft, NFTFromBytes)
}
