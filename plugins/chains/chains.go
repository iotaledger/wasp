package chains

import (
	"fmt"
	"github.com/iotaledger/wasp/packages/chain"
	"github.com/iotaledger/wasp/packages/coretypes"
	registry_pkg "github.com/iotaledger/wasp/packages/registry"
	"github.com/iotaledger/wasp/plugins/nodeconn"
	"github.com/iotaledger/wasp/plugins/peering"
	"github.com/iotaledger/wasp/plugins/registry"
)

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
	_, ok := allChains[chainArr]
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
		allChains[chainArr] = c
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

	c, ok := allChains[chr.ChainID.Array()]
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
	ret, ok := allChains[addrArr]
	if ok && ret.IsDismissed() {
		delete(allChains, addrArr)
		nodeconn.NodeConn.Unsubscribe(chainID.AliasAddress)
		return nil
	}
	return ret
}
