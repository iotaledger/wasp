package isc

import (
	"fmt"

	iotago "github.com/iotaledger/iota.go/v3"
	"github.com/iotaledger/wasp/packages/util/rwutil"
)

var emptyNativeTokenID = iotago.NativeTokenID{}

func NativeTokenIDFromBytes(data []byte) (ret iotago.NativeTokenID, err error) {
	rr := rwutil.NewBytesReader(data)
	rr.ReadN(ret[:])
	return ret, rr.Err
}

func MustNativeTokenIDFromBytes(data []byte) iotago.NativeTokenID {
	ret, err := NativeTokenIDFromBytes(data)
	if err != nil {
		panic(fmt.Errorf("MustNativeTokenIDFromBytes: %w", err))
	}
	return ret
}

func IsEmptyNativeTokenID(nativeTokenID iotago.NativeTokenID) bool {
	return nativeTokenID == emptyNativeTokenID
}
