package database

import (
	"github.com/iotaledger/wasp/plugins/testplugins/testaddresses"
)

func deleteTestingPartitions() {
	for i := 0; i < testaddresses.NumAddresses(); i++ {
		addr, _ := testaddresses.GetAddress(i)
		if err := deletePartition(addr); err != nil {
			log.Errorf("failed to deleted database partition for testing address %s: %v", addr.String(), err)
		} else {
			log.Infof("successfully deleted database partition for testing address %s", addr.String())
		}
	}
}
