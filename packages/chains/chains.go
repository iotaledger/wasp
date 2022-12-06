// Copyright 2020 IOTA Stiftung
// SPDX-License-Identifier: Apache-2.0

package chains

import (
	"context"
	"errors"
	"fmt"
	"sync"
	"time"

	"golang.org/x/xerrors"

	"github.com/iotaledger/hive.go/core/logger"
	"github.com/iotaledger/wasp/packages/chain"
	"github.com/iotaledger/wasp/packages/chain/cmtLog"
	"github.com/iotaledger/wasp/packages/chain/statemanager/smGPA/smGPAUtils"
	"github.com/iotaledger/wasp/packages/database"
	"github.com/iotaledger/wasp/packages/isc"
	"github.com/iotaledger/wasp/packages/metrics/nodeconnmetrics"
	"github.com/iotaledger/wasp/packages/peering"
	"github.com/iotaledger/wasp/packages/registry"
	"github.com/iotaledger/wasp/packages/state"
	"github.com/iotaledger/wasp/packages/vm/processors"
)

type Provider func() *Chains // TODO: Use DI instead of that.

func (chains Provider) ChainProvider() func(chainID *isc.ChainID) chain.Chain {
	return func(chainID *isc.ChainID) chain.Chain {
		return chains().Get(chainID)
	}
}

type ChainProvider func(chainID *isc.ChainID) chain.Chain

type Chains struct {
	ctx                              context.Context
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

	chainRecordRegistryProvider registry.ChainRecordRegistryProvider
	dkShareRegistryProvider     registry.DKShareRegistryProvider
	nodeIdentityProvider        registry.NodeIdentityProvider
	consensusStateCmtLog        cmtLog.Store

	mutex     sync.RWMutex
	allChains map[isc.ChainID]*activeChain
}

type activeChain struct {
	chain      chain.Chain
	cancelFunc context.CancelFunc
}

func New(
	log *logger.Logger,
	nodeConnection chain.NodeConnection,
	processorConfig *processors.Config,
	offledgerBroadcastUpToNPeers int, // TODO: Unused for now.
	offledgerBroadcastInterval time.Duration, // TODO: Unused for now.
	pullMissingRequestsFromCommittee bool, // TODO: Unused for now.
	networkProvider peering.NetworkProvider,
	chainStateStoreProvider database.ChainStateKVStoreProvider,
	rawBlocksEnabled bool,
	rawBlocksDir string,
	chainRecordRegistryProvider registry.ChainRecordRegistryProvider,
	dkShareRegistryProvider registry.DKShareRegistryProvider,
	nodeIdentityProvider registry.NodeIdentityProvider,
	consensusStateCmtLog cmtLog.Store,
) *Chains {
	ret := &Chains{
		log:                              log,
		allChains:                        map[isc.ChainID]*activeChain{},
		nodeConnection:                   nodeConnection,
		processorConfig:                  processorConfig,
		offledgerBroadcastUpToNPeers:     offledgerBroadcastUpToNPeers,
		offledgerBroadcastInterval:       offledgerBroadcastInterval,
		pullMissingRequestsFromCommittee: pullMissingRequestsFromCommittee,
		networkProvider:                  networkProvider,
		chainStateStoreProvider:          chainStateStoreProvider,
		rawBlocksEnabled:                 rawBlocksEnabled,
		rawBlocksDir:                     rawBlocksDir,
		chainRecordRegistryProvider:      chainRecordRegistryProvider,
		dkShareRegistryProvider:          dkShareRegistryProvider,
		nodeIdentityProvider:             nodeIdentityProvider,
		consensusStateCmtLog:             consensusStateCmtLog,
	}
	return ret
}

func (c *Chains) Run(ctx context.Context) error {
	// inline func used to release the lock with defer before calling "activateAllFromRegistry"
	if err := func() error {
		c.mutex.Lock()
		defer c.mutex.Unlock()

		if c.ctx != nil {
			return errors.New("chains already running")
		}
		c.ctx = ctx

		return nil
	}(); err != nil {
		return err
	}

	return c.activateAllFromRegistry() //nolint:contextcheck
}

