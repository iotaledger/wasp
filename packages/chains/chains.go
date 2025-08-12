// Copyright 2020 IOTA Stiftung
// SPDX-License-Identifier: Apache-2.0

// Package chains provides functionality for managing multiple blockchain instances.
package chains

import (
	"context"
	"errors"
	"fmt"
	"sync"
	"time"

	"github.com/iotaledger/hive.go/lo"
	"github.com/iotaledger/hive.go/log"

	"github.com/iotaledger/wasp/v2/packages/chain"
	"github.com/iotaledger/wasp/v2/packages/chain/cmtlog"
	"github.com/iotaledger/wasp/v2/packages/chain/mempool"
	"github.com/iotaledger/wasp/v2/packages/chain/statemanager/gpa"
	"github.com/iotaledger/wasp/v2/packages/chain/statemanager/gpa/utils"
	"github.com/iotaledger/wasp/v2/packages/chain/statemanager/snapshots"
	"github.com/iotaledger/wasp/v2/packages/chains/accessmanager"
	"github.com/iotaledger/wasp/v2/packages/cryptolib"
	"github.com/iotaledger/wasp/v2/packages/database"
	"github.com/iotaledger/wasp/v2/packages/isc"
	"github.com/iotaledger/wasp/v2/packages/metrics"
	"github.com/iotaledger/wasp/v2/packages/peering"
	"github.com/iotaledger/wasp/v2/packages/registry"
	"github.com/iotaledger/wasp/v2/packages/shutdown"
	"github.com/iotaledger/wasp/v2/packages/state"
	"github.com/iotaledger/wasp/v2/packages/state/indexedstore"
	"github.com/iotaledger/wasp/v2/packages/util"
	"github.com/iotaledger/wasp/v2/packages/vm/core/accounts"
	"github.com/iotaledger/wasp/v2/packages/vm/processors"
	"github.com/iotaledger/wasp/v2/packages/webapi/interfaces"
)

type Provider func() *ChainRunner // TODO: Use DI instead of that.

type ChainProvider func() chain.Chain

type ChainRunner struct {
	ctx                        context.Context
	log                        log.Logger
	nodeConnection             chain.NodeConnection
	processorConfig            *processors.Config
	deriveAliasOutputByQuorum  bool
	pipeliningLimit            int
	postponeRecoveryMilestones int
	consensusDelay             time.Duration
	recoveryTimeout            time.Duration

	networkProvider              peering.NetworkProvider
	trustedNetworkManager        peering.TrustedNetworkManager
	trustedNetworkListenerCancel context.CancelFunc
	chainStateStoreProvider      database.ChainStateKVStoreProvider

	walLoadToStore                      bool
	walEnabled                          bool
	walFolderPath                       string
	smBlockCacheMaxSize                 int
	smBlockCacheBlocksInCacheDuration   time.Duration
	smBlockCacheBlockCleaningPeriod     time.Duration
	smStateManagerGetBlockNodeCount     int
	smStateManagerGetBlockRetry         time.Duration
	smStateManagerRequestCleaningPeriod time.Duration
	smStateManagerStatusLogPeriod       time.Duration
	smStateManagerTimerTickPeriod       time.Duration
	smPruningMinStatesToKeep            int
	smPruningMaxStatesToDelete          int
	snapshotToLoad                      *state.BlockHash
	snapshotPeriod                      uint32
	snapshotDelay                       uint32
	snapshotFolderPath                  string
	snapshotNetworkPaths                []string

	chainRecordRegistryProvider registry.ChainRecordRegistryProvider
	dkShareRegistryProvider     registry.DKShareRegistryProvider
	nodeIdentityProvider        registry.NodeIdentityProvider
	consensusStateRegistry      cmtlog.ConsensusStateRegistry
	chainListener               chain.ChainListener

	mutex           *sync.RWMutex
	chain           chain.Chain
	chainCancelFunc context.CancelFunc
	accessMgr       accessmanager.AccessMgr

	cleanupFunc         context.CancelFunc
	shutdownCoordinator *shutdown.Coordinator

	chainMetricsProvider *metrics.ChainMetricsProvider

	validatorFeeAddr *cryptolib.Address

	mempoolSettings          mempool.Settings
	mempoolBroadcastInterval time.Duration
}

