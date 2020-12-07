package codec

import (
	"github.com/iotaledger/wasp/packages/coretypes"
)

func DecodeHname(b []byte) (coretypes.Hname, bool, error) {
	if b == nil {
		return 0, false, nil
	}
	r, err := coretypes.NewHnameFromBytes(b)
	return r, err == nil, err
}

func EncodeHname(value coretypes.Hname) []byte {
	return value.Bytes()
}