func (c *Chains) activateAllFromRegistry() error {
	var innerErr error
	if err := c.chainRecordRegistryProvider.ForEachActiveChainRecord(func(chainRecord *registry.ChainRecord) bool {
		chainID := chainRecord.ChainID()
		if err := c.Activate(chainID); err != nil {
			innerErr = fmt.Errorf("cannot activate chain %s: %w", chainRecord.ChainID(), err)
			return false
		}

		return true
	}); err != nil {
		return err
	}

	return innerErr
}

// Activate activates chain on the Wasp node.
func (c *Chains) Activate(chainID isc.ChainID) error {
	c.mutex.Lock()
	defer c.mutex.Unlock()
	if c.ctx == nil {
		return xerrors.Errorf("run chains first")
	}
	//
	// Check, maybe it is already running.
	if _, ok := c.allChains[chainID]; ok {
		c.log.Debugf("Chain %v is already activated", chainID.String())
		return nil
	}
	//
	// Activate the chain in the persistent store, if it is not activated yet.
	chainRecord, err := c.chainRecordRegistryProvider.ChainRecord(chainID)
	if err != nil {
		return xerrors.Errorf("cannot get chain record for %v: %w", chainID, err)
	}
	if !chainRecord.Active {
		if _, err := c.chainRecordRegistryProvider.ActivateChainRecord(chainID); err != nil {
			return xerrors.Errorf("cannot activate chain: %w", err)
		}
	}
	//
	// Load or initialize new chain store.
	chainKVStore, err := c.chainStateStoreProvider(chainID)
	if err != nil {
		return fmt.Errorf("error when creating chain KV store: %w", err)
	}
	chainStore := state.NewStore(chainKVStore)
	chainState, err := chainStore.LatestState()
	if err != nil {
		chainStore = state.InitChainStore(chainKVStore)
	} else {
		chainIDInState, errChainID := chainState.Has(state.KeyChainID)
		if errChainID != nil || !chainIDInState {
			chainStore = state.InitChainStore(chainKVStore)
		}
	}

	var chainWAL smGPAUtils.BlockWAL
	if c.rawBlocksEnabled {
		chainWAL, err = smGPAUtils.NewBlockWAL(c.log, c.rawBlocksDir, &chainID, smGPAUtils.NewBlockWALMetrics())
		if err != nil {
			panic(xerrors.Errorf("cannot create WAL: %w", err))
		}
	} else {
		chainWAL = smGPAUtils.NewEmptyBlockWAL()
	}

	chainCtx, chainCancel := context.WithCancel(c.ctx)
	newChain, err := chain.New(
		chainCtx,
		&chainID,
		chainStore,
		c.nodeConnection,
		c.nodeIdentityProvider.NodeIdentity(),
		c.processorConfig,
		c.dkShareRegistryProvider,
		c.consensusStateCmtLog,
		chainWAL,
		c.networkProvider,
		c.log,
	)
	if err != nil {
		chainCancel()
		return xerrors.Errorf("Chains.Activate: failed to create chain object: %w", err)
	}
	c.allChains[chainID] = &activeChain{
		chain:      newChain,
		cancelFunc: chainCancel,
	}

	c.log.Infof("activated chain: %s", chainID.String())
	return nil
}

// Deactivate chain in the node.
func (c *Chains) Deactivate(chainID isc.ChainID) error {
	c.mutex.Lock()
	defer c.mutex.Unlock()

	if _, err := c.chainRecordRegistryProvider.DeactivateChainRecord(chainID); err != nil {
		return xerrors.Errorf("cannot deactivate chain %v: %w", chainID, err)
	}

	ch, ok := c.allChains[chainID]
	if !ok {
		c.log.Debugf("chain is not active: %s", chainID.String())
		return nil
	}
	ch.cancelFunc()
	delete(c.allChains, chainID)
	c.log.Debugf("chain has been deactivated: %s", chainID.String())
	return nil
}

// Get returns active chain object or nil if it doesn't exist
// lazy unsubscribing
func (c *Chains) Get(chainID *isc.ChainID) chain.Chain {
	c.mutex.RLock()
	defer c.mutex.RUnlock()

	ret, ok := c.allChains[*chainID]
	if !ok {
		return nil
	}
	return ret.chain
}

func (c *Chains) GetNodeConnectionMetrics() nodeconnmetrics.NodeConnectionMetrics {
	return c.nodeConnection.GetMetrics()
}
