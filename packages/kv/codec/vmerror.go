package codec

import (
	"fmt"

	"github.com/iotaledger/wasp/packages/isc"
)

func DecodeVMErrorCode(b []byte, def ...isc.VMErrorCode) (isc.VMErrorCode, error) {
	if b == nil {
		if len(def) == 0 {
			return isc.VMErrorCode{}, fmt.Errorf("cannot decode nil bytes")
		}
		return def[0], nil
	}
	return isc.VMErrorCodeFromBytes(b)
}

func MustDecodeVMErrorCode(b []byte, def ...isc.VMErrorCode) isc.VMErrorCode {
	code, err := DecodeVMErrorCode(b, def...)
	if err != nil {
		panic(err)
	}
	return code
}

func EncodeVMErrorCode(code isc.VMErrorCode) []byte {
	return code.Bytes()
}
