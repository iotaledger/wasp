package codec

import (
	"github.com/iotaledger/wasp/packages/iscp"
	"golang.org/x/xerrors"
)

func DecodeHname(b []byte, def ...iscp.Hname) (iscp.Hname, error) {
	if b == nil {
		if len(def) == 0 {
			return 0, xerrors.Errorf("cannot decode nil bytes")
		}
		return def[0], nil
	}
	return iscp.HnameFromBytes(b)
}

func EncodeHname(value iscp.Hname) []byte {
	return value.Bytes()
}
