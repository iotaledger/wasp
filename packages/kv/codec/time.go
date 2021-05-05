package codec

import (
	"time"

	"github.com/iotaledger/wasp/packages/util"
)

var zeroUnixNano = time.Time{}.UnixNano()

func init() {
	if zeroUnixNano != -6795364578871345152 {
		panic("inconsistency: zeroUnixNano != -6795364578871345152")
	}
}

func DecodeTime(b []byte) (time.Time, bool, error) {
	if b == nil {
		// special behavior for backward compatibility: nil value is treated as the absence of a value, not an error
		return time.Time{}, false, nil
	}
	nanos, err := util.Int64From8Bytes(b)
	if err != nil {
		return time.Time{}, false, err
	}
	if nanos == 0 {
		return time.Time{}, true, nil
	}
	return time.Unix(0, nanos), true, nil
}

var b8 [8]byte

func EncodeTime(value time.Time) []byte {
	nanos := value.UnixNano()
	if nanos == zeroUnixNano {
		return b8[:]
	}
	return util.Int64To8Bytes(nanos)
}
