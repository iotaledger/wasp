package codec

import (
	iotago "github.com/iotaledger/iota.go/v3"
)

func DecodeAddress(b []byte, def ...iotago.Address) (iotago.Address, error) {
	panic("refactor me")
	//if b == nil {
	//	if len(def) == 0 {
	//		return nil, xerrors.Errorf("cannot decode nil bytes")
	//	}
	//	return def[0], nil
	//}
	//ret, _, err := iotago.AddressFromBytes(b)
	//return ret, err
}

func EncodeAddress(value iotago.Address) []byte {
	panic("refactor me")
	//return value.Bytes()
}
