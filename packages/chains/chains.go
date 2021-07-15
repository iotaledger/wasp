package chains

import (
	"sync"
	"time"

	"github.com/iotaledger/goshimmer/packages/ledgerstate"
	txstream "github.com/iotaledger/goshimmer/packages/txstream/client"
	"github.com/iotaledger/hive.go/events"
	"github.com/iotaledger/hive.go/logger"
	"github.com/iotaledger/wasp/packages/chain"
	"github.com/iotaledger/wasp/packages/chain/chainimpl"
	"github.com/iotaledger/wasp/packages/iscp"
	"github.com/iotaledger/wasp/packages/parameters"
	peering_pkg "github.com/iotaledger/wasp/packages/peering"
	"github.com/iotaledger/wasp/packages/registry"
	"github.com/iotaledger/wasp/packages/vm/processors"
	"github.com/iotaledger/wasp/plugins/database"
	"github.com/iotaledger/wasp/plugins/peering"
	"golang.org/x/xerrors"
)

type Provider func() *Chains

type Chains struct {
	mutex                            sync.RWMutex
	log                              *logger.Logger
	allChains                        map[[ledgerstate.AddressLength]byte]chain.Chain
	nodeConn                         *txstream.Client
	processorConfig                  *processors.Config
	offledgerBroadcastUpToNPeers     int
	offledgerBroadcastInterval       time.Duration
	pullMissingRequestsFromCommittee bool
}

func New(
	log *logger.Logger,
	processorConfig *processors.Config,
	offledgerBroadcastUpToNPeers int,
	offledgerBroadcastInterval time.Duration,
	pullMissingRequestsFromCommittee bool,
) *Chains {
	ret := &Chains{
		log:                              log,
		allChains:                        make(map[[ledgerstate.AddressLength]byte]chain.Chain),
		processorConfig:                  processorConfig,
		offledgerBroadcastUpToNPeers:     offledgerBroadcastUpToNPeers,
		offledgerBroadcastInterval:       offledgerBroadcastInterval,
		pullMissingRequestsFromCommittee: pullMissingRequestsFromCommittee,
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

func (c *Chains) ActivateAllFromRegistry(registryProvider registry.Provider) error {
	chainRecords, err := registryProvider().GetChainRecords()
	if err != nil {
		return err
	}

	astr := make([]string, len(chainRecords))
	for i := range astr {
		astr[i] = chainRecords[i].ChainID.String()[:10] + ".."
	}
	c.log.Debugf("loaded %d chain record(s) from registry: %+v", len(chainRecords), astr)

	for _, chr := range chainRecords {
		if chr.Active {
			if err := c.Activate(chr, registryProvider); err != nil {
				c.log.Errorf("cannot activate chain %s: %v", chr.ChainID, err)
			}
		}
	}
	return nil
}

// Activate activates chain on the Wasp node:
// - creates chain object
// - insert it into the runtime registry
// - subscribes for related transactions in he IOTA node
func (c *Chains) Activate(chr *registry.ChainRecord, registryProvider registry.Provider) error {
	c.mutex.Lock()
	defer c.mutex.Unlock()

	if !chr.Active {
		return xerrors.Errorf("cannot activate chain for deactivated chain record")
	}
	chainArr := chr.ChainID.Array()
	_, ok := c.allChains[chainArr]
	if ok {
		c.log.Debugf("chain is already active: %s", chr.ChainID.String())
		return nil
	}
	// create new chain object
	peerNetworkConfig, err := peering_pkg.NewStaticPeerNetworkConfigProvider(
		parameters.GetString(parameters.PeeringMyNetID),
		parameters.GetInt(parameters.PeeringPort),
		chr.Peers...,
	)
	if err != nil {
		return xerrors.Errorf("cannot create peer network config provider")
	}

	defaultRegistry := registryProvider()
	chainKVStore := database.GetOrCreateKVStore(chr.ChainID)
	newChain := chainimpl.NewChain(
		chr.ChainID,
		c.log,
		c.nodeConn,
		peerNetworkConfig,
		chainKVStore,
		peering.DefaultNetworkProvider(),
		defaultRegistry,
		defaultRegistry,
		defaultRegistry,
		c.processorConfig,
		c.offledgerBroadcastUpToNPeers,
		c.offledgerBroadcastInterval,
		c.pullMissingRequestsFromCommittee,
	)
	if newChain == nil {
		return xerrors.New("Chains.Activate: failed to create chain object")
	}
	c.allChains[chainArr] = newChain
	c.nodeConn.Subscribe(chr.ChainID.AliasAddress)
	c.log.Infof("activated chain: %s", chr.ChainID.String())
	return nil
}

// Deactivate deactivates chain in the node
func (c *Chains) Deactivate(chr *registry.ChainRecord) error {
	c.mutex.Lock()
	defer c.mutex.Unlock()

	ch, ok := c.allChains[chr.ChainID.Array()]
	if !ok || ch.IsDismissed() {
		c.log.Debugf("chain is not active: %s", chr.ChainID.String())
		return nil
	}
	ch.Dismiss("deactivate")
	c.nodeConn.Unsubscribe(chr.ChainID.AliasAddress)
	c.log.Debugf("chain has been deactivated: %s", chr.ChainID.String())
	return nil
}

// Get returns active chain object or nil if it doesn't exist
// lazy unsubscribing
func (c *Chains) Get(chainID *iscp.ChainID) chain.Chain {
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
