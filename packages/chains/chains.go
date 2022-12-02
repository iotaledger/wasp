// Copyright 2020 IOTA Stiftung
// SPDX-License-Identifier: Apache-2.0

package chains

import (
	"fmt"
	"sync"
	"time"

	"golang.org/x/xerrors"

	"github.com/iotaledger/hive.go/core/logger"
	"github.com/iotaledger/wasp/packages/chain"
	"github.com/iotaledger/wasp/packages/chain/chainimpl"
	"github.com/iotaledger/wasp/packages/chain/consensus/journal"
	"github.com/iotaledger/wasp/packages/database"
	"github.com/iotaledger/wasp/packages/isc"
	"github.com/iotaledger/wasp/packages/metrics"
	"github.com/iotaledger/wasp/packages/metrics/nodeconnmetrics"
	"github.com/iotaledger/wasp/packages/peering"
	"github.com/iotaledger/wasp/packages/registry"
	"github.com/iotaledger/wasp/packages/vm/processors"
)

type Provider func() *Chains

func (chains Provider) ChainProvider() func(chainID *isc.ChainID) chain.Chain {
	return func(chainID *isc.ChainID) chain.Chain {
		return chains().Get(chainID)
	}
}

type ChainProvider func(chainID *isc.ChainID) chain.Chain

type Chains struct {
	log                              *logger.Logger
	nodeConnection                   chain.NodeConnection
	processorConfig                  *processors.Config
	offledgerBroadcastUpToNPeers     int
	offledgerBroadcastInterval       time.Duration
	pullMissingRequestsFromCommittee bool
	networkProvider                  peering.NetworkProvider
	chainStateStoreProvider          database.ChainStateKVStoreProvider
	rawBlocksEnabled                 bool
	rawBlocksDir                     string

	mutex     sync.RWMutex
	allChains map[isc.ChainID]chain.Chain
}

func New(
	log *logger.Logger,
	nodeConnection chain.NodeConnection,
	processorConfig *processors.Config,
	offledgerBroadcastUpToNPeers int,
	offledgerBroadcastInterval time.Duration,
	pullMissingRequestsFromCommittee bool,
	networkProvider peering.NetworkProvider,
	chainStateStoreProvider database.ChainStateKVStoreProvider,
	rawBlocksEnabled bool,
	rawBlocksDir string,
) *Chains {
	ret := &Chains{
		log:                              log,
		allChains:                        make(map[isc.ChainID]chain.Chain),
		nodeConnection:                   nodeConnection,
		processorConfig:                  processorConfig,
		offledgerBroadcastUpToNPeers:     offledgerBroadcastUpToNPeers,
		offledgerBroadcastInterval:       offledgerBroadcastInterval,
		pullMissingRequestsFromCommittee: pullMissingRequestsFromCommittee,
		networkProvider:                  networkProvider,
		chainStateStoreProvider:          chainStateStoreProvider,
		rawBlocksEnabled:                 rawBlocksEnabled,
		rawBlocksDir:                     rawBlocksDir,
	}
	return ret
}

func (c *Chains) Dismiss() {
	c.mutex.RLock()
	defer c.mutex.RUnlock()

	for _, ch := range c.allChains {
		ch.Dismiss("shutdown")
	}
	c.allChains = make(map[isc.ChainID]chain.Chain)
}

func (c *Chains) ActivateAllFromRegistry(
	chainRecordRegistryProvider registry.ChainRecordRegistryProvider,
	dkShareRegistryProvider registry.DKShareRegistryProvider,
	nodeIdentityProvider registry.NodeIdentityProvider,
	consensusJournalRegistryProvider journal.Provider,
	allMetrics *metrics.Metrics,
	w *wal.WAL,
) error {
	var innerErr error
	if err := chainRecordRegistryProvider.ForEachActiveChainRecord(func(chainRecord *registry.ChainRecord) bool {
		if err := c.Activate(chainRecord, dkShareRegistryProvider, nodeIdentityProvider, consensusJournalRegistryProvider, allMetrics, w); err != nil {
			innerErr = fmt.Errorf("cannot activate chain %s: %w", chainRecord.ChainID(), err)
			return false
		}

		return true
	}); err != nil {
		return err
	}

	return innerErr
}

// Activate activates chain on the Wasp node:
// - creates chain object
// - insert it into the runtime registry
// - subscribes for related transactions in the L1 node
func (c *Chains) Activate(
	chr *registry.ChainRecord,
	dkShareRegistryProvider registry.DKShareRegistryProvider,
	nodeIdentityProvider registry.NodeIdentityProvider,
	consensusJournalRegistryProvider journal.Provider,
	allMetrics *metrics.Metrics,
	w *wal.WAL,
) error {
	c.mutex.Lock()
	defer c.mutex.Unlock()

	if !chr.Active {
		return xerrors.Errorf("cannot activate chain for deactivated chain record")
	}

	chainID := chr.ChainID()
	ret, ok := c.allChains[chainID]
	if ok && !ret.IsDismissed() {
		c.log.Debugf("chain is already active: %s", chainID.String())
		return nil
	}

	// create new chain object
	chainStateStore, err := c.chainStateStoreProvider(chainID)
	if err != nil {
		return err
	}

	chainMetrics := allMetrics.NewChainMetrics(&chainID)
	chainWAL, err := w.NewChainWAL(&chainID)
	if err != nil {
		c.log.Debugf("Error creating wal object: %v", err)
		chainWAL = wal.NewDefault()
	}
	newChain := chainimpl.NewChain(
		&chainID,
		c.log,
		c.nodeConnection,
		chainStateStore,
		c.networkProvider,
		dkShareRegistryProvider,
		nodeIdentityProvider,
		c.processorConfig,
		c.offledgerBroadcastUpToNPeers,
		c.offledgerBroadcastInterval,
		c.pullMissingRequestsFromCommittee,
		chainMetrics,
		consensusJournalRegistryProvider,
		chainWAL,
		c.rawBlocksEnabled,
		c.rawBlocksDir,
	)
	if newChain == nil {
		return xerrors.New("Chains.Activate: failed to create chain object")
	}
	c.allChains[chainID] = newChain
	c.log.Infof("activated chain: %s", chainID.String())
	return nil
}

// Deactivate deactivates chain in the node
func (c *Chains) Deactivate(chr *registry.ChainRecord) error {
	c.mutex.Lock()
	defer c.mutex.Unlock()

	chainID := chr.ChainID()
	ch, ok := c.allChains[chainID]
	if !ok || ch.IsDismissed() {
		c.log.Debugf("chain is not active: %s", chainID.String())
		return nil
	}
	ch.Dismiss("deactivate")
	c.nodeConnection.UnregisterChain(&chainID)
	c.log.Debugf("chain has been deactivated: %s", chainID.String())
	return nil
}

// Get returns active chain object or nil if it doesn't exist
// lazy unsubscribing
func (c *Chains) Get(chainID *isc.ChainID, includeDeactivated ...bool) chain.Chain {
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
	return c.nodeConnection.GetMetrics()
}
