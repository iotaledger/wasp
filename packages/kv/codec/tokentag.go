package codec

import (
	iotago "github.com/iotaledger/iota.go/v3"
	"golang.org/x/xerrors"
)

func DecodeTokenTag(b []byte, def ...iotago.TokenTag) (iotago.TokenTag, error) {
	if len(b) != iotago.TokenTagLength {
		if len(def) == 0 {
			return iotago.TokenTag{}, xerrors.Errorf("wrong data length")
		}
		return def[0], nil
	}
	var ret iotago.TokenTag
	copy(ret[:], b)
	return ret, nil
}

func EncodeTokenTag(value iotago.TokenTag) []byte {
	return value[:]
}
