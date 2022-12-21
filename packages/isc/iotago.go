package isc

import (
	"fmt"

	"github.com/iotaledger/hive.go/core/marshalutil"
	iotago "github.com/iotaledger/iota.go/v3"
)

const Million = uint64(1_000_000)

var emptyOutputID = iotago.OutputID{}
var emptyNativeTokenID = iotago.NativeTokenID{}

func DecodeOutputID(b []byte, def ...iotago.OutputID) (iotago.OutputID, error) {
	if len(b) != iotago.OutputIDLength {
		if len(def) == 0 {
			return iotago.OutputID{}, fmt.Errorf("expected OutputID size %d, got %d bytes", iotago.OutputIDLength, len(b))
		}
		return def[0], nil
	}
	var ret iotago.OutputID
	copy(ret[:], b)
	return ret, nil
}

func EncodeOutputID(value iotago.OutputID) []byte {
	return value[:]
}

func OutputIDFromMarshalUtil(mu *marshalutil.MarshalUtil) (iotago.OutputID, error) {
	data, err := mu.ReadBytes(iotago.OutputIDLength)
	if err != nil {
		return iotago.OutputID{}, err
	}

	outputID, err := DecodeOutputID(data)
	if err != nil {
		return iotago.OutputID{}, err
	}

	return outputID, nil
}

func OutputIDToMarshalUtil(outputID iotago.OutputID, mu *marshalutil.MarshalUtil) *marshalutil.MarshalUtil {
	return mu.WriteBytes(EncodeOutputID(outputID))
}

func IsEmptyOutputID(outputID iotago.OutputID) bool {
	return outputID == emptyOutputID
}

func IsEmptyNativeTokenID(nativeTokenID iotago.NativeTokenID) bool {
	return nativeTokenID == emptyNativeTokenID
}
