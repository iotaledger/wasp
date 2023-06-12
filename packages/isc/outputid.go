package isc

import (
	"github.com/iotaledger/hive.go/serializer/v2/marshalutil"
	iotago "github.com/iotaledger/iota.go/v3"
	"github.com/iotaledger/wasp/packages/util/rwutil"
)

const Million = uint64(1_000_000)

var emptyOutputID = iotago.OutputID{}

func EncodeOutputID(value iotago.OutputID) []byte {
	return value[:]
}

func OutputIDFromBytes(data []byte) (ret iotago.OutputID, err error) {
	rr := rwutil.NewBytesReader(data)
	rr.ReadN(ret[:])
	return ret, rr.Err
}

func OutputIDFromMarshalUtil(mu *marshalutil.MarshalUtil) (ret iotago.OutputID, err error) {
	rr := rwutil.NewMuReader(mu)
	rr.ReadN(ret[:])
	return ret, rr.Err
}

func OutputIDToMarshalUtil(outputID iotago.OutputID, mu *marshalutil.MarshalUtil) *marshalutil.MarshalUtil {
	return mu.WriteBytes(outputID[:])
}

func IsEmptyOutputID(outputID iotago.OutputID) bool {
	return outputID == emptyOutputID
}
