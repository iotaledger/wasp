package codec

import (
	"github.com/iotaledger/wasp/packages/util"
	"time"
)

func DecodeTime(b []byte) (time.Time, bool, error) {
	if b == nil {
		// special behavior for backward compatibility: nil value is treated as the absence of a value, not an error
		return time.Time{}, false, nil
	}

	r, err := util.Uint64From8Bytes(b)
	return time.Unix(0, int64(r)), err == nil, err
}

func EncodeTime(value time.Time) []byte {
	return util.Uint64To8Bytes(uint64(value.UnixNano()))
}
