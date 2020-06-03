package testaddresses

import "github.com/iotaledger/goshimmer/dapps/valuetransfers/packages/address"

var (
	specAddresses = []string{
		"ncNRHhJgnukJtoJP9PJZgVQGUqUkZ3jkaxZZko9S5RFn",
		"jooDoUeFxxyqpMkzZi6pQ1tSVZHJSswJ2RgzuHAceBGj",
		"gH8gtohjHeUhzReSvpZoetpGjYjzftxnfR5o56fVaqhv",
	}
	enabledAddress = []bool{true, false, false}
)

func init() {
	if len(specAddresses) != len(enabledAddress) {
		panic("wrong test addresses")
	}
	for _, s := range specAddresses {
		if _, err := address.FromBase58(s); err != nil {
			panic(err)
		}

	}
}

func NumAddresses() int {
	return len(specAddresses)
}

func GetAddress(idx int) (*address.Address, bool) {
	addr, _ := address.FromBase58(specAddresses[idx])
	return &addr, enabledAddress[idx]
}

func MustGetAddress(idx int) *address.Address {
	addr, enabled := GetAddress(idx)
	if !enabled {
		panic("address disabled")
	}
	return addr
}

func IsAddressDisabled(addr address.Address) bool {
	for i, s := range specAddresses {
		if s == addr.String() {
			return !enabledAddress[i]
		}
	}
	return false
}
