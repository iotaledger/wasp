package committees

import (
	"github.com/iotaledger/goshimmer/dapps/valuetransfers/packages/address"
	"github.com/iotaledger/hive.go/daemon"
	"github.com/iotaledger/hive.go/logger"
	"github.com/iotaledger/hive.go/node"
	"github.com/iotaledger/wasp/packages/committee"
	"github.com/iotaledger/wasp/packages/registry"
	"github.com/iotaledger/wasp/plugins/nodeconn"
	"sync"
)

const PluginName = "Committees"

var (
	Plugin = node.NewPlugin(PluginName, node.Enabled, configure, run)
	log    *logger.Logger

	committeesByAddress = make(map[address.Address]committee.Committee)
	committeesMutex     = &sync.RWMutex{}

	initialLoadWG sync.WaitGroup
)

func init() {
	initialLoadWG.Add(1)
}

func configure(_ *node.Plugin) {
	log = logger.NewLogger(PluginName)
}

func run(_ *node.Plugin) {
	err := daemon.BackgroundWorker(PluginName, func(shutdownSignal <-chan struct{}) {
		lst, err := registry.GetBootupRecords()
		if err != nil {
			log.Error("failed to load bootup records from registry: %v", err)
			return
		}
		log.Debugf("loaded %d bootup record(s) from registry", len(lst))

		addrs := make([]address.Address, 0, len(lst))
		for _, scdata := range lst {
			if cmt := ActivateCommittee(scdata, false); cmt != nil {
				addrs = append(addrs, scdata.Address)
			}
		}
		nodeconn.Subscribe(addrs)
		initialLoadWG.Done()

		<-shutdownSignal

		log.Infof("shutdown signal received: dismissing committees..")
		go func() {
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

func WaitInitialLoad() {
	initialLoadWG.Wait()
}

func ActivateCommittee(bootupData *registry.BootupData, subscribe bool) committee.Committee {
	committeesMutex.Lock()
	defer committeesMutex.Unlock()

	_, ok := committeesByAddress[bootupData.Address]
	if ok {
		log.Warnf("committee already active: %s", bootupData.Address)
		return nil
	}
	c := committee.New(bootupData, log)
	if c != nil {
		committeesByAddress[bootupData.Address] = c
		if subscribe {
			nodeconn.Subscribe([]address.Address{bootupData.Address})
		}
		log.Infof("activated committee for addr %s", bootupData.Address.String())
	}
	return c
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
