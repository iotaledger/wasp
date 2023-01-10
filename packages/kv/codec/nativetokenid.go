package codec

import (
	"errors"

	iotago "github.com/iotaledger/iota.go/v3"
)

func DecodeNativeTokenID(b []byte, def ...iotago.NativeTokenID) (iotago.NativeTokenID, error) {
	if len(b) != iotago.NativeTokenIDLength {
		if len(def) == 0 {
			return iotago.NativeTokenID{}, errors.New("wrong data length")
		}
		return def[0], nil
	}
	var ret iotago.NativeTokenID
	copy(ret[:], b)
	return ret, nil
}

func MustDecodeNativeTokenID(b []byte, def ...iotago.NativeTokenID) iotago.NativeTokenID {
	ret, err := DecodeNativeTokenID(b, def...)
	if err != nil {
		panic(err)
	}
	return ret
}

func EncodeNativeTokenID(value iotago.NativeTokenID) []byte {
	return value[:]
}
