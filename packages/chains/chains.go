// Copyright 2020 IOTA Stiftung
// SPDX-License-Identifier: Apache-2.0

package chains

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"sync"
	"time"

	"github.com/iotaledger/hive.go/ds/shrinkingmap"
	"github.com/iotaledger/hive.go/lo"
	"github.com/iotaledger/hive.go/log"

	"github.com/iotaledger/wasp/packages/chain"
	"github.com/iotaledger/wasp/packages/chain/cmt_log"
	"github.com/iotaledger/wasp/packages/chain/mempool"
	"github.com/iotaledger/wasp/packages/chain/statemanager/gpa"
	"github.com/iotaledger/wasp/packages/chain/statemanager/gpa/sm_gpa_utils"
	"github.com/iotaledger/wasp/packages/chain/statemanager/snapshots"
	"github.com/iotaledger/wasp/packages/chains/accessmanager"
	"github.com/iotaledger/wasp/packages/cryptolib"
	"github.com/iotaledger/wasp/packages/database"
	"github.com/iotaledger/wasp/packages/isc"
	"github.com/iotaledger/wasp/packages/metrics"
	"github.com/iotaledger/wasp/packages/peering"
	"github.com/iotaledger/wasp/packages/registry"
	"github.com/iotaledger/wasp/packages/shutdown"
	"github.com/iotaledger/wasp/packages/state"
	"github.com/iotaledger/wasp/packages/state/indexedstore"
	"github.com/iotaledger/wasp/packages/util"
	"github.com/iotaledger/wasp/packages/vm/core/accounts"
	"github.com/iotaledger/wasp/packages/vm/processors"
	"github.com/iotaledger/wasp/packages/webapi/interfaces"
)

type Provider func() *Chains // TODO: Use DI instead of that.

type ChainProvider func(chainID isc.ChainID) chain.Chain

type Chains struct {
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
	defaultSnapshotToLoad               *state.BlockHash
	snapshotsToLoad                     map[isc.ChainIDKey]state.BlockHash
	snapshotPeriod                      uint32
	snapshotDelay                       uint32
	snapshotFolderPath                  string
	snapshotNetworkPaths                []string

	chainRecordRegistryProvider registry.ChainRecordRegistryProvider
	dkShareRegistryProvider     registry.DKShareRegistryProvider
	nodeIdentityProvider        registry.NodeIdentityProvider
	consensusStateRegistry      cmt_log.ConsensusStateRegistry
	chainListener               chain.ChainListener

	mutex     *sync.RWMutex
	allChains *shrinkingmap.ShrinkingMap[isc.ChainID, *activeChain]
	accessMgr accessmanager.AccessMgr

	cleanupFunc         context.CancelFunc
	shutdownCoordinator *shutdown.Coordinator

	chainMetricsProvider *metrics.ChainMetricsProvider

	validatorFeeAddr *cryptolib.Address

	mempoolSettings          mempool.Settings
	mempoolBroadcastInterval time.Duration
}

