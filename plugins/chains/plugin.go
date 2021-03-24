package chains

import (
	"fmt"
	"github.com/iotaledger/goshimmer/packages/ledgerstate"
	"sync"

	"github.com/iotaledger/wasp/packages/coretypes"

	"github.com/iotaledger/hive.go/daemon"
	"github.com/iotaledger/hive.go/logger"
	"github.com/iotaledger/hive.go/node"
	"github.com/iotaledger/wasp/packages/chain"
	registry_pkg "github.com/iotaledger/wasp/packages/registry"
	"github.com/iotaledger/wasp/plugins/nodeconn"
	"github.com/iotaledger/wasp/plugins/peering"
	"github.com/iotaledger/wasp/plugins/registry"
)

const PluginName = "Chains"

var (
	log *logger.Logger

	chains      = make(map[[ledgerstate.AddressLength]byte]chain.Chain)
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
		chainRecords, err := registry_pkg.GetChainRecords()
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
					log.Errorf("cannot activate chain %s: %v", chr.ChainID, err)
				}
			}
		}

		<-shutdownSignal

		func() {
			log.Infof("shutdown signal received: dismissing chains..")
			chainsMutex.RLock()
			defer chainsMutex.RUnlock()

			for _, com := range chains {
				com.Dismiss()
			}
			log.Infof("shutdown signal received: dismissing chains.. Done")
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
func ActivateChain(chr *registry_pkg.ChainRecord) error {
	chainsMutex.Lock()
	defer chainsMutex.Unlock()

	if !chr.Active {
		return fmt.Errorf("cannot activate chain for deactivated chain record")
	}
	chainArr := chr.ChainID.Array()
	_, ok := chains[chainArr]
	if ok {
		log.Debugf("chain is already active: %s", chr.ChainID.String())
		return nil
	}
	// create new chain object
	defaultRegistry := registry.DefaultRegistry()
	c := chain.New(chr, log, peering.DefaultNetworkProvider(), defaultRegistry, defaultRegistry, func() {
		nodeconn.NodeConn.Subscribe(chr.ChainID.AliasAddress)
	})
	if c != nil {
		chains[chainArr] = c
		log.Infof("activated chain:\n%s", chr.String())
	} else {
		log.Infof("failed to activate chain:\n%s", chr.String())
	}
	return nil
}

// DeactivateChain deactivates chain in the node
func DeactivateChain(chr *registry_pkg.ChainRecord) error {
	chainsMutex.Lock()
	defer chainsMutex.Unlock()

	c, ok := chains[chr.ChainID.Array()]
	if !ok || c.IsDismissed() {
		log.Debugf("chain is not active: %s", chr.ChainID.String())
		return nil
	}
	c.Dismiss()
	log.Debugf("chain has been deactivated: %s", chr.ChainID.String())
	return nil
}

// GetChain returns active chain object or nil if it doesn't exist
func GetChain(chainID *coretypes.ChainID) chain.Chain {
	chainsMutex.RLock()
	defer chainsMutex.RUnlock()

	addrArr := chainID.Array()
	ret, ok := chains[addrArr]
	if ok && ret.IsDismissed() {
		delete(chains, addrArr)
		nodeconn.NodeConn.Unsubscribe(chainID.AliasAddress)
		return nil
	}
	return ret
}
