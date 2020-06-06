package database

import (
	"github.com/iotaledger/goshimmer/dapps/valuetransfers/packages/address"
	"github.com/iotaledger/hive.go/kvstore"
	"github.com/iotaledger/wasp/plugins/config"
	"github.com/iotaledger/wasp/plugins/testplugins/testaddresses"
)

func deleteTestingPartitions() {
	if config.Node.GetBool(CfgDatabaseInMemory) {
		return
	}

	for i := 0; i < testaddresses.NumAddresses(); i++ {
		addr, _ := testaddresses.GetAddress(i)
		//if err := GetPartition(addr).Clear(); err != nil {
		if err := deletePartition(addr); err != nil {
			log.Debugf("failed to deleted database partition for testing address %s: %v", addr.String(), err)
		} else {
			log.Debugf("successfully deleted database partition for testing address %s", addr.String())
		}
	}
}

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
