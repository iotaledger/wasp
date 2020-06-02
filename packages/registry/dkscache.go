package registry

import (
	"fmt"
	"sync"

	"github.com/iotaledger/goshimmer/dapps/valuetransfers/packages/address"
	"github.com/iotaledger/wasp/packages/tcrypto"
)

var (
	dkscache      = make(map[address.Address]*tcrypto.DKShare)
	dkscacheMutex = &sync.RWMutex{}
)

// GetDKShare retrieves distributed key share from registry or the cache
// returns dkshare, exists flag and error
func GetDKShare(addr *address.Address) (*tcrypto.DKShare, bool, error) {
	dkscacheMutex.RLock()
	ret, ok := dkscache[*addr]
	if ok {
		defer dkscacheMutex.RUnlock()
		return ret, true, nil
	}
	// switching to write lock
	dkscacheMutex.RUnlock()
	dkscacheMutex.Lock()
	defer dkscacheMutex.Unlock()

	var err error
	ok, err = ExistDKShareInRegistry(addr)
	if err != nil {
		return nil, false, err
	}
	if !ok {
		return nil, false, nil
	}
	ks, err := LoadDKShare(addr, false)
	if err != nil {
		return nil, false, err
	}
	dkscache[*addr] = ks
	return ks, true, nil
}

func GetCommittedDKShare(base58addr string) (*tcrypto.DKShare, error) {
	addr, err := address.FromBase58(base58addr)
	if err != nil {
		return nil, err
	}
	if addr.Version() != address.VersionBLS {
		return nil, fmt.Errorf("Not a BLS address: %s", base58addr)
	}
	ks, ok, err := GetDKShare(&addr)
	if err != nil {
		return nil, err
	}
	if !ok || !ks.Committed {
		return nil, fmt.Errorf("Key share not found: %s", base58addr)
	}
	return ks, nil
}