type activeChain struct {
	chain      chain.Chain
	cancelFunc context.CancelFunc
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
	snapshotsToLoad []string,
	snapshotPeriod uint32,
	snapshotDelay uint32,
	snapshotFolderPath string,
	snapshotNetworkPaths []string,
	chainRecordRegistryProvider registry.ChainRecordRegistryProvider,
	dkShareRegistryProvider registry.DKShareRegistryProvider,
	nodeIdentityProvider registry.NodeIdentityProvider,
	consensusStateRegistry cmt_log.ConsensusStateRegistry,
	chainListener chain.ChainListener,
	mempoolSettings mempool.Settings,
	mempoolBroadcastInterval time.Duration,
	shutdownCoordinator *shutdown.Coordinator,
	chainMetricsProvider *metrics.ChainMetricsProvider,
) *Chains {
	var validatorFeeAddr *cryptolib.Address
	if validatorAddrStr != "" {
		addr, err := cryptolib.NewAddressFromHexString(validatorAddrStr)
		if err != nil {
			panic(fmt.Errorf("error parsing validator.address: %s", err.Error()))
		}
		validatorFeeAddr = addr
	}
	ret := &Chains{
		log:                                 log,
		mutex:                               &sync.RWMutex{},
		allChains:                           shrinkingmap.New[isc.ChainID, *activeChain](),
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
	ret.initSnapshotsToLoad(snapshotsToLoad)
	ret.chainListener = NewChainsListener(chainListener, ret.chainAccessUpdatedCB)
	return ret
}

func (c *Chains) initSnapshotsToLoad(configs []string) {
	c.defaultSnapshotToLoad = nil
	c.snapshotsToLoad = make(map[isc.ChainIDKey]state.BlockHash)
	for _, config := range configs {
		configSplit := strings.Split(config, ":")
		// NOTE: Split does not return 0 length slice if second parameter is not zero length string; this is not checked
		if len(configSplit) == 1 {
			blockHash, err := state.BlockHashFromString(configSplit[0])
			if err != nil {
				c.log.LogWarnf("Parsing snapshots to load: %s is not a block hash: %v", configSplit[0], err)
				continue
			}
			c.defaultSnapshotToLoad = &blockHash
		} else {
			chainID, err := isc.ChainIDFromString(configSplit[0])
			if err != nil {
				c.log.LogWarnf("Parsing snapshots to load: %s in %s is not a chain ID: %v", configSplit[0], config, err)
				continue
			}
			blockHash, err := state.BlockHashFromString(configSplit[1])
			if err != nil {
				c.log.LogWarnf("Parsing snapshots to load: %s in %s is not a block hash: %v", configSplit[1], config, err)
				continue
			}
			c.snapshotsToLoad[chainID.Key()] = blockHash
		}
	}
}

func (c *Chains) Run(ctx context.Context) error {
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
		if chain, exists := c.allChains.Get(event.ChainRecord.ChainID()); exists {
			chain.chain.ConfigUpdated(event.ChainRecord.AccessNodes)
		}
	}).Unhook
	c.cleanupFunc = unhook

	return c.activateAllFromRegistry() //nolint:contextcheck
}

func (c *Chains) Close() {
	util.ExecuteIfNotNil(c.cleanupFunc)
	c.mutex.RLock()
	defer c.mutex.RUnlock()
	c.allChains.ForEach(func(_ isc.ChainID, ac *activeChain) bool {
		ac.cancelFunc()
		return true
	})
	c.shutdownCoordinator.WaitNestedWithLogging(1 * time.Second)
	c.shutdownCoordinator.Done()
	util.ExecuteIfNotNil(c.trustedNetworkListenerCancel)
	c.trustedNetworkListenerCancel = nil
}

func (c *Chains) trustedPeersUpdatedCB(trustedPeers []*peering.TrustedPeer) {
	trustedPubKeys := lo.Map(trustedPeers, func(tp *peering.TrustedPeer) *cryptolib.PublicKey { return tp.PubKey() })
	c.accessMgr.TrustedNodes(trustedPubKeys)
}

func (c *Chains) chainServersUpdatedCB(chainID isc.ChainID, servers []*cryptolib.PublicKey) {
	c.mutex.RLock()
	defer c.mutex.RUnlock()
	ch, exists := c.allChains.Get(chainID)
	if !exists {
		return
	}
	ch.chain.ServersUpdated(servers)
}

func (c *Chains) chainAccessUpdatedCB(chainID isc.ChainID, accessNodes []*cryptolib.PublicKey) {
	c.accessMgr.ChainAccessNodes(chainID, accessNodes)
}

func (c *Chains) activateAllFromRegistry() error {
	var innerErr error
	if err := c.chainRecordRegistryProvider.ForEachActiveChainRecord(func(chainRecord *registry.ChainRecord) bool {
		chainID := chainRecord.ChainID()
		if err := c.activateWithoutLocking(chainID); err != nil {
			innerErr = fmt.Errorf("cannot activate chain %s: %w", chainRecord.ChainID(), err)
			return false
		}

		return true
	}); err != nil {
		return err
	}

	return innerErr
}