func New(
	log log.Logger,
	nodeConnection chain.NodeConnection,
	processorConfig *processors.Config,
	validatorAddrStr string,
	deriveAliasOutputByQuorum bool,
	pipeliningLimit int,
	postponeRecoveryMilestones int,
	consensusDelay time.Duration,
	recoveryTimeout time.Duration,
	networkProvider peering.NetworkProvider,
	trustedNetworkManager peering.TrustedNetworkManager,
	chainStateStoreProvider database.ChainStateKVStoreProvider,
	walLoadToStore bool,
	walEnabled bool,
	walFolderPath string,
	smBlockCacheMaxSize int,
	smBlockCacheBlocksInCacheDuration time.Duration,
	smBlockCacheBlockCleaningPeriod time.Duration,
	smStateManagerGetBlockNodeCount int,
	smStateManagerGetBlockRetry time.Duration,
	smStateManagerRequestCleaningPeriod time.Duration,
	smStateManagerStatusLogPeriod time.Duration,
	smStateManagerTimerTickPeriod time.Duration,
	smPruningMinStatesToKeep int,
	smPruningMaxStatesToDelete int,
	snapshotToLoad string,
	snapshotPeriod uint32,
	snapshotDelay uint32,
	snapshotFolderPath string,
	snapshotNetworkPaths []string,
	chainRecordRegistryProvider registry.ChainRecordRegistryProvider,
	dkShareRegistryProvider registry.DKShareRegistryProvider,
	nodeIdentityProvider registry.NodeIdentityProvider,
	consensusStateRegistry cmtlog.ConsensusStateRegistry,
	chainListener chain.ChainListener,
	mempoolSettings mempool.Settings,
	mempoolBroadcastInterval time.Duration,
	shutdownCoordinator *shutdown.Coordinator,
	chainMetricsProvider *metrics.ChainMetricsProvider,
) *ChainRunner {
	var validatorFeeAddr *cryptolib.Address
	if validatorAddrStr != "" {
		addr, err := cryptolib.NewAddressFromHexString(validatorAddrStr)
		if err != nil {
			panic(fmt.Errorf("error parsing validator.address: %s", err.Error()))
		}
		validatorFeeAddr = addr
	}
	ret := &ChainRunner{
		log:                                 log,
		mutex:                               &sync.RWMutex{},
		nodeConnection:                      nodeConnection,
		processorConfig:                     processorConfig,
		deriveAliasOutputByQuorum:           deriveAliasOutputByQuorum,
		pipeliningLimit:                     pipeliningLimit,
		consensusDelay:                      consensusDelay,
		recoveryTimeout:                     recoveryTimeout,
		networkProvider:                     networkProvider,
		trustedNetworkManager:               trustedNetworkManager,
		chainStateStoreProvider:             chainStateStoreProvider,
		walLoadToStore:                      walLoadToStore,
		walEnabled:                          walEnabled,
		walFolderPath:                       walFolderPath,
		smBlockCacheMaxSize:                 smBlockCacheMaxSize,
		smBlockCacheBlocksInCacheDuration:   smBlockCacheBlocksInCacheDuration,
		smBlockCacheBlockCleaningPeriod:     smBlockCacheBlockCleaningPeriod,
		smStateManagerGetBlockNodeCount:     smStateManagerGetBlockNodeCount,
		smStateManagerGetBlockRetry:         smStateManagerGetBlockRetry,
		smStateManagerRequestCleaningPeriod: smStateManagerRequestCleaningPeriod,
		smStateManagerStatusLogPeriod:       smStateManagerStatusLogPeriod,
		smStateManagerTimerTickPeriod:       smStateManagerTimerTickPeriod,
		smPruningMinStatesToKeep:            smPruningMinStatesToKeep,
		smPruningMaxStatesToDelete:          smPruningMaxStatesToDelete,
		snapshotPeriod:                      snapshotPeriod,
		snapshotDelay:                       snapshotDelay,
		snapshotFolderPath:                  snapshotFolderPath,
		snapshotNetworkPaths:                snapshotNetworkPaths,
		chainRecordRegistryProvider:         chainRecordRegistryProvider,
		dkShareRegistryProvider:             dkShareRegistryProvider,
		nodeIdentityProvider:                nodeIdentityProvider,
		chainListener:                       nil, // See bellow.
		mempoolSettings:                     mempoolSettings,
		mempoolBroadcastInterval:            mempoolBroadcastInterval,
		consensusStateRegistry:              consensusStateRegistry,
		shutdownCoordinator:                 shutdownCoordinator,
		chainMetricsProvider:                chainMetricsProvider,
		validatorFeeAddr:                    validatorFeeAddr,
	}
	ret.initSnapshotToLoad(snapshotToLoad)
	ret.chainListener = NewChainsListener(chainListener, ret.chainAccessUpdatedCB)
	return ret
}

func (c *ChainRunner) initSnapshotToLoad(config string) {
	c.snapshotToLoad = nil
	blockHash, err := state.BlockHashFromString(config)
	if err != nil {
		c.log.LogErrorf("Parsing snapshots to load: %s is not a block hash: %v", config, err)
		return
	}
	c.snapshotToLoad = &blockHash
}

