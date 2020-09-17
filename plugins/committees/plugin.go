package committees

import (
	"fmt"
	"sync"

	"github.com/iotaledger/goshimmer/dapps/valuetransfers/packages/address"
	"github.com/iotaledger/hive.go/daemon"
	"github.com/iotaledger/hive.go/logger"
	"github.com/iotaledger/hive.go/node"
	"github.com/iotaledger/wasp/packages/committee"
	"github.com/iotaledger/wasp/packages/registry"
	"github.com/iotaledger/wasp/plugins/nodeconn"
)

const PluginName = "Committees"

var (
	log *logger.Logger

	committeesByAddress = make(map[address.Address]committee.Committee)
	committeesMutex     = &sync.RWMutex{}
)

func Init() *node.Plugin {
	return node.NewPlugin(PluginName, node.Enabled, configure, run)
}

func configure(_ *node.Plugin) {
	log = logger.NewLogger(PluginName)
}

func run(_ *node.Plugin) {
	err := daemon.BackgroundWorker(PluginName, func(shutdownSignal <-chan struct{}) {
		bootupRecords, err := registry.GetBootupRecords()
		if err != nil {
			log.Error("failed to load bootup records from registry: %v", err)
			return
		}

		astr := make([]string, len(bootupRecords))
		for i := range astr {
			astr[i] = bootupRecords[i].Address.String()[:10] + ".."
		}
		log.Debugf("loaded %d bootup record(s) from registry: %+v", len(bootupRecords), astr)

		for _, bootupData := range bootupRecords {
			if bootupData.Active {
				if err := ActivateCommittee(bootupData); err != nil {
					log.Errorf("cannot activate committee %s: %v", bootupData.Address, err)
				}
			}
		}

		<-shutdownSignal

		func() {
			log.Infof("shutdown signal received: dismissing committees..")
			committeesMutex.RLock()
			defer committeesMutex.RUnlock()

			for _, com := range committeesByAddress {
				com.Dismiss()
			}
			log.Infof("shutdown signal received: dismissing committees.. Done")
		}()
	})
	if err != nil {
		log.Error(err)
		return
	}
}

func ActivateCommittee(bootupData *registry.BootupData) error {
	committeesMutex.Lock()
	defer committeesMutex.Unlock()

	if !bootupData.Active {
		return fmt.Errorf("cannot activate committee for inactive SC")
	}

	_, ok := committeesByAddress[bootupData.Address]
	if ok {
		log.Debugf("committee already active: %s", bootupData.Address)
		return nil
	}
	c := committee.New(bootupData, log, func() {
		nodeconn.Subscribe(bootupData.Address, bootupData.Color)
	})
	if c != nil {
		committeesByAddress[bootupData.Address] = c
		log.Infof("activated smart contract:\n%s", bootupData.String())
	} else {
		log.Infof("failed to activate smart contract:\n%s", bootupData.String())
	}
	return nil
}

func DeactivateCommittee(bootupData *registry.BootupData) error {
	committeesMutex.Lock()
	defer committeesMutex.Unlock()

	if bootupData.Active {
		return fmt.Errorf("cannot deactivate committee for active SC")
	}

	c, ok := committeesByAddress[bootupData.Address]
	if !ok || c.IsDismissed() {
		log.Debugf("committee already inactive: %s", bootupData.Address)
		return nil
	}
	c.Dismiss()
	return nil
}

func CommitteeByAddress(addr address.Address) committee.Committee {
	committeesMutex.RLock()
	defer committeesMutex.RUnlock()

	ret, ok := committeesByAddress[addr]
	if ok && ret.IsDismissed() {
		delete(committeesByAddress, addr)
		nodeconn.Unsubscribe(addr)
		return nil
	}
	return ret
}

func iterateCommittees(f func(c committee.Committee)) {
	committeesMutex.RLock()
	defer committeesMutex.RUnlock()

	for _, c := range committeesByAddress {
		if !c.IsDismissed() {
			f(c)
		}
	}
}
