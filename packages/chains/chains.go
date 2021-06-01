package chains

import (
	"sync"

	txstream "github.com/iotaledger/goshimmer/packages/txstream/client"

	"golang.org/x/xerrors"

	"github.com/iotaledger/goshimmer/packages/ledgerstate"
	"github.com/iotaledger/hive.go/events"
	"github.com/iotaledger/hive.go/logger"
	"github.com/iotaledger/wasp/packages/chain"
	"github.com/iotaledger/wasp/packages/chain/chainimpl"
	"github.com/iotaledger/wasp/packages/coretypes"
	"github.com/iotaledger/wasp/packages/registry/chainrecord"
	"github.com/iotaledger/wasp/plugins/database"
	"github.com/iotaledger/wasp/plugins/peering"
	"github.com/iotaledger/wasp/plugins/registry"
)

type Chains struct {
	mutex     sync.RWMutex
	log       *logger.Logger
	allChains map[[ledgerstate.AddressLength]byte]chain.Chain
	nodeConn  *txstream.Client
}

func New(log *logger.Logger) *Chains {
	ret := &Chains{
		log:       log,
		allChains: make(map[[ledgerstate.AddressLength]byte]chain.Chain),
	}
	return ret
}

func (c *Chains) Dismiss() {
	c.mutex.RLock()
	defer c.mutex.RUnlock()

	for _, ch := range c.allChains {
		ch.Dismiss("shutdown")
	}
	c.allChains = make(map[[ledgerstate.AddressLength]byte]chain.Chain)
}

func (c *Chains) Attach(nodeConn *txstream.Client) {
	if c.nodeConn != nil {
		c.log.Panicf("Chains: already attached")
	}
	c.nodeConn = nodeConn
	c.nodeConn.Events.TransactionReceived.Attach(events.NewClosure(c.dispatchTransactionMsg))
	c.nodeConn.Events.InclusionStateReceived.Attach(events.NewClosure(c.dispatchInclusionStateMsg))
	c.nodeConn.Events.OutputReceived.Attach(events.NewClosure(c.dispatchOutputMsg))
	c.nodeConn.Events.UnspentAliasOutputReceived.Attach(events.NewClosure(c.dispatchUnspentAliasOutputMsg))
	// TODO attach to off-ledger request module
}

func (c *Chains) ActivateAllFromRegistry(chainRecordProvider coretypes.ChainRecordRegistryProvider) error {
	chainRecords, err := chainRecordProvider.GetChainRecords()
	if err != nil {
		return err
	}

	astr := make([]string, len(chainRecords))
	for i := range astr {
		astr[i] = coretypes.NewChainID(chainRecords[i].ChainAddr).String()[:10] + ".."
	}
	c.log.Debugf("loaded %d chain record(s) from registry: %+v", len(chainRecords), astr)

	for _, chr := range chainRecords {
		if chr.Active {
			if err := c.Activate(chr); err != nil {
				c.log.Errorf("cannot activate chain %s: %v", chr.ChainAddr, err)
			}
		}
	}
	return nil
}

// Activate activates chain on the Wasp node:
// - creates chain object
// - insert it into the runtime registry
// - subscribes for related transactions in he IOTA node
func (c *Chains) Activate(chr *chainrecord.ChainRecord) error {
	c.mutex.Lock()
	defer c.mutex.Unlock()

	if !chr.Active {
		return xerrors.Errorf("cannot activate chain for deactivated chain record")
	}
	chainArr := chr.ChainAddr.Array()
	_, ok := c.allChains[chainArr]
	if ok {
		c.log.Debugf("chain is already active: %s", chr.ChainAddr.String())
		return nil
	}
	// create new chain object
	defaultRegistry := registry.DefaultRegistry()
	chainKVStore := database.GetOrCreateKVStore(chr.ChainAddr)
	newChain := chainimpl.NewChain(
		chr,
		c.log,
		c.nodeConn,
		peering.DefaultPeerNetworkConfig(),
		chainKVStore,
		peering.DefaultNetworkProvider(),
		defaultRegistry,
		defaultRegistry,
		defaultRegistry,
	)
	if newChain == nil {
		return xerrors.New("Chains.Activate: failed to create chain object")
	}
	c.allChains[chainArr] = newChain
	c.nodeConn.Subscribe(chr.ChainAddr)
	c.log.Infof("activated chain: %s", chr.ChainAddr.String())
	return nil
}

// Deactivate deactivates chain in the node
func (c *Chains) Deactivate(chr *chainrecord.ChainRecord) error {
	c.mutex.Lock()
	defer c.mutex.Unlock()

	ch, ok := c.allChains[chr.ChainAddr.Array()]
	if !ok || ch.IsDismissed() {
		c.log.Debugf("chain is not active: %s", chr.ChainAddr.String())
		return nil
	}
	ch.Dismiss("deactivate")
	c.nodeConn.Unsubscribe(chr.ChainAddr)
	c.log.Debugf("chain has been deactivated: %s", chr.ChainAddr.String())
	return nil
}

// Get returns active chain object or nil if it doesn't exist
// lazy unsubscribing
func (c *Chains) Get(chainID *coretypes.ChainID) chain.Chain {
	c.mutex.RLock()
	defer c.mutex.RUnlock()

	addrArr := chainID.Array()
	ret, ok := c.allChains[addrArr]
	if ok && ret.IsDismissed() {
		delete(c.allChains, addrArr)
		c.nodeConn.Unsubscribe(chainID.AliasAddress)
		return nil
	}
	return ret
}
