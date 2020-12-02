package chains

import (
	"fmt"
	"github.com/iotaledger/goshimmer/dapps/valuetransfers/packages/address"
	"github.com/iotaledger/wasp/packages/coret"
	"sync"

	"github.com/iotaledger/hive.go/daemon"
	"github.com/iotaledger/hive.go/logger"
	"github.com/iotaledger/hive.go/node"
	"github.com/iotaledger/wasp/packages/chain"
	"github.com/iotaledger/wasp/packages/registry"
	"github.com/iotaledger/wasp/plugins/nodeconn"
)

const PluginName = "Chains"

var (
	log *logger.Logger

	chains      = make(map[coret.ChainID]chain.Chain)
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
		chainRecords, err := registry.GetChainRecords()
		if err != nil {
			log.Error("failed to load chain records from registry: %v", err)
			return
		}

		astr := make([]string, len(chainRecords))
		for i := range astr {
			astr[i] = chainRecords[i].ChainID.String()[:10] + ".."
		}
		log.Debugf("loaded %d chain record(s) from registry: %+v", len(chainRecords), astr)

		for _, chr := range chainRecords {
			if chr.Active {
				if err := ActivateChain(chr); err != nil {
					log.Errorf("cannot activate committee %s: %v", chr.ChainID, err)
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
func ActivateChain(chr *registry.ChainRecord) error {
	chainsMutex.Lock()
	defer chainsMutex.Unlock()

	if !chr.Active {
		return fmt.Errorf("cannot activate chain for deactivated chain record")
	}

	_, ok := chains[chr.ChainID]
	if ok {
		log.Debugf("chain is already active: %s", chr.ChainID.String())
		return nil
	}
	// create new chain object
	c := chain.New(chr, log, func() {
		nodeconn.Subscribe((address.Address)(chr.ChainID), chr.Color)
	})
	if c != nil {
		chains[chr.ChainID] = c
		log.Infof("activated chain:\n%s", chr.String())
	} else {
		log.Infof("failed to activate chain:\n%s", chr.String())
	}
	return nil
}

// DeactivateChain deactivates chain in the node
func DeactivateChain(chr *registry.ChainRecord) error {
	chainsMutex.Lock()
	defer chainsMutex.Unlock()

	c, ok := chains[chr.ChainID]
	if !ok || c.IsDismissed() {
		log.Debugf("chain is not active: %s", chr.ChainID.String())
		return nil
	}
	c.Dismiss()
	log.Debugf("chain has been deactivated: %s", chr.ChainID.String())
	return nil
}

// GetChain returns active chain object or nil if it doesn't exist
func GetChain(chainID coret.ChainID) chain.Chain {
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
