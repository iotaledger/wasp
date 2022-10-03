package codec

import (
	"golang.org/x/xerrors"

	"github.com/iotaledger/wasp/packages/util"
)

func EncodeRatio32(r util.Ratio32) []byte {
	return r.Bytes()
}

func DecodeRatio32(b []byte, def ...util.Ratio32) (util.Ratio32, error) {
	if b == nil {
		if len(def) == 0 {
			return util.Ratio32{}, xerrors.Errorf("cannot decode nil bytes")
		}
		return def[0], nil
	}
	return util.Ratio32FromBytes(b)
}

func MustDecodeRatio32(bytes []byte, def ...util.Ratio32) util.Ratio32 {
	ret, err := DecodeRatio32(bytes, def...)
	if err != nil {
		panic(err)
	}
	return ret
}
