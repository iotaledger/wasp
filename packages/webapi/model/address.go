package model

import (
	"encoding/json"
	"github.com/iotaledger/goshimmer/packages/ledgerstate"
)

// Address is the base58-encoded representation of ledgerstate.Address
type Address string

func NewAddress(address ledgerstate.Address) Address {
	return Address(address.Base58())
}

func (a Address) MarshalJSON() ([]byte, error) {
	return json.Marshal(string(a))
}

func (a *Address) UnmarshalJSON(b []byte) error {
	var s string
	if err := json.Unmarshal(b, &s); err != nil {
		return err
	}
	_, err := ledgerstate.AddressFromBase58EncodedString(s)
	*a = Address(s)
	return err
}

func (a Address) Address() ledgerstate.Address {
	addr, err := ledgerstate.AddressFromBase58EncodedString(string(a))
	if err != nil {
		panic(err)
	}
	return addr
}
