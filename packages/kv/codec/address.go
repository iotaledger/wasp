package codec

import (
	"github.com/iotaledger/hive.go/serializer/v2"
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
	if len(b) == 0 {
		return nil, xerrors.Errorf("cannot decode address from empty byte slice")
	}
	typeByte := b[0]
	addr, err := iotago.AddressSelector(uint32(typeByte))
	if err != nil {
		return nil, err
	}
	_, err = addr.Deserialize(b, serializer.DeSeriModePerformValidation, nil)
	if err != nil {
		return nil, err
	}
	return addr, nil
}

func EncodeAddress(addr iotago.Address) []byte {
	addressInBytes, err := addr.Serialize(serializer.DeSeriModeNoValidation, nil)
	if err != nil {
		panic("cannot encode address")
	}
	return addressInBytes
}
