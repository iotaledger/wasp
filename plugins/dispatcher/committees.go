package dispatcher

import (
	"github.com/iotaledger/goshimmer/dapps/valuetransfers/packages/address"
	"github.com/iotaledger/goshimmer/dapps/valuetransfers/packages/balance"
	"github.com/iotaledger/wasp/packages/committee"
	"github.com/iotaledger/wasp/packages/registry"
	"sync"
)

// unique key for a smart contract is Color of its scid

var (
	committeesByAddress = make(map[address.Address]committee.Committee)
	committeesByColor   = make(map[balance.Color]committee.Committee)
	committeesMutex     = &sync.RWMutex{}
)

func GetSubscriptionList() []address.Address {
	committeesMutex.RLock()
	defer committeesMutex.RUnlock()

	ret := make([]address.Address, len(committeesByAddress))
	i := 0
	for addr := range committeesByAddress {
		ret[i] = addr
	}
	return ret
}

func loadAllSContracts() ([]address.Address, error) {
	committeesMutex.Lock()
	defer committeesMutex.Unlock()

	sclist, err := registry.GetSCDataList()
	if err != nil {
		return nil, err
	}
	addrs := make([]address.Address, 0)
	for _, scdata := range sclist {
		if c, err := committee.New(scdata, log); err == nil {
			committeesByAddress[scdata.Address] = c
			committeesByColor[scdata.Color] = c
			addrs = append(addrs, scdata.Address)
		} else {
			log.Warn(err)
		}
	}
	return addrs, nil
}

func committeeByColor(color balance.Color) committee.Committee {
	committeesMutex.RLock()
	defer committeesMutex.RUnlock()

	ret, ok := committeesByColor[color]
	if ok && ret.IsDismissed() {
		delete(committeesByAddress, ret.Address())
		delete(committeesByColor, color)
		return nil
	}
	return ret
}

func CommitteeByAddress(addr address.Address) committee.Committee {
	committeesMutex.RLock()
	defer committeesMutex.RUnlock()

	ret, ok := committeesByAddress[addr]
	if ok && ret.IsDismissed() {
		delete(committeesByAddress, addr)
		delete(committeesByColor, ret.Color())
		return nil
	}
	return ret
}
