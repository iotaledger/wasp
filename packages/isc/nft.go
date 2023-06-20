package isc

import (
	"io"

	iotago "github.com/iotaledger/iota.go/v3"
	"github.com/iotaledger/wasp/packages/util/rwutil"
)

type NFT struct {
	ID       iotago.NFTID
	Issuer   iotago.Address
	Metadata []byte  // (ImmutableMetadata)
	Owner    AgentID // can be nil
}

func NFTFromBytes(data []byte) (*NFT, error) {
	return rwutil.ReadFromBytes(data, new(NFT))
}

func NFTFromReader(rr *rwutil.Reader) (ret *NFT, err error) {
	ret = new(NFT)
	rr.Read(ret)
	return ret, rr.Err
}

func (nft *NFT) Bytes() []byte {
	return rwutil.WriteToBytes(nft)
}

func (nft *NFT) Read(r io.Reader) error {
	rr := rwutil.NewReader(r)
	rr.ReadN(nft.ID[:])
	nft.Issuer = AddressFromReader(rr)
	nft.Metadata = rr.ReadBytes()
	nft.Owner = AgentIDFromReader(rr)
	return rr.Err
}

func (nft *NFT) Write(w io.Writer) error {
	ww := rwutil.NewWriter(w)
	ww.WriteN(nft.ID[:])
	AddressToWriter(ww, nft.Issuer)
	ww.WriteBytes(nft.Metadata)
	AgentIDToWriter(ww, nft.Owner)
	return ww.Err
}
