package main

import (
	"github.com/iotaledger/hive.go/node"
	"github.com/iotaledger/wasp/packages/parameters"
	"github.com/iotaledger/wasp/plugins/banner"
	"github.com/iotaledger/wasp/plugins/chains"
	"github.com/iotaledger/wasp/plugins/config"
	"github.com/iotaledger/wasp/plugins/dashboard"
	"github.com/iotaledger/wasp/plugins/database"
	"github.com/iotaledger/wasp/plugins/dkg"
	"github.com/iotaledger/wasp/plugins/gracefulshutdown"
	"github.com/iotaledger/wasp/plugins/logger"
	"github.com/iotaledger/wasp/plugins/metrics"
	"github.com/iotaledger/wasp/plugins/nodeconn"
	"github.com/iotaledger/wasp/plugins/peering"
	"github.com/iotaledger/wasp/plugins/processors"
	"github.com/iotaledger/wasp/plugins/profiling"
	"github.com/iotaledger/wasp/plugins/publishernano"
	"github.com/iotaledger/wasp/plugins/registry"
	"github.com/iotaledger/wasp/plugins/users"
	"github.com/iotaledger/wasp/plugins/wal"
	"github.com/iotaledger/wasp/plugins/wasmtimevm"
	"github.com/iotaledger/wasp/plugins/webapi"
)

func main() {
	params := parameters.Init()

	plugins := node.Plugins(
		users.Init(params),
		banner.Init(),
		config.Init(params),
		logger.Init(params),
		gracefulshutdown.Init(),
		nodeconn.Init(),
		database.Init(),
		registry.Init(),
		peering.Init(),
		dkg.Init(),
		processors.Init(),
		wasmtimevm.Init(),
		wal.Init(),
		chains.Init(),
		metrics.Init(),
		webapi.Init(),
		publishernano.Init(),
		dashboard.Init(),
		profiling.Init(),
	)

	node.Run(
		plugins,
	)
}
