package isc

import (
	bcs "github.com/iotaledger/bcs-go"
	"github.com/iotaledger/wasp/clients/iota-go/iotago"
	"github.com/iotaledger/wasp/packages/cryptolib"
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
