package util

import (
	"math"

	iotago "github.com/iotaledger/iota.go/v3"
	"github.com/iotaledger/wasp/packages/util/rwutil"
)

func OutputFromReader(rr *rwutil.Reader) (output iotago.Output) {
	size := rr.ReadSize32()
	if size == 0 {
		return nil
	}
	kind := rr.ReadKind()
	if rr.Err == nil {
		output, rr.Err = iotago.OutputSelector(uint32(kind))
	}
	rr.PushBack().WriteSize32(size).WriteKind(kind)
	rr.ReadSerialized(output, math.MaxInt32)
	return output
}

func OutputToWriter(ww *rwutil.Writer, output iotago.Output) {
	if output == nil {
		ww.WriteSize32(0)
		return
	}
	ww.WriteSerialized(output, math.MaxInt32)
}

func OutputFromBytes(data []byte) (iotago.Output, error) {
	rr := rwutil.NewBytesReader(data)
	return OutputFromReader(rr), rr.Err
}

func OuutputToBytes(output iotago.Output) []byte {
	ww := rwutil.NewBytesWriter()
	OutputToWriter(ww, output)
	return ww.Bytes()
}
