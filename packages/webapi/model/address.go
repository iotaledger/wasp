package model

import (
	"encoding/json"
)

// Address is the base58-encoded representation of iotago.Address
type Address string

func NewAddress(address iotago.Address) Address {
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
	_, err := iotago.AddressFromBase58EncodedString(s)
	*a = Address(s)
	return err
}

func (a Address) Address() iotago.Address {
	addr, err := iotago.AddressFromBase58EncodedString(string(a))
	if err != nil {
		panic(err)
	}
	return addr
}
