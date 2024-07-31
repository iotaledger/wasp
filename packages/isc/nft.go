package isc

import (
	"io"

	"github.com/iotaledger/wasp/packages/cryptolib"
	"github.com/iotaledger/wasp/packages/util/rwutil"
	"github.com/iotaledger/wasp/sui-go/sui"
)

type NFT struct {
	ID       sui.ObjectID
	Issuer   *cryptolib.Address
	Metadata []byte
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
	nft.Issuer = cryptolib.NewEmptyAddress()
	rr.Read(nft.Issuer)
	nft.Metadata = rr.ReadBytes()
	nft.Owner = AgentIDFromReader(rr)
	return rr.Err
}

func (nft *NFT) Write(w io.Writer) error {
	ww := rwutil.NewWriter(w)
	ww.WriteN(nft.ID[:])
	ww.Write(nft.Issuer)
	ww.WriteBytes(nft.Metadata)
	AgentIDToWriter(ww, nft.Owner)
	return ww.Err
}
