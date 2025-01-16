package isc

import (
	"github.com/iotaledger/wasp/clients/iota-go/iotago"
	"github.com/iotaledger/wasp/packages/cryptolib"
	"github.com/iotaledger/wasp/packages/util/bcs"
)

type NFT struct {
	ID       iotago.ObjectID
	Issuer   *cryptolib.Address
	Metadata []byte
	Owner    AgentID // can be nil
}

func NFTFromBytes(data []byte) (*NFT, error) {
	return bcs.Unmarshal[*NFT](data)
}

func (nft *NFT) Bytes() []byte {
	return bcs.MustMarshal(nft)
}

// CollectionNFTObjectID returns the address of the collection NFT, if the NFT
// belongs to a collection.
func (nft *NFT) CollectionNFTObjectID() (iotago.ObjectID, bool) {
	// TODO implement me
	return iotago.ObjectID{}, false
}