func (c *ChainRunner) Run(ctx context.Context) error {
	if err := c.nodeConnection.WaitUntilInitiallySynced(ctx); err != nil {
		return fmt.Errorf("waiting for L1 node to become sync failed, error: %w", err)
	}

	c.mutex.Lock()
	defer c.mutex.Unlock()

	if c.ctx != nil {
		return errors.New("chains already running")
	}
	c.ctx = ctx

	c.accessMgr = accessmanager.New(ctx, c.chainServersUpdatedCB, c.nodeIdentityProvider.NodeIdentity(), c.networkProvider, c.log.NewChildLogger("AM"))
	c.trustedNetworkListenerCancel = c.trustedNetworkManager.TrustedPeersListener(c.trustedPeersUpdatedCB)

	unhook := c.chainRecordRegistryProvider.Events().ChainRecordModified.Hook(func(event *registry.ChainRecordModifiedEvent) {
		c.mutex.RLock()
		defer c.mutex.RUnlock()
		if c.chain != nil {
			c.chain.ConfigUpdated(event.ChainRecord.AccessNodes)
		}
	}).Unhook
	c.cleanupFunc = unhook

	return c.activateWithoutLocking() //nolint:contextcheck
}

func (c *ChainRunner) Close() {
	util.ExecuteIfNotNil(c.cleanupFunc)
	c.mutex.RLock()
	defer c.mutex.RUnlock()
	c.chainCancelFunc()
	c.shutdownCoordinator.WaitNestedWithLogging(1 * time.Second)
	c.shutdownCoordinator.Done()
	util.ExecuteIfNotNil(c.trustedNetworkListenerCancel)
	c.trustedNetworkListenerCancel = nil
}

func (c *ChainRunner) trustedPeersUpdatedCB(trustedPeers []*peering.TrustedPeer) {
	trustedPubKeys := lo.Map(trustedPeers, func(tp *peering.TrustedPeer) *cryptolib.PublicKey { return tp.PubKey() })
	c.accessMgr.TrustedNodes(trustedPubKeys)
}

func (c *ChainRunner) chainServersUpdatedCB(servers []*cryptolib.PublicKey) {
	c.mutex.RLock()
	defer c.mutex.RUnlock()

	if c.chain == nil {
		return
	}

	c.chain.ServersUpdated(servers)
}

func (c *ChainRunner) chainAccessUpdatedCB(accessNodes []*cryptolib.PublicKey) {
	c.accessMgr.ChainAccessNodes(accessNodes)
}

