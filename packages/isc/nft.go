package isc

import (
	"github.com/iotaledger/hive.go/core/marshalutil"
	"github.com/iotaledger/hive.go/serializer/v2"
	iotago "github.com/iotaledger/iota.go/v3"
)

type NFT struct {
	ID       iotago.NFTID
	Issuer   iotago.Address
	Metadata []byte  // (ImmutableMetadata)
	Owner    AgentID // can be nil
}

func (nft *NFT) Bytes(withID ...bool) []byte {
	writeID := true
	if len(withID) > 0 {
		writeID = withID[0]
	}
	mu := marshalutil.New()
	if writeID {
		mu.WriteBytes(nft.ID[:])
	}
	issuerBytes, err := nft.Issuer.Serialize(serializer.DeSeriModeNoValidation, nil)
	if err != nil {
		panic("Unexpected error serializing NFT")
	}
	mu.WriteBytes(issuerBytes)
	mu.WriteUint16(uint16(len(nft.Metadata)))
	mu.WriteBytes(nft.Metadata)
	if nft.Owner != nil {
		mu.WriteBytes(nft.Owner.Bytes())
	}
	return mu.Bytes()
}

func NFTFromMarshalUtil(mu *marshalutil.MarshalUtil, withID ...bool) (*NFT, error) {
	readID := true
	if len(withID) > 0 {
		readID = withID[0]
	}

	ret := &NFT{}
	if readID {
		idBytes, err := mu.ReadBytes(iotago.NFTIDLength)
		if err != nil {
			return nil, err
		}
		copy(ret.ID[:], idBytes)
	}

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
	if done, err2 := mu.DoneReading(); err2 != nil {
		return nil, err2
	} else if done {
		return ret, nil
	}
	ret.Owner, err = AgentIDFromMarshalUtil(mu)
	if err != nil {
		return nil, err
	}
	return ret, nil
}

func NFTFromBytes(bytes []byte, withID ...bool) (*NFT, error) {
	return NFTFromMarshalUtil(marshalutil.New(bytes), withID...)
}
