package chainrunner

import (
	"context"
	"fmt"
	"time"

	"go.uber.org/dig"

	"github.com/iotaledger/hive.go/app"
	hiveshutdown "github.com/iotaledger/hive.go/app/shutdown"

	"github.com/iotaledger/wasp/v2/packages/chain"
	"github.com/iotaledger/wasp/v2/packages/chain/cmtlog"
	"github.com/iotaledger/wasp/v2/packages/chain/mempool"
	chainrunner "github.com/iotaledger/wasp/v2/packages/chainrunner"
	"github.com/iotaledger/wasp/v2/packages/daemon"
	"github.com/iotaledger/wasp/v2/packages/database"
	"github.com/iotaledger/wasp/v2/packages/metrics"
	"github.com/iotaledger/wasp/v2/packages/peering"
	"github.com/iotaledger/wasp/v2/packages/publisher"
	"github.com/iotaledger/wasp/v2/packages/registry"
	"github.com/iotaledger/wasp/v2/packages/shutdown"
	"github.com/iotaledger/wasp/v2/packages/vm/processors"
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
	Chains          *chainrunner.ChainRunner
}

func initConfigParams(c *dig.Container) error {
	type cfgResult struct {
		dig.Out
		APICacheTTL time.Duration `name:"apiCacheTTL"`
	}

	if err := c.Provide(func() cfgResult {
		return cfgResult{
			APICacheTTL: ParamsChainRunner.APICacheTTL,
		}
	}); err != nil {
		Component.LogPanic(err.Error())
	}

	chain.RedeliveryPeriod = ParamsChainRunner.RedeliveryPeriod
	chain.PrintStatusPeriod = ParamsChainRunner.PrintStatusPeriod
	chain.ConsensusInstsInAdvance = ParamsChainRunner.ConsensusInstsInAdvance
	chain.AwaitReceiptCleanupEvery = ParamsChainRunner.AwaitReceiptCleanupEvery

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

	type chainRunnerResult struct {
		dig.Out

		Runner *chainrunner.ChainRunner
	}

	if err := c.Provide(func(deps chainsDeps) chainRunnerResult {
		return chainRunnerResult{
			Runner: chainrunner.New(
				Component.Logger,
				deps.NodeConnection,
				deps.ProcessorsConfig,
				ParamsValidator.Address,
				ParamsChainRunner.DeriveAliasOutputByQuorum,
				ParamsChainRunner.PipeliningLimit,
				ParamsChainRunner.PostponeRecoveryMilestones,
				ParamsChainRunner.ConsensusDelay,
				ParamsChainRunner.RecoveryTimeout,
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
				ParamsSnapshotManager.SnapshotToLoad,
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
					TTL:                        ParamsChainRunner.MempoolTTL,
					OnLedgerRefreshMinInterval: ParamsChainRunner.MempoolOnLedgerRefreshMinInterval,
					MaxOffledgerInPool:         ParamsChainRunner.MempoolMaxOffledgerInPool,
					MaxOnledgerInPool:          ParamsChainRunner.MempoolMaxOnledgerInPool,
					MaxTimedInPool:             ParamsChainRunner.MempoolMaxTimedInPool,
					MaxOnledgerToPropose:       ParamsChainRunner.MempoolMaxOnledgerToPropose,
					MaxOffledgerToPropose:      ParamsChainRunner.MempoolMaxOffledgerToPropose,
					MaxOffledgerPerAccount:     ParamsChainRunner.MempoolMaxOffledgerPerAccount,
				},
				ParamsChainRunner.BroadcastInterval,
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
