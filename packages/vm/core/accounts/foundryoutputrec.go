package accounts

import (
	"io"

	iotago "github.com/iotaledger/iota.go/v3"
	"github.com/iotaledger/wasp/packages/kv/codec"
	"github.com/iotaledger/wasp/packages/util/rwutil"
)

// foundryOutputRec contains information to reconstruct output
type foundryOutputRec struct {
	OutputID    iotago.OutputID
	Amount      uint64 // always storage deposit
	TokenScheme iotago.TokenScheme
	Metadata    []byte
}

func (rec *foundryOutputRec) Bytes() []byte {
	return rwutil.WriteToBytes(rec)
}

func foundryOutputRecFromBytes(data []byte) (*foundryOutputRec, error) {
	return rwutil.ReadFromBytes(data, new(foundryOutputRec))
}

func mustFoundryOutputRecFromBytes(data []byte) *foundryOutputRec {
	ret, err := foundryOutputRecFromBytes(data)
	if err != nil {
		panic(err)
	}
	return ret
}

func (rec *foundryOutputRec) Read(r io.Reader) error {
	rr := rwutil.NewReader(r)
	rr.ReadN(rec.OutputID[:])
	rec.Amount = rr.ReadUint64()
	tokenScheme := rr.ReadBytes()
	if rr.Err == nil {
		rec.TokenScheme, rr.Err = codec.DecodeTokenScheme(tokenScheme)
	}
	rec.Metadata = rr.ReadBytes()
	return rr.Err
}

func (rec *foundryOutputRec) Write(w io.Writer) error {
	ww := rwutil.NewWriter(w)
	ww.WriteN(rec.OutputID[:])
	ww.WriteUint64(rec.Amount)
	if ww.Err == nil {
		tokenScheme := codec.EncodeTokenScheme(rec.TokenScheme)
		ww.WriteBytes(tokenScheme)
	}
	ww.WriteBytes(rec.Metadata)
	return ww.Err
}
