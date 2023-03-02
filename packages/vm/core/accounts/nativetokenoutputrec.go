package accounts

import (
	"fmt"
	"math/big"

	"github.com/iotaledger/hive.go/serializer/v2/marshalutil"
	"github.com/iotaledger/wasp/packages/kv/codec"
	"github.com/iotaledger/wasp/packages/util"
)

type nativeTokenOutputRec struct {
	StorageBaseTokens uint64 // always storage deposit
	Amount            *big.Int
	BlockIndex        uint32
	OutputIndex       uint16
}

func (f *nativeTokenOutputRec) Bytes() []byte {
	mu := marshalutil.New()
	mu.WriteUint32(f.BlockIndex).
		WriteUint16(f.OutputIndex).
		WriteUint64(f.StorageBaseTokens)
	util.WriteBytes8ToMarshalUtil(codec.EncodeBigIntAbs(f.Amount), mu)
	return mu.Bytes()
}

func (f *nativeTokenOutputRec) String() string {
	return fmt.Sprintf("Native Token Account: base tokens: %d, amount: %d, block: %d, outIdx: %d",
		f.StorageBaseTokens, f.Amount, f.BlockIndex, f.OutputIndex)
}

func nativeTokenOutputRecFromMarshalUtil(mu *marshalutil.MarshalUtil) (*nativeTokenOutputRec, error) {
	ret := &nativeTokenOutputRec{}
	var err error
	if ret.BlockIndex, err = mu.ReadUint32(); err != nil {
		return nil, err
	}
	if ret.OutputIndex, err = mu.ReadUint16(); err != nil {
		return nil, err
	}
	if ret.StorageBaseTokens, err = mu.ReadUint64(); err != nil {
		return nil, err
	}
	bigIntBin, err := util.ReadBytes8FromMarshalUtil(mu)
	if err != nil {
		return nil, err
	}
	ret.Amount = big.NewInt(0).SetBytes(bigIntBin)
	return ret, nil
}

func mustNativeTokenOutputRecFromBytes(data []byte) *nativeTokenOutputRec {
	ret, err := nativeTokenOutputRecFromMarshalUtil(marshalutil.New(data))
	if err != nil {
		panic(err)
	}
	return ret
}
