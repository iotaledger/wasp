package testaddresses

import "github.com/iotaledger/goshimmer/dapps/valuetransfers/packages/address"

var (
	specAddresses = []string{
		"goAqoJfcsr5pngfy4h1Xz1hdGV5RRMWd4CHPhWpf48hu",
		"e7ZcV4rubyz8wTJCfvqYDESZ628G327UBHfhGWVGbqqP",
		"mJrb8R5ko5HDTEhNr6YBDVJb5sYon4bDmHDPiwa1WFsf",
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
