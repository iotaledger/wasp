package database

import (
	"github.com/iotaledger/hive.go/kvstore"
	"github.com/iotaledger/wasp/plugins/testplugins/testaddresses"
	"github.com/mr-tron/base58"
)

func deleteTestingPartitions() {
	for i := 0; i < testaddresses.NumAddresses(); i++ {
		addr, _ := testaddresses.GetAddress(i)
		if err := GetPartition(addr).Clear(); err != nil {
			log.Errorf("failed to deleted database partition for testing address %s: %v", addr.String(), err)
		} else {
			log.Infof("successfully deleted database partition for testing address %s", addr.String())
		}
	}
	// checking if state exists
	for i := 0; i < testaddresses.NumAddresses(); i++ {
		addr, _ := testaddresses.GetAddress(i)
		if exist, err := KeyExistInPartition(addr, ObjectTypeVariableState); err != nil {
			log.Errorf("checking if state exists for address %s: %v", addr.String(), err)
		} else {
			log.Infof("after cleanup of %s: state exist = %v", addr.String(), exist)
			if exist {
				_ = GetPartition(addr).Iterate(kvstore.EmptyPrefix, func(key kvstore.Key, value kvstore.Value) bool {
					log.Debugw("after cleanup", "type", key[0], "key", base58.Encode(key), "value", base58.Encode(value))
					return true
				})
			}
		}
	}

}
