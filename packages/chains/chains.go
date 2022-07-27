// Copyright 2020 IOTA Stiftung
// SPDX-License-Identifier: Apache-2.0

package chains

import (
	"sync"
	"time"

	"github.com/iotaledger/hive.go/logger"
	"github.com/iotaledger/wasp/packages/chain"
	"github.com/iotaledger/wasp/packages/chain/chainimpl"
	"github.com/iotaledger/wasp/packages/database/dbmanager"
	"github.com/iotaledger/wasp/packages/iscp"
	"github.com/iotaledger/wasp/packages/metrics"
	"github.com/iotaledger/wasp/packages/metrics/nodeconnmetrics"
	"github.com/iotaledger/wasp/packages/peering"
	"github.com/iotaledger/wasp/packages/registry"
	"github.com/iotaledger/wasp/packages/vm/processors"
	"github.com/iotaledger/wasp/packages/wal"
	"golang.org/x/xerrors"
)

type Provider func() *Chains

func (chains Provider) ChainProvider() func(chainID *iscp.ChainID) chain.Chain {
	return func(chainID *iscp.ChainID) chain.Chain {
		return chains().Get(chainID)
	}
}

type ChainProvider func(chainID *iscp.ChainID) chain.Chain

type Chains struct {
	mutex                            sync.RWMutex
	log                              *logger.Logger
	allChains                        map[iscp.ChainID]chain.Chain
	nodeConn                         chain.NodeConnection
	processorConfig                  *processors.Config
	offledgerBroadcastUpToNPeers     int
	offledgerBroadcastInterval       time.Duration
	pullMissingRequestsFromCommittee bool
	networkProvider                  peering.NetworkProvider
	getOrCreateKVStore               dbmanager.ChainKVStoreProvider
}

func New(
	log *logger.Logger,
	processorConfig *processors.Config,
	offledgerBroadcastUpToNPeers int,
	offledgerBroadcastInterval time.Duration,
	pullMissingRequestsFromCommittee bool,
	networkProvider peering.NetworkProvider,
	getOrCreateKVStore dbmanager.ChainKVStoreProvider,
) *Chains {
	ret := &Chains{
		log:                              log,
		allChains:                        make(map[iscp.ChainID]chain.Chain),
		processorConfig:                  processorConfig,
		offledgerBroadcastUpToNPeers:     offledgerBroadcastUpToNPeers,
		offledgerBroadcastInterval:       offledgerBroadcastInterval,
		pullMissingRequestsFromCommittee: pullMissingRequestsFromCommittee,
		networkProvider:                  networkProvider,
		getOrCreateKVStore:               getOrCreateKVStore,
	}
	return ret
}

func (c *Chains) Dismiss() {
	c.mutex.RLock()
	defer c.mutex.RUnlock()

	for _, ch := range c.allChains {
		ch.Dismiss("shutdown")
	}
	c.allChains = make(map[iscp.ChainID]chain.Chain)
}

func (c *Chains) SetNodeConn(nodeConn chain.NodeConnection) {
	if c.nodeConn != nil {
		c.log.Panicf("Chains: node conn already set")
	}
	c.nodeConn = nodeConn
}

func (c *Chains) ActivateAllFromRegistry(registryProvider registry.Provider, allMetrics *metrics.Metrics, w *wal.WAL) error {
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
			if err := c.Activate(chr, registryProvider, allMetrics, w); err != nil {
				c.log.Errorf("cannot activate chain %s: %v", chr.ChainID, err)
			}
		}
	}
	return nil
}

// Activate activates chain on the Wasp node:
// - creates chain object
// - insert it into the runtime registry
// - subscribes for related transactions in the L1 node
func (c *Chains) Activate(chr *registry.ChainRecord, registryProvider registry.Provider, allMetrics *metrics.Metrics, w *wal.WAL) error {
	c.mutex.Lock()
	defer c.mutex.Unlock()

	if !chr.Active {
		return xerrors.Errorf("cannot activate chain for deactivated chain record")
	}
	ret, ok := c.allChains[chr.ChainID]
	if ok && !ret.IsDismissed() {
		c.log.Debugf("chain is already active: %s", chr.ChainID.String())
		return nil
	}
	// create new chain object
	defaultRegistry := registryProvider()
	chainKVStore := c.getOrCreateKVStore(&chr.ChainID)
	chainMetrics := allMetrics.NewChainMetrics(&chr.ChainID)
	chainWAL, err := w.NewChainWAL(&chr.ChainID)
	if err != nil {
		c.log.Debugf("Error creating wal object: %v", err)
		chainWAL = wal.NewDefault()
	}
	newChain := chainimpl.NewChain(
		&chr.ChainID,
		c.log,
		c.nodeConn,
		chainKVStore,
		c.networkProvider,
		defaultRegistry,
		c.processorConfig,
		c.offledgerBroadcastUpToNPeers,
		c.offledgerBroadcastInterval,
		c.pullMissingRequestsFromCommittee,
		chainMetrics,
		chainWAL,
	)
	if newChain == nil {
		return xerrors.New("Chains.Activate: failed to create chain object")
	}
	c.allChains[chr.ChainID] = newChain
	c.log.Infof("activated chain: %s", chr.ChainID.String())
	return nil
}

// Deactivate deactivates chain in the node
func (c *Chains) Deactivate(chr *registry.ChainRecord) error {
	c.mutex.Lock()
	defer c.mutex.Unlock()

	ch, ok := c.allChains[chr.ChainID]
	if !ok || ch.IsDismissed() {
		c.log.Debugf("chain is not active: %s", chr.ChainID.String())
		return nil
	}
	ch.Dismiss("deactivate")
	c.nodeConn.UnregisterChain(&chr.ChainID)
	c.log.Debugf("chain has been deactivated: %s", chr.ChainID.String())
	return nil
}

// Get returns active chain object or nil if it doesn't exist
// lazy unsubscribing
func (c *Chains) Get(chainID *iscp.ChainID, includeDeactivated ...bool) chain.Chain {
	c.mutex.RLock()
	defer c.mutex.RUnlock()

	ret, ok := c.allChains[*chainID]

	if len(includeDeactivated) > 0 && includeDeactivated[0] {
		return ret
	}
	if ok && ret.IsDismissed() {
		return nil
	}
	return ret
}

func (c *Chains) GetNodeConnectionMetrics() nodeconnmetrics.NodeConnectionMetrics {
	return c.nodeConn.GetMetrics()
}
