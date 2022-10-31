package app

import (
	_ "net/http/pprof"

	"github.com/iotaledger/hive.go/core/app"
	"github.com/iotaledger/hive.go/core/app/core/shutdown"
	"github.com/iotaledger/hive.go/core/app/plugins/profiling"
	"github.com/iotaledger/inx-app/inx"

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
	"github.com/iotaledger/wasp/packages/wasp"
	"github.com/iotaledger/wasp/plugins/dashboard"
	"github.com/iotaledger/wasp/plugins/metrics"
	"github.com/iotaledger/wasp/plugins/publishernano"
	"github.com/iotaledger/wasp/plugins/wal"
	"github.com/iotaledger/wasp/plugins/webapi"
)

func App() *app.App {
	return app.New(wasp.Name, wasp.Version,
		app.WithVersionCheck("iotaledger", "wasp"),
		app.WithInitComponent(InitComponent),
		app.WithCoreComponents([]*app.CoreComponent{
			inx.CoreComponent,
			shutdown.CoreComponent,
			nodeconn.CoreComponent,
			users.CoreComponent,
			logger.CoreComponent,
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
		AdditionalConfigs: []*app.ConfigurationSet{
			app.NewConfigurationSet("users", "users", "usersConfigFilePath", "usersConfig", false, true, false, "users.json", "u"),
		},
	}
}
