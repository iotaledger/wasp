package jsonable

import (
	"encoding/json"

	"github.com/iotaledger/goshimmer/dapps/valuetransfers/packages/address"
)

// Address is the base58-encoded representation of address.Address
type Address string

func NewAddress(address *address.Address) Address {
	return Address(address.String())
}

func (a Address) MarshalJSON() ([]byte, error) {
	return json.Marshal(string(a))
}

func (a *Address) UnmarshalJSON(b []byte) error {
	var s string
	if err := json.Unmarshal(b, &s); err != nil {
		return err
	}
	_, err := address.FromBase58(s)
	*a = Address(s)
	return err
}

func (a Address) Address() address.Address {
	addr, err := address.FromBase58(string(a))
	if err != nil {
		panic(err)
	}
	return addr
}
