package chains

import (
	"fmt"
	"github.com/iotaledger/goshimmer/dapps/valuetransfers/packages/address"
	"github.com/iotaledger/wasp/packages/coretypes"
	"sync"

	"github.com/iotaledger/hive.go/daemon"
	"github.com/iotaledger/hive.go/logger"
	"github.com/iotaledger/hive.go/node"
	"github.com/iotaledger/wasp/packages/committee"
	"github.com/iotaledger/wasp/packages/registry"
	"github.com/iotaledger/wasp/plugins/nodeconn"
)

const PluginName = "Chains"

var (
	log *logger.Logger

	chains      = make(map[coretypes.ChainID]committee.Committee)
	chainsMutex = &sync.RWMutex{}
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
			astr[i] = bootupRecords[i].ChainID.String()[:10] + ".."
		}
		log.Debugf("loaded %d bootup record(s) from registry: %+v", len(bootupRecords), astr)

		for _, bootupData := range bootupRecords {
			if bootupData.Active {
				if err := ActivateChain(bootupData); err != nil {
					log.Errorf("cannot activate committee %s: %v", bootupData.ChainID, err)
				}
			}
		}

		<-shutdownSignal

		func() {
			log.Infof("shutdown signal received: dismissing committees..")
			chainsMutex.RLock()
			defer chainsMutex.RUnlock()

			for _, com := range chains {
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

// ActivateChain activates chain on the Wasp node:
// - creates chain object
// - insert it into the runtime registry
// - subscribes for related transactions in he IOTA node
func ActivateChain(bootupData *registry.BootupData) error {
	chainsMutex.Lock()
	defer chainsMutex.Unlock()

	if !bootupData.Active {
		return fmt.Errorf("cannot activate chain for deactivated bootup record")
	}

	_, ok := chains[bootupData.ChainID]
	if ok {
		log.Debugf("chain is already active: %s", bootupData.ChainID.String())
		return nil
	}
	// create new chain object
	c := committee.New(bootupData, log, func() {
		nodeconn.Subscribe((address.Address)(bootupData.ChainID), bootupData.Color)
	})
	if c != nil {
		chains[bootupData.ChainID] = c
		log.Infof("activated chain:\n%s", bootupData.String())
	} else {
		log.Infof("failed to activate chain:\n%s", bootupData.String())
	}
	return nil
}

// DeactivateChain deactivates chain in the node
func DeactivateChain(bootupData *registry.BootupData) error {
	chainsMutex.Lock()
	defer chainsMutex.Unlock()

	c, ok := chains[bootupData.ChainID]
	if !ok || c.IsDismissed() {
		log.Debugf("chain is not active: %s", bootupData.ChainID.String())
		return nil
	}
	c.Dismiss()
	log.Debugf("chain has been deactivated: %s", bootupData.ChainID.String())
	return nil
}

// GetChain returns active chain object or nil if it doesn't exist
func GetChain(chainID coretypes.ChainID) committee.Committee {
	chainsMutex.RLock()
	defer chainsMutex.RUnlock()

	ret, ok := chains[chainID]
	if ok && ret.IsDismissed() {
		delete(chains, chainID)
		nodeconn.Unsubscribe((address.Address)(chainID))
		return nil
	}
	return ret
}
