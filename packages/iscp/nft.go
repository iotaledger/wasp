package iscp

import (
	"github.com/iotaledger/hive.go/marshalutil"
	"github.com/iotaledger/hive.go/serializer/v2"
	iotago "github.com/iotaledger/iota.go/v3"
)

type NFT struct {
	ID       iotago.NFTID
	Issuer   iotago.Address
	Metadata []byte // (ImmutableMetadata)
}

func (nft *NFT) Bytes() []byte {
	mu := marshalutil.New()
	mu.WriteBytes(nft.ID[:])
	issuerBytes, err := nft.Issuer.Serialize(serializer.DeSeriModeNoValidation, nil)
	if err != nil {
		panic("Unexpected error serializing NFT")
	}
	mu.WriteBytes(issuerBytes)
	mu.WriteUint16(uint16(len(nft.Metadata)))
	mu.WriteBytes(nft.Metadata)
	return mu.Bytes()
}

func NFTFromMarshalUtil(mu *marshalutil.MarshalUtil) (*NFT, error) {
	ret := &NFT{}
	idBytes, err := mu.ReadBytes(iotago.NFTIDLength)
	if err != nil {
		return nil, err
	}
	copy(ret.ID[:], idBytes)

	currentOffset := mu.ReadOffset()
	issuerAddrType, err := mu.ReadByte()
	if err != nil {
		return nil, err
	}
	ret.Issuer, err = iotago.AddressSelector(uint32(issuerAddrType)) // this is silly, it gets casted to `byte` again inside
	if err != nil {
		return nil, err
	}
	mu.ReadSeek(-1)

	bytesRead, err := ret.Issuer.Deserialize(mu.ReadRemainingBytes(), serializer.DeSeriModeNoValidation, nil)
	if err != nil {
		return nil, err
	}
	mu.ReadSeek(currentOffset + bytesRead)

	metadataLen, err := mu.ReadUint16()
	if err != nil {
		return nil, err
	}
	ret.Metadata, err = mu.ReadBytes(int(metadataLen))
	if err != nil {
		return nil, err
	}
	return ret, nil
}

func NFTFromBytes(bytes []byte) (*NFT, error) {
	return NFTFromMarshalUtil(marshalutil.New(bytes))
}
