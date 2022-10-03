package codec

import (
	"golang.org/x/xerrors"

	"github.com/iotaledger/wasp/packages/isc"
)

func DecodeHname(b []byte, def ...isc.Hname) (isc.Hname, error) {
	if b == nil {
		if len(def) == 0 {
			return 0, xerrors.Errorf("cannot decode nil bytes")
		}
		return def[0], nil
	}
	return isc.HnameFromBytes(b)
}

func MustDecodeHname(b []byte, def ...isc.Hname) isc.Hname {
	r, err := DecodeHname(b, def...)
	if err != nil {
		panic(err)
	}
	return r
}

func EncodeHname(value isc.Hname) []byte {
	return value.Bytes()
}
