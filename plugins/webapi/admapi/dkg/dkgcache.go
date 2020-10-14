// handles dkgCache used during DKG process
package dkg

import (
	"fmt"
	"github.com/iotaledger/wasp/packages/tcrypto"
	"sync"
)

var (
	dkgCache      = make(map[int]*tcrypto.DKShare)
	dkgCacheMutex = &sync.RWMutex{}
)

func getFromDkgCache(tmpId int) *tcrypto.DKShare {
	dkgCacheMutex.RLock()
	defer dkgCacheMutex.RUnlock()
	ret, _ := dkgCache[tmpId]
	return ret
}

func putToDkgCache(tmpId int, dkshare *tcrypto.DKShare) error {
	dkgCacheMutex.Lock()
	defer dkgCacheMutex.Unlock()

	if dkshare == nil {
		delete(dkgCache, tmpId)
		return nil
	}
	if _, ok := dkgCache[tmpId]; ok {
		return fmt.Errorf("duplicate tmpId %d during DKG", tmpId)
	}
	dkgCache[tmpId] = dkshare
	return nil
}
