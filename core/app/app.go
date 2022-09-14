package app

import (
	_ "net/http/pprof"

	"github.com/iotaledger/hive.go/core/app"
	"github.com/iotaledger/hive.go/core/app/core/shutdown"
	"github.com/iotaledger/hive.go/core/app/plugins/profiling"

	"github.com/iotaledger/wasp/core/chains"
	"github.com/iotaledger/wasp/core/database"
	"github.com/iotaledger/wasp/core/dkg"
	"github.com/iotaledger/wasp/core/logger"
	"github.com/iotaledger/wasp/core/nodeconn"
	"github.com/iotaledger/wasp/core/peering"
	"github.com/iotaledger/wasp/core/processors"
	"github.com/iotaledger/wasp/core/registry"
	"github.com/iotaledger/wasp/core/users"
	"github.com/iotaledger/wasp/core/wasmtimevm"
	"github.com/iotaledger/wasp/plugins/dashboard"
	"github.com/iotaledger/wasp/plugins/metrics"
	"github.com/iotaledger/wasp/plugins/publishernano"
	"github.com/iotaledger/wasp/plugins/wal"
	"github.com/iotaledger/wasp/plugins/webapi"
)

var (
	// Name of the app.
	Name = "WASP"

	// Version of the app.
	Version = "0.3.0-alpha.1"
)

func App() *app.App {
	return app.New(Name, Version,
		app.WithVersionCheck("iotaledger", "wasp"),
		app.WithInitComponent(InitComponent),
		app.WithCoreComponents([]*app.CoreComponent{
			shutdown.CoreComponent,
			users.CoreComponent,
			logger.CoreComponent,
			nodeconn.CoreComponent,
			database.CoreComponent,
			registry.CoreComponent,
			peering.CoreComponent,
			dkg.CoreComponent,
			processors.CoreComponent,
			wasmtimevm.CoreComponent,
			chains.CoreComponent,
		}...),
		app.WithPlugins([]*app.Plugin{
			profiling.Plugin,
			wal.Plugin,
			metrics.Plugin,
			webapi.Plugin,
			publishernano.Plugin,
			dashboard.Plugin,
		}...),
	)
}

var InitComponent *app.InitComponent

func init() {
	InitComponent = &app.InitComponent{
		Component: &app.Component{
			Name: "App",
		},
		NonHiddenFlags: []string{
			"app.checkForUpdates",
			"app.profile",
			"config",
			"help",
			"peering",
			"version",
		},
	}
}
