package codec

import (
	"github.com/iotaledger/goshimmer/dapps/valuetransfers/packages/address"
)

func DecodeAddress(b []byte) (address.Address, bool, error) {
	if b == nil {
		return address.Address{}, false, nil
	}
	ret, _, err := address.FromBytes(b)
	if err != nil {
		return address.Address{}, false, err
	}
	return ret, true, nil
}

func EncodeAddress(value address.Address) []byte {
	return value.Bytes()
}
