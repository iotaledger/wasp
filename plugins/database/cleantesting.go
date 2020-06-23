package database

import (
	"github.com/iotaledger/goshimmer/dapps/valuetransfers/packages/address"
	"github.com/iotaledger/hive.go/kvstore"
)

// temporary. Until DeletePrefix is fixed
func deletePartition(addr *address.Address) error {
	dbase := GetPartition(addr)
	keys := make([][]byte, 0)
	err := dbase.IterateKeys(kvstore.EmptyPrefix, func(key kvstore.Key) bool {
		k := make([]byte, len(key))
		copy(k, key)
		keys = append(keys, k)
		return true
	})
	if err != nil {
		return err
	}
	b := dbase.Batched()
	for _, k := range keys {
		if err = b.Delete(k); err != nil {
			b.Cancel()
			return err
		}
	}
	return b.Commit()
}