// activateWithoutLocking activates a chain in the node.
func (c *ChainRunner) activateWithoutLocking() error { //nolint:funlen
	if c.ctx == nil {
		return errors.New("run chains first")
	}
	if c.ctx.Err() != nil {
		return errors.New("node is shutting down")
	}

	//
	// Check, maybe it is already running.
	if c.chain != nil {
		c.log.LogDebugf("Chain is already activated")
		return nil
	}

	//
	// Activate the chain in the persistent store, if it is not activated yet.
	chainRecord, err := c.chainRecordRegistryProvider.ChainRecord()
	if err != nil {
		return fmt.Errorf("chain record does not exist: %w", err)
	}
	if !chainRecord.Active {
		if _, err2 := c.chainRecordRegistryProvider.ActivateChainRecord(); err2 != nil {
			return fmt.Errorf("cannot activate chain: %w", err2)
		}
	}

	chainID := chainRecord.ChainID()

	chainKVStore, writeMutex, err := c.chainStateStoreProvider(chainID)
	if err != nil {
		return fmt.Errorf("error when creating chain KV store: %w", err)
	}

	chainMetrics := c.chainMetricsProvider.GetChainMetrics(chainID)

	// Initialize WAL
	chainLog := c.log.NewChildLogger(chainID.ShortString())
	var chainWAL utils.BlockWAL
	if c.walEnabled {
		chainWAL, err = utils.NewBlockWAL(chainLog, c.walFolderPath, chainMetrics.BlockWAL)
		if err != nil {
			panic(fmt.Errorf("cannot create WAL: %w", err))
		}
	} else {
		chainWAL = utils.NewEmptyBlockWAL()
	}

	stateManagerParameters := gpa.NewStateManagerParameters()
	stateManagerParameters.BlockCacheMaxSize = c.smBlockCacheMaxSize
	stateManagerParameters.BlockCacheBlocksInCacheDuration = c.smBlockCacheBlocksInCacheDuration
	stateManagerParameters.BlockCacheBlockCleaningPeriod = c.smBlockCacheBlockCleaningPeriod
	stateManagerParameters.StateManagerGetBlockNodeCount = c.smStateManagerGetBlockNodeCount
	stateManagerParameters.StateManagerGetBlockRetry = c.smStateManagerGetBlockRetry
	stateManagerParameters.StateManagerRequestCleaningPeriod = c.smStateManagerRequestCleaningPeriod
	stateManagerParameters.StateManagerStatusLogPeriod = c.smStateManagerStatusLogPeriod
	stateManagerParameters.StateManagerTimerTickPeriod = c.smStateManagerTimerTickPeriod
	stateManagerParameters.PruningMinStatesToKeep = c.smPruningMinStatesToKeep
	stateManagerParameters.PruningMaxStatesToDelete = c.smPruningMaxStatesToDelete

	// Initialize Snapshotter
	chainStore := indexedstore.New(state.NewStoreWithMetrics(chainKVStore, writeMutex, chainMetrics.State))
	chainCtx, chainCancel := context.WithCancel(c.ctx)
	validatorAgentID := accounts.CommonAccount()
	if c.validatorFeeAddr != nil {
		validatorAgentID = isc.NewAddressAgentID(c.validatorFeeAddr)
	}
	chainShutdownCoordinator := c.shutdownCoordinator.Nested(fmt.Sprintf("Chain-%s", chainID.AsAddress().String()))
	chainSnapshotManager, err := snapshots.NewSnapshotManager(
		chainCtx,
		chainShutdownCoordinator.Nested("SnapMgr"),
		chainID,
		c.snapshotToLoad,
		c.snapshotPeriod,
		c.snapshotDelay,
		c.snapshotFolderPath,
		c.snapshotNetworkPaths,
		chainStore,
		chainMetrics.Snapshots,
		chainLog,
	)
	if err != nil {
		panic(fmt.Errorf("cannot create Snapshotter: %w", err))
	}

	newChain, err := chain.New(
		chainCtx,
		chainLog,
		chainID,
		chainStore,
		c.nodeConnection,
		c.nodeIdentityProvider.NodeIdentity(),
		c.processorConfig,
		c.dkShareRegistryProvider,
		c.consensusStateRegistry,
		c.walLoadToStore,
		chainWAL,
		chainSnapshotManager,
		c.chainListener,
		chainRecord.AccessNodes,
		c.networkProvider,
		chainMetrics,
		chainShutdownCoordinator,
		func() { c.chainMetricsProvider.RegisterChain(chainID) },
		func() { c.chainMetricsProvider.UnregisterChain(chainID) },
		c.deriveAliasOutputByQuorum,
		c.pipeliningLimit,
		c.postponeRecoveryMilestones,
		c.consensusDelay,
		c.recoveryTimeout,
		validatorAgentID,
		stateManagerParameters,
		c.mempoolSettings,
		c.mempoolBroadcastInterval,
		0,
	)
	if err != nil {
		chainCancel()
		return fmt.Errorf("Chains.Activate: failed to create chain object: %w", err)
	}
	c.chain = newChain
	c.chainCancelFunc = chainCancel

	c.log.LogInfof("activated chain: %v = %s", chainID.ShortString(), chainID.String())
	return nil
}

// Activate activates a chain in the node.
func (c *ChainRunner) Activate() error {
	c.mutex.Lock()
	defer c.mutex.Unlock()

	return c.activateWithoutLocking()
}

// Deactivate a chain in the node.
func (c *ChainRunner) Deactivate() error {
	c.mutex.Lock()
	defer c.mutex.Unlock()

	if c.chain == nil {
		c.log.LogDebugf("chain is not active")
		return nil
	}

	if _, err := c.chainRecordRegistryProvider.DeactivateChainRecord(); err != nil {
		return fmt.Errorf("cannot deactivate chain %v: %w", err)
	}

	c.chainCancelFunc()
	c.accessMgr.ChainDismissed()

	c.log.LogDebugf("chain has been deactivated: %v = %s", c.chain.ID().ShortString(), c.chain.ID().String())
	c.chain = nil
	c.chainCancelFunc = nil

	return nil
}

// Get returns active chain object or nil if it doesn't exist
// lazy unsubscribing
func (c *ChainRunner) Get() (chain.Chain, error) {
	c.mutex.RLock()
	defer c.mutex.RUnlock()

	if c.chain == nil {
		return nil, interfaces.ErrChainNotFound
	}
	return c.chain, nil
}

func (c *ChainRunner) ValidatorAddress() *cryptolib.Address {
	return c.validatorFeeAddr
}

func (c *ChainRunner) IsArchiveNode() bool {
	return c.smPruningMinStatesToKeep < 1
}
