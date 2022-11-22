package util

import (
	iotago "github.com/iotaledger/iota.go/v3"
	"github.com/iotaledger/wasp/packages/parameters"
)

type ComparableString string

func (c ComparableString) Key() string {
	return string(c)
}

func (c ComparableString) String() string {
	return string(c)
}

type ComparableAddress struct {
	address iotago.Address
}

func NewComparableAddress(address iotago.Address) *ComparableAddress {
	return &ComparableAddress{
		address: address,
	}
}

func (c *ComparableAddress) Address() iotago.Address {
	return c.address
}

func (c *ComparableAddress) Key() string {
	return c.address.Key()
}

func (c *ComparableAddress) String() string {
	return c.address.Bech32(parameters.L1().Protocol.Bech32HRP)
}
