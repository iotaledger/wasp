package accounts

import (
	"fmt"
	"io"

	iotago "github.com/iotaledger/iota.go/v3"
	"github.com/iotaledger/wasp/packages/util/rwutil"
)

type NFTOutputRec struct {
	BlockIndex  uint32
	OutputIndex uint16
	Output      *iotago.NFTOutput
}

func nftOutputRecFromBytes(data []byte) (*NFTOutputRec, error) {
	return rwutil.ReadFromBytes(data, new(NFTOutputRec))
}

func mustNFTOutputRecFromBytes(data []byte) *NFTOutputRec {
	ret, err := nftOutputRecFromBytes(data)
	if err != nil {
		panic(err)
	}
	return ret
}

func (rec *NFTOutputRec) Bytes() []byte {
	return rwutil.WriteToBytes(rec)
}

func (rec *NFTOutputRec) String() string {
	return fmt.Sprintf("NFT Record: base tokens: %d, ID: %s, block: %d, outIdx: %d",
		rec.Output.Deposit(), rec.Output.NFTID, rec.BlockIndex, rec.OutputIndex)
}

func (rec *NFTOutputRec) Read(r io.Reader) error {
	rr := rwutil.NewReader(r)
	rec.BlockIndex = rr.ReadUint32()
	rec.OutputIndex = rr.ReadUint16()
	rec.Output = new(iotago.NFTOutput)
	rr.ReadSerialized(rec.Output)
	return rr.Err
}

func (rec *NFTOutputRec) Write(w io.Writer) error {
	ww := rwutil.NewWriter(w)
	ww.WriteUint32(rec.BlockIndex)
	ww.WriteUint16(rec.OutputIndex)
	ww.WriteSerialized(rec.Output)
	return ww.Err
}
