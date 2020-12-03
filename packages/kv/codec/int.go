package codec

import (
	"github.com/iotaledger/wasp/packages/util"
)

func DecodeInt64(b []byte) (int64, bool, error) {
	if b == nil {
		return 0, false, nil
	}
	//if len(b) != 8 {
	//	return 0, false, fmt.Errorf("value %s is not an int64", hex.EncodeToString(b))
	//}
	// 3 Dec, 20202 not needed, Uint64From8Bytes does it all
	r, err := util.Uint64From8Bytes(b)
	return int64(r), err == nil, err
}

func EncodeInt64(value int64) []byte {
	return util.Uint64To8Bytes(uint64(value))
}
