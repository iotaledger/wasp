package codec

import (
	"time"

	"golang.org/x/xerrors"

	"github.com/iotaledger/wasp/packages/util"
)

var zeroUnixNano = time.Time{}.UnixNano()

func init() {
	if zeroUnixNano != -6795364578871345152 {
		panic("inconsistency: zeroUnixNano != -6795364578871345152")
	}
}

func DecodeTime(b []byte, def ...time.Time) (time.Time, error) {
	if b == nil {
		if len(def) == 0 {
			return time.Time{}, xerrors.Errorf("cannot decode nil bytes")
		}
		return def[0], nil
	}
	nanos, err := util.Int64From8Bytes(b)
	if err != nil {
		return time.Time{}, err
	}
	if nanos == 0 {
		return time.Time{}, nil
	}
	return time.Unix(0, nanos), nil
}

var b8 [8]byte

func EncodeTime(value time.Time) []byte {
	nanos := value.UnixNano()
	if nanos == zeroUnixNano {
		return b8[:]
	}
	return util.Int64To8Bytes(nanos)
}
