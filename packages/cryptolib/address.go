package cryptolib

import (
	iotago "github.com/iotaledger/iota.go/v3"
	"github.com/iotaledger/iota.go/v3/bech32"
)

const AddressSize = 32

type Address [AddressSize]byte

func newAddressFromBytes(addr [AddressSize]byte) *Address {
	result := Address(addr)
	return &result
}

func (a *Address) Equals(other *Address) bool {
	return *a == *other
}

func (a *Address) Bytes() []byte {
	return a[:]
}

func (a *Address) Bech32(hrp iotago.NetworkPrefix) string {
	s, err := bech32.Encode(string(hrp), a.Bytes())
	if err != nil {
		panic(err)
	}
	return s
}

func (a *Address) String() string {
	return EncodeHex(a.Bytes())
}
