package dispatcher

import (
	"github.com/iotaledger/goshimmer/dapps/valuetransfers/packages/address"
	"github.com/iotaledger/goshimmer/dapps/valuetransfers/packages/balance"
	"github.com/iotaledger/goshimmer/plugins/wasp/committee"
	"github.com/iotaledger/goshimmer/plugins/wasp/registry"
	"sync"
)

// unique key for a smart contract is Color of its scid

var (
	scontractsByAddress = make(map[address.Address]committee.Committee)
	scontractsByColor   = make(map[balance.Color]committee.Committee)
	scontractsMutex     = &sync.RWMutex{}
)

func GetSubscriptionList() []address.Address {
	scontractsMutex.RLock()
	defer scontractsMutex.RUnlock()

	ret := make([]address.Address, len(scontractsByAddress))
	i := 0
	for addr := range scontractsByAddress {
		ret[i] = addr
	}
	return ret
}

func loadAllSContracts(ownAddr *registry.PortAddr) ([]address.Address, error) {
	scontractsMutex.Lock()
	defer scontractsMutex.Unlock()

	sclist, err := registry.GetSCDataList(ownAddr)
	if err != nil {
		return nil, err
	}
	addrs := make([]address.Address, 0)
	for _, scdata := range sclist {
		if c, err := committee.New(scdata); err == nil {
			scontractsByAddress[*scdata.Address] = c
			scontractsByColor[*scdata.Color] = c
			addrs = append(addrs, *scdata.Address)
		} else {
			log.Warn(err)
		}
	}
	return addrs, nil
}

func committeeByColor(color *balance.Color) committee.Committee {
	scontractsMutex.RLock()
	defer scontractsMutex.RUnlock()

	ret, _ := scontractsByColor[*color]
	return ret
}

func committeeByAddress(address *address.Address) committee.Committee {
	scontractsMutex.RLock()
	defer scontractsMutex.RUnlock()

	ret, _ := scontractsByAddress[*address]
	return ret
}
