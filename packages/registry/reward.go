package registry

import (
	"github.com/iotaledger/goshimmer/dapps/valuetransfers/packages/address"
)

func GetRewardAddress(scaddr *address.Address) *address.Address {
	//TODO null address means rewards are not paid
	return new(address.Address)
}
