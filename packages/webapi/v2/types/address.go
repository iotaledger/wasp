package types

import (
	"encoding/json"

	iotago "github.com/iotaledger/iota.go/v3"
	"github.com/iotaledger/wasp/packages/parameters"
)

// Address is the string representation of iotago.Address
type Address string

func NewAddress(address iotago.Address) Address {
	return Address(address.Bech32(parameters.L1().Protocol.Bech32HRP))
}

func (a Address) MarshalJSON() ([]byte, error) {
	return json.Marshal(string(a))
}

func (a *Address) UnmarshalJSON(b []byte) error {
	var s string
	if err := json.Unmarshal(b, &s); err != nil {
		return err
	}
	_, _, err := iotago.ParseBech32(s)
	*a = Address(s)
	return err
}

func (a Address) Address() iotago.Address {
	_, addr, err := iotago.ParseBech32(string(a))
	if err != nil {
		panic(err)
	}
	return addr
}
