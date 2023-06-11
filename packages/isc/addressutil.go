package isc

import (
	"fmt"

	"github.com/iotaledger/hive.go/serializer/v2"
	"github.com/iotaledger/hive.go/serializer/v2/marshalutil"
	iotago "github.com/iotaledger/iota.go/v3"
	"github.com/iotaledger/wasp/packages/util/rwutil"
)

func BytesFromAddress(address iotago.Address) []byte {
	addressInBytes, err := address.Serialize(serializer.DeSeriModeNoValidation, nil)
	if err != nil {
		return nil
	}
	return addressInBytes
}

// AddressFromBytes unmarshals an Address from a sequence of bytes.
func AddressFromBytes(bytes []byte) (address iotago.Address, err error) {
	marshalUtil := marshalutil.New(bytes)
	if address, err = AddressFromMarshalUtil(marshalUtil); err != nil {
		err = fmt.Errorf("failed to parse Address from MarshalUtil: %w", err)
	}
	return
}

func AddressFromMarshalUtil(mu *marshalutil.MarshalUtil) (iotago.Address, error) {
	rr := rwutil.NewMuReader(mu)
	return AddressFromReader(rr), rr.Err
}

func AddressFromReader(rr *rwutil.Reader) (ret iotago.Address) {
	kind := rr.ReadKind()
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

func AddressToWriter(ww *rwutil.Writer, a iotago.Address) {
	if a == nil {
		panic("nil address")
	}
	if ww.Err == nil {
		buf, _ := a.Serialize(serializer.DeSeriModeNoValidation, nil)
		ww.WriteN(buf)
	}
}
