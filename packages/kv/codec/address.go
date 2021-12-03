package codec

import (
	iotago "github.com/iotaledger/iota.go/v3"
	"golang.org/x/xerrors"
)

func DecodeAddress(b []byte, def ...iotago.Address) (iotago.Address, error) {
	if b == nil {
		if len(def) == 0 {
			return nil, xerrors.Errorf("cannot decode nil bytes")
		}
		return def[0], nil
	}
	ret, _, err := iotago.AddressFromBytes(b)
	return ret, err
}

func EncodeAddress(value iotago.Address) []byte {
	return value.Bytes()
}
