package isc

import (
	"fmt"

	iotago "github.com/iotaledger/iota.go/v3"
	"github.com/iotaledger/wasp/packages/util/rwutil"
)

func NativeTokenIDFromBytes(data []byte) (ret iotago.NativeTokenID, err error) {
	rr := rwutil.NewBytesReader(data)
	rr.ReadN(ret[:])
	rr.Close()
	return ret, rr.Err
}

func MustNativeTokenIDFromBytes(data []byte) iotago.NativeTokenID {
	ret, err := NativeTokenIDFromBytes(data)
	if err != nil {
		panic(fmt.Errorf("MustNativeTokenIDFromBytes: %w", err))
	}
	return ret
}

func NativeTokenIDToBytes(tokenID iotago.NativeTokenID) []byte {
	return tokenID[:]
}
