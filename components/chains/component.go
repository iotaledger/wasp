package chains

import (
	"context"
	"fmt"
	"time"

	"go.uber.org/dig"

	"github.com/iotaledger/hive.go/app"
	hiveshutdown "github.com/iotaledger/hive.go/app/shutdown"

	"github.com/iotaledger/wasp/packages/chain"
	"github.com/iotaledger/wasp/packages/chain/cmtlog"
	"github.com/iotaledger/wasp/packages/chain/mempool"
	"github.com/iotaledger/wasp/packages/chains"
	"github.com/iotaledger/wasp/packages/daemon"
	"github.com/iotaledger/wasp/packages/database"
	"github.com/iotaledger/wasp/packages/metrics"
	"github.com/iotaledger/wasp/packages/peering"
	"github.com/iotaledger/wasp/packages/publisher"
	"github.com/iotaledger/wasp/packages/registry"
	"github.com/iotaledger/wasp/packages/shutdown"
	"github.com/iotaledger/wasp/packages/vm/processors"
)

func init() {
	Component = &app.Component{
		Name:             "Chains",
		DepsFunc:         func(cDeps dependencies) { deps = cDeps },
		Params:           params,
		InitConfigParams: initConfigParams,
		Provide:          provide,
		Run:              run,
	}
}

var (
	Component *app.Component
	deps      dependencies
)

type dependencies struct {
	dig.In

	ShutdownHandler *hiveshutdown.ShutdownHandler
	Chains          *chains.Chains
}

func initConfigParams(c *dig.Container) error {
	type cfgResult struct {
		dig.Out
		APICacheTTL time.Duration `name:"apiCacheTTL"`
	}

	if err := c.Provide(func() cfgResult {
		return cfgResult{
			APICacheTTL: ParamsChains.APICacheTTL,
		}
	}); err != nil {
		Component.LogPanic(err.Error())
	}

	chain.RedeliveryPeriod = ParamsChains.RedeliveryPeriod
	chain.PrintStatusPeriod = ParamsChains.PrintStatusPeriod
	chain.ConsensusInstsInAdvance = ParamsChains.ConsensusInstsInAdvance
	chain.AwaitReceiptCleanupEvery = ParamsChains.AwaitReceiptCleanupEvery

	return nil
}

func provide(c *dig.Container) error {
	type chainsDeps struct {
		dig.In

		NodeConnection              chain.NodeConnection
		ProcessorsConfig            *processors.Config
		NetworkProvider             peering.NetworkProvider       `name:"networkProvider"`
		TrustedNetworkManager       peering.TrustedNetworkManager `name:"trustedNetworkManager"`
		ChainStateDatabaseManager   *database.ChainStateDatabaseManager
		ChainRecordRegistryProvider registry.ChainRecordRegistryProvider
		DKShareRegistryProvider     registry.DKShareRegistryProvider
		NodeIdentityProvider        registry.NodeIdentityProvider
		ConsensusStateRegistry      cmtlog.ConsensusStateRegistry
		ChainListener               *publisher.Publisher
		ChainMetricsProvider        *metrics.ChainMetricsProvider
	}

	type chainsResult struct {
		dig.Out

		Chains *chains.Chains
	}

	if err := c.Provide(func(deps chainsDeps) chainsResult {
		return chainsResult{
			Chains: chains.New(
				Component.Logger,
				deps.NodeConnection,
				deps.ProcessorsConfig,
				ParamsValidator.Address,
				ParamsChains.DeriveAliasOutputByQuorum,
				ParamsChains.PipeliningLimit,
				ParamsChains.PostponeRecoveryMilestones,
				ParamsChains.ConsensusDelay,
				ParamsChains.RecoveryTimeout,
				deps.NetworkProvider,
				deps.TrustedNetworkManager,
				deps.ChainStateDatabaseManager.ChainStateKVStore,
				ParamsWAL.LoadToStore,
				ParamsWAL.Enabled,
				ParamsWAL.Path,
				ParamsStateManager.BlockCacheMaxSize,
				ParamsStateManager.BlockCacheBlocksInCacheDuration,
				ParamsStateManager.BlockCacheBlockCleaningPeriod,
				ParamsStateManager.StateManagerGetBlockNodeCount,
				ParamsStateManager.StateManagerGetBlockRetry,
				ParamsStateManager.StateManagerRequestCleaningPeriod,
				ParamsStateManager.StateManagerStatusLogPeriod,
				ParamsStateManager.StateManagerTimerTickPeriod,
				ParamsStateManager.PruningMinStatesToKeep,
				ParamsStateManager.PruningMaxStatesToDelete,
				ParamsSnapshotManager.SnapshotsToLoad,
				ParamsSnapshotManager.Period,
				ParamsSnapshotManager.Delay,
				ParamsSnapshotManager.LocalPath,
				ParamsSnapshotManager.NetworkPaths,
				deps.ChainRecordRegistryProvider,
				deps.DKShareRegistryProvider,
				deps.NodeIdentityProvider,
				deps.ConsensusStateRegistry,
				deps.ChainListener,
				mempool.Settings{
					TTL:                        ParamsChains.MempoolTTL,
					OnLedgerRefreshMinInterval: ParamsChains.MempoolOnLedgerRefreshMinInterval,
					MaxOffledgerInPool:         ParamsChains.MempoolMaxOffledgerInPool,
					MaxOnledgerInPool:          ParamsChains.MempoolMaxOnledgerInPool,
					MaxTimedInPool:             ParamsChains.MempoolMaxTimedInPool,
					MaxOnledgerToPropose:       ParamsChains.MempoolMaxOnledgerToPropose,
					MaxOffledgerToPropose:      ParamsChains.MempoolMaxOffledgerToPropose,
					MaxOffledgerPerAccount:     ParamsChains.MempoolMaxOffledgerPerAccount,
				},
				ParamsChains.BroadcastInterval,
				shutdown.NewCoordinator("chains", Component.NewChildLogger("Shutdown")),
				deps.ChainMetricsProvider,
			),
		}
	}); err != nil {
		Component.LogPanic(err.Error())
	}

	return nil
}

func run() error {
	err := Component.Daemon().BackgroundWorker(Component.Name, func(ctx context.Context) {
		if err := deps.Chains.Run(ctx); err != nil {
			deps.ShutdownHandler.SelfShutdown(fmt.Sprintf("Starting %s failed, error: %s", Component.Name, err.Error()), true)
			return
		}

		<-ctx.Done()
		deps.Chains.Close()
	}, daemon.PriorityChains)
	if err != nil {
		Component.LogError(err.Error())
		return err
	}

	return nil
}
