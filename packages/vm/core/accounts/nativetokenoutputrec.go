package accounts

import (
	"fmt"
	"io"
	"math/big"

	iotago "github.com/iotaledger/iota.go/v3"
	"github.com/iotaledger/wasp/packages/util/rwutil"
)

type nativeTokenOutputRec struct {
	OutputID          iotago.OutputID
	Amount            *big.Int
	StorageBaseTokens uint64 // always storage deposit
}

func nativeTokenOutputRecFromBytes(data []byte) (*nativeTokenOutputRec, error) {
	return rwutil.ReadFromBytes(data, new(nativeTokenOutputRec))
}

func mustNativeTokenOutputRecFromBytes(data []byte) *nativeTokenOutputRec {
	ret, err := nativeTokenOutputRecFromBytes(data)
	if err != nil {
		panic(err)
	}
	return ret
}

func (rec *nativeTokenOutputRec) Bytes() []byte {
	return rwutil.WriteToBytes(rec)
}

func (rec *nativeTokenOutputRec) String() string {
	return fmt.Sprintf("Native Token Account: base tokens: %d, amount: %d, outID: %s",
		rec.StorageBaseTokens, rec.Amount, rec.OutputID)
}

func (rec *nativeTokenOutputRec) Read(r io.Reader) error {
	rr := rwutil.NewReader(r)
	rr.ReadN(rec.OutputID[:])
	rec.Amount = rr.ReadUint256()
	rec.StorageBaseTokens = rr.ReadAmount64()
	return rr.Err
}

func (rec *nativeTokenOutputRec) Write(w io.Writer) error {
	ww := rwutil.NewWriter(w)
	ww.WriteN(rec.OutputID[:])
	ww.WriteUint256(rec.Amount)
	ww.WriteAmount64(rec.StorageBaseTokens)
	return ww.Err
}
