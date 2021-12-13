package iscp

import (
	iotago "github.com/iotaledger/iota.go/v3"
	"golang.org/x/xerrors"
)

func DecodeOutputID(b []byte, def ...iotago.OutputID) (iotago.OutputID, error) {
	if len(b) != iotago.OutputIDLength {
		if len(def) == 0 {
			return iotago.OutputID{}, xerrors.Errorf("expected OutputID size %d, got %d bytes",
				iotago.OutputIDLength, len(b))
		}
		return def[0], nil
	}
	var ret iotago.OutputID
	copy(ret[:], b)
	return ret, nil
}

func EncodeOutputID(value iotago.OutputID) []byte {
	return value[:]
}
