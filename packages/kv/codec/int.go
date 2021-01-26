package codec

import (
	"github.com/iotaledger/wasp/packages/util"
)

func DecodeInt64(b []byte) (int64, bool, error) {
	if b == nil {
		// special behavior for backward compatibility: nil value is treated as the absence of a value, not an error
		return 0, false, nil
	}

	r, err := util.Uint64From8Bytes(b)
	return int64(r), err == nil, err
}

func EncodeInt64(value int64) []byte {
	return util.Uint64To8Bytes(uint64(value))
}
