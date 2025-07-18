package util

import (
	"github.com/iotaledger/wasp/v2/packages/cryptolib"
)

type ComparableString string

func (c ComparableString) Key() string {
	return string(c)
}

func (c ComparableString) String() string {
	return string(c)
}

type ComparableAddress struct {
	address *cryptolib.Address
}

func NewComparableAddress(address *cryptolib.Address) *ComparableAddress {
	return &ComparableAddress{
		address: address,
	}
}

func (c *ComparableAddress) Address() *cryptolib.Address {
	return c.address
}

func (c *ComparableAddress) Key() cryptolib.AddressKey {
	return c.address.Key()
}

func (c *ComparableAddress) String() string {
	return c.address.String()
}
