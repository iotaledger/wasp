package committees

import (
	"fmt"
	"github.com/iotaledger/goshimmer/dapps/valuetransfers/packages/address"
	"github.com/iotaledger/goshimmer/dapps/valuetransfers/packages/balance"
	"github.com/iotaledger/hive.go/daemon"
	"github.com/iotaledger/hive.go/logger"
	"github.com/iotaledger/hive.go/node"
	"github.com/iotaledger/wasp/packages/committee"
	"github.com/iotaledger/wasp/packages/registry"
	"github.com/iotaledger/wasp/plugins/nodeconn"
	"github.com/iotaledger/wasp/plugins/testplugins/testaddresses"
	"sync"
)

// PluginName is the name of the config plugin.
const PluginName = "Committees"

var (
	// Plugin is the plugin instance of the config plugin.
	Plugin = node.NewPlugin(PluginName, node.Enabled, configure, run)
	log    *logger.Logger

	committeesByAddress = make(map[address.Address]committee.Committee)
	committeesByColor   = make(map[balance.Color]committee.Committee)
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
		sclist, err := registry.GetSCDataList()
		if err != nil {
			log.Error("failed to load SC meta data from registry: %v", err)
			return
		}
		log.Debugf("loaded %d SC data record(s) from registry", len(sclist))

		addrs := make([]address.Address, 0, len(sclist))
		for _, scdata := range sclist {
			if testaddresses.IsAddressDisabled(scdata.Address) {
				log.Debugf("skipping disabled address %s", scdata.Address.String())
				continue
			}
			if err := RegisterCommittee(scdata, false); err != nil {
				log.Warn(err)
			} else {
				log.Infof("SC registered: addr %s color %s", scdata.Address.String(), scdata.Color.String())
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

func RegisterCommittee(scdata *registry.SCMetaData, subscribe bool) error {
	committeesMutex.Lock()
	defer committeesMutex.Unlock()

	_, ok := committeesByAddress[scdata.Address]
	if ok {
		return fmt.Errorf("committee already registered: %s", scdata.Address)
	}
	if c, err := committee.New(scdata, log); err == nil {
		committeesByAddress[scdata.Address] = c
		committeesByColor[scdata.Color] = c
		if subscribe {
			nodeconn.Subscribe([]address.Address{scdata.Address})
		}
	} else {
		return err
	}

	return nil
}

func CommitteeByColor(color balance.Color) committee.Committee {
	committeesMutex.RLock()
	defer committeesMutex.RUnlock()

	ret, ok := committeesByColor[color]
	if ok && ret.IsDismissed() {
		delete(committeesByAddress, ret.Address())
		delete(committeesByColor, color)
		nodeconn.Unsubscribe(ret.Address())
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
		nodeconn.Unsubscribe(addr)
		return nil
	}
	return ret
}
