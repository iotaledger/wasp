package codec

import (
	"github.com/iotaledger/wasp/packages/iscp"
)

func DecodeHname(b []byte) (iscp.Hname, bool, error) {
	if b == nil {
		return 0, false, nil
	}
	r, err := iscp.HnameFromBytes(b)
	return r, err == nil, err
}

func EncodeHname(value iscp.Hname) []byte {
	return value.Bytes()
}
