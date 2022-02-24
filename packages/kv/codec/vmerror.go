package codec

import (
	"github.com/iotaledger/wasp/packages/iscp"
	"golang.org/x/xerrors"
)

func DecodeVMErrorCode(b []byte, def ...iscp.VMErrorCode) (iscp.VMErrorCode, error) {
	if b == nil {
		if len(def) == 0 {
			return iscp.VMErrorCode{}, xerrors.Errorf("cannot decode nil bytes")
		}
		return def[0], nil
	}
	return iscp.VMErrorCodeFromBytes(b)
}

func MustDecodeVMErrorCode(b []byte, def ...iscp.VMErrorCode) iscp.VMErrorCode {
	code, err := DecodeVMErrorCode(b, def...)
	if err != nil {
		panic(err)
	}
	return code
}

func EncodeVMErrorCode(code iscp.VMErrorCode) []byte {
	return code.Bytes()
}
