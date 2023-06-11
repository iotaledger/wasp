package isc

import (
	"github.com/iotaledger/hive.go/serializer/v2"
	"github.com/iotaledger/hive.go/serializer/v2/marshalutil"
	iotago "github.com/iotaledger/iota.go/v3"
	"github.com/iotaledger/wasp/packages/util/rwutil"
)

const AddressIsNil rwutil.Kind = 0x80

func AddressFromBytes(data []byte) (iotago.Address, error) {
	rr := rwutil.NewBytesReader(data)
	return AddressFromReader(rr), rr.Err
}

func AddressFromMarshalUtil(mu *marshalutil.MarshalUtil) (iotago.Address, error) {
	rr := rwutil.NewMuReader(mu)
	return AddressFromReader(rr), rr.Err
}

func AddressFromReader(rr *rwutil.Reader) (ret iotago.Address) {
	kind := rr.ReadKind()
	if kind == AddressIsNil {
		return nil
	}
	if rr.Err == nil {
		ret, rr.Err = iotago.AddressSelector(uint32(kind))
	}
	rr.PushBack().WriteKind(kind)
	data := make([]byte, ret.Size())
	rr.ReadN(data)
	if rr.Err != nil {
		return ret
	}
	_, rr.Err = ret.Deserialize(data, serializer.DeSeriModeNoValidation, nil)
	return ret
}

func AddressToWriter(ww *rwutil.Writer, address iotago.Address) {
	if address == nil {
		ww.WriteKind(AddressIsNil)
		return
	}
	if ww.Err == nil {
		buf, _ := address.Serialize(serializer.DeSeriModeNoValidation, nil)
		ww.WriteN(buf)
	}
}

func BytesFromAddress(address iotago.Address) []byte {
	ww := rwutil.NewBytesWriter()
	AddressToWriter(ww, address)
	return ww.Bytes()
}
