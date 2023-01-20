package accounts

import (
	"fmt"

	"github.com/iotaledger/hive.go/core/marshalutil"
	"github.com/iotaledger/hive.go/serializer/v2"
	iotago "github.com/iotaledger/iota.go/v3"
)

type NFTOutputRec struct {
	Output      *iotago.NFTOutput
	BlockIndex  uint32
	OutputIndex uint16
}

func (r *NFTOutputRec) Bytes() []byte {
	mu := marshalutil.New()
	mu.WriteUint32(r.BlockIndex).
		WriteUint16(r.OutputIndex)
	outBytes, err := r.Output.Serialize(serializer.DeSeriModeNoValidation, nil)
	if err != nil {
		panic("error serializing NFToutput")
	}
	mu.WriteBytes(outBytes)
	return mu.Bytes()
}

func (r *NFTOutputRec) String() string {
	return fmt.Sprintf("NFT Record: base tokens: %d, ID: %s, block: %d, outIdx: %d",
		r.Output.Deposit(), r.Output.NFTID, r.BlockIndex, r.OutputIndex)
}

func NFTOutputRecFromMarshalUtil(mu *marshalutil.MarshalUtil) (*NFTOutputRec, error) {
	ret := &NFTOutputRec{}
	var err error
	if ret.BlockIndex, err = mu.ReadUint32(); err != nil {
		return nil, err
	}
	if ret.OutputIndex, err = mu.ReadUint16(); err != nil {
		return nil, err
	}
	ret.Output = &iotago.NFTOutput{}
	if _, err := ret.Output.Deserialize(mu.ReadRemainingBytes(), serializer.DeSeriModeNoValidation, nil); err != nil {
		return nil, err
	}
	return ret, nil
}

func mustNFTOutputRecFromBytes(data []byte) *NFTOutputRec {
	ret, err := NFTOutputRecFromMarshalUtil(marshalutil.New(data))
	if err != nil {
		panic(err)
	}
	return ret
}