// activateWithoutLocking activates a chain in the node.
func (c *Chains) activateWithoutLocking(chainID isc.ChainID) error { //nolint:funlen
	if c.ctx == nil {
		return errors.New("run chains first")
	}
	if c.ctx.Err() != nil {
		return errors.New("node is shutting down")
	}

	//
	// Check, maybe it is already running.
	if c.allChains.Has(chainID) {
		c.log.LogDebugf("Chain %v = %v is already activated", chainID.ShortString(), chainID.String())
		return nil
	}
	//
	// Activate the chain in the persistent store, if it is not activated yet.
	chainRecord, err := c.chainRecordRegistryProvider.ChainRecord(chainID)
	if err != nil {
		return fmt.Errorf("cannot get chain record for %v: %w", chainID, err)
	}
	if !chainRecord.Active {
		if _, err2 := c.chainRecordRegistryProvider.ActivateChainRecord(chainID); err2 != nil {
			return fmt.Errorf("cannot activate chain: %w", err2)
		}
	}

	chainKVStore, writeMutex, err := c.chainStateStoreProvider(chainID)
	if err != nil {
		return fmt.Errorf("error when creating chain KV store: %w", err)
	}

	chainMetrics := c.chainMetricsProvider.GetChainMetrics(chainID)

	// Initialize WAL
	chainLog := c.log.NewChildLogger(chainID.ShortString())
	var chainWAL sm_gpa_utils.BlockWAL
	if c.walEnabled {
		chainWAL, err = sm_gpa_utils.NewBlockWAL(chainLog, c.walFolderPath, chainID, chainMetrics.BlockWAL)
		if err != nil {
			panic(fmt.Errorf("cannot create WAL: %w", err))
		}
	} else {
		chainWAL = sm_gpa_utils.NewEmptyBlockWAL()
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
	blockHash, ok := c.snapshotsToLoad[chainID.Key()]
	var snapshotToLoad *state.BlockHash
	if ok {
		snapshotToLoad = &blockHash
	} else {
		snapshotToLoad = c.defaultSnapshotToLoad
	}
	chainSnapshotManager, err := snapshots.NewSnapshotManager(
		chainCtx,
		chainShutdownCoordinator.Nested("SnapMgr"),
		chainID,
		snapshotToLoad,
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
	c.allChains.Set(chainID, &activeChain{
		chain:      newChain,
		cancelFunc: chainCancel,
	})

	c.log.LogInfof("activated chain: %v = %s", chainID.ShortString(), chainID.String())
	return nil
}

// Activate activates a chain in the node.
func (c *Chains) Activate(chainID isc.ChainID) error {
	c.mutex.Lock()
	defer c.mutex.Unlock()

	return c.activateWithoutLocking(chainID)
}

// Deactivate a chain in the node.
func (c *Chains) Deactivate(chainID isc.ChainID) error {
	c.mutex.Lock()
	defer c.mutex.Unlock()

	if _, err := c.chainRecordRegistryProvider.DeactivateChainRecord(chainID); err != nil {
		return fmt.Errorf("cannot deactivate chain %v: %w", chainID, err)
	}

	ch, exists := c.allChains.Get(chainID)
	if !exists {
		c.log.LogDebugf("chain is not active: %v = %s", chainID.ShortString(), chainID.String())
		return nil
	}
	ch.cancelFunc()
	c.accessMgr.ChainDismissed(chainID)
	c.allChains.Delete(chainID)
	c.log.LogDebugf("chain has been deactivated: %v = %s", chainID.ShortString(), chainID.String())
	return nil
}

// Get returns active chain object or nil if it doesn't exist
// lazy unsubscribing
func (c *Chains) Get(chainID isc.ChainID) (chain.Chain, error) {
	c.mutex.RLock()
	defer c.mutex.RUnlock()

	ret, exists := c.allChains.Get(chainID)
	if !exists {
		return nil, interfaces.ErrChainNotFound
	}
	return ret.chain, nil
}

func (c *Chains) GetFirst() (chain.Chain, error) {
	c.mutex.RLock()
	defer c.mutex.RUnlock()

	if c.allChains.Size() == 0 {
		return nil, interfaces.ErrChainNotFound
	}

	ret := c.allChains.Values()[0]
	return ret.chain, nil
}

func (c *Chains) ValidatorAddress() *cryptolib.Address {
	return c.validatorFeeAddr
}

func (c *Chains) IsArchiveNode() bool {
	return c.smPruningMinStatesToKeep < 1
}
