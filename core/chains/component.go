package chains

import (
	"context"
	"time"

	"github.com/labstack/gommon/log"
	"go.uber.org/dig"

	"github.com/iotaledger/hive.go/core/app"
	"github.com/iotaledger/wasp/core/database"
	"github.com/iotaledger/wasp/core/nodeconn"
	"github.com/iotaledger/wasp/core/peering"
	"github.com/iotaledger/wasp/core/processors"
	"github.com/iotaledger/wasp/core/registry"
	_ "github.com/iotaledger/wasp/packages/chain/chainimpl"
	"github.com/iotaledger/wasp/packages/chains"
	metricspkg "github.com/iotaledger/wasp/packages/metrics"
	"github.com/iotaledger/wasp/packages/parameters"
	"github.com/iotaledger/wasp/packages/util/ready"
	"github.com/iotaledger/wasp/packages/wal"
	"github.com/iotaledger/wasp/plugins/metrics"
)

func init() {
	CoreComponent = &app.CoreComponent{
		Component: &app.Component{
			Name:           "Chains",
			DepsFunc:       func(cDeps dependencies) { deps = cDeps },
			Params:         params,
			InitConfigPars: initConfigPars,
			Configure:      configure,
			Run:            run,
		},
	}
}

var (
	CoreComponent *app.CoreComponent
	deps          dependencies

	initialized *ready.Ready
	allChains   *chains.Chains
	allMetrics  *metricspkg.Metrics
)

type dependencies struct {
	dig.In

	MetricsEnabled bool `name:"metricsEnabled"`
	WAL            *wal.WAL
}

func initConfigPars(c *dig.Container) error {
	type cfgResult struct {
		dig.Out
		APICacheTTL time.Duration `name:"apiCacheTTL"`
	}

	if err := c.Provide(func() cfgResult {
		return cfgResult{
			APICacheTTL: ParamsChains.APICacheTTL,
		}
	}); err != nil {
		CoreComponent.LogPanic(err)
	}

	return nil
}

func configure() error {
	initialized = ready.New(CoreComponent.Name)
	return nil
}

func run() error {
	allChains = chains.New(
		CoreComponent.Logger(),
		processors.Config,
		ParamsChains.BroadcastUpToNPeers,
		ParamsChains.BroadcastInterval,
		ParamsChains.PullMissingRequestsFromCommittee,
		peering.DefaultNetworkProvider(),
		database.GetOrCreateKVStore,
		ParamsRawBlocks.Enabled,
		ParamsRawBlocks.Directory,
	)

	err := CoreComponent.Daemon().BackgroundWorker(CoreComponent.Name, func(ctx context.Context) {
		if deps.MetricsEnabled {
			allMetrics = metrics.AllMetrics()
		}

		allChains.SetNodeConn(nodeconn.NodeConnection())
		if err := allChains.ActivateAllFromRegistry(registry.DefaultRegistry, allMetrics, deps.WAL); err != nil {
			log.Errorf("failed to read chain activation records from registry: %v", err)
			return
		}

		initialized.SetReady()

		<-ctx.Done()

		log.Info("dismissing chains...")
		go func() {
			allChains.Dismiss()
			log.Info("dismissing chains... Done")
		}()
	}, parameters.PriorityChains)
	if err != nil {
		log.Error(err)
		return err
	}

	return nil
}

func AllChains() *chains.Chains {
	initialized.MustWait(5 * time.Second)
	return allChains
}
