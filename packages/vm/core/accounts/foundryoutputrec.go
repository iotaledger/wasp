package accounts

import (
	"io"

	iotago "github.com/iotaledger/iota.go/v3"
	"github.com/iotaledger/wasp/packages/kv/codec"
	"github.com/iotaledger/wasp/packages/util/rwutil"
)

// foundryOutputRec contains information to reconstruct output
type foundryOutputRec struct {
	BlockIndex  uint32
	OutputIndex uint16
	Amount      uint64 // always storage deposit
	TokenScheme iotago.TokenScheme
	Metadata    []byte
}

func (rec *foundryOutputRec) Bytes() []byte {
	return rwutil.WriterToBytes(rec)
}

func mustFoundryOutputRecFromBytes(data []byte) *foundryOutputRec {
	ret, err := rwutil.ReaderFromBytes(data, new(foundryOutputRec))
	if err != nil {
		panic(err)
	}
	return ret
}

func (rec *foundryOutputRec) Read(r io.Reader) error {
	rr := rwutil.NewReader(r)
	rec.BlockIndex = rr.ReadUint32()
	rec.OutputIndex = rr.ReadUint16()
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
	ww.WriteUint32(rec.BlockIndex)
	ww.WriteUint16(rec.OutputIndex)
	ww.WriteUint64(rec.Amount)
	if ww.Err == nil {
		tokenScheme := codec.EncodeTokenScheme(rec.TokenScheme)
		ww.WriteBytes(tokenScheme)
	}
	ww.WriteBytes(rec.Metadata)
	return ww.Err
}
