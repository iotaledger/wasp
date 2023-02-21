package app

import (
	_ "net/http/pprof"

	"github.com/iotaledger/hive.go/core/app"
	"github.com/iotaledger/hive.go/core/app/core/shutdown"
	"github.com/iotaledger/hive.go/core/app/plugins/profiling"
	"github.com/iotaledger/inx-app/core/inx"
	"github.com/iotaledger/wasp/core/chains"
	"github.com/iotaledger/wasp/core/database"
	"github.com/iotaledger/wasp/core/dkg"
	"github.com/iotaledger/wasp/core/logger"
	"github.com/iotaledger/wasp/core/nodeconn"
	"github.com/iotaledger/wasp/core/peering"
	"github.com/iotaledger/wasp/core/processors"
	"github.com/iotaledger/wasp/core/publisher"
	"github.com/iotaledger/wasp/core/registry"
	"github.com/iotaledger/wasp/core/users"
	"github.com/iotaledger/wasp/core/wasmtimevm"
	"github.com/iotaledger/wasp/plugins/profilingrecorder"
	"github.com/iotaledger/wasp/plugins/prometheus"
	"github.com/iotaledger/wasp/plugins/webapi"
)

var (
	// Name of the app.
	Name = "Wasp"

	// Version of the app.
	// This field is populated by the scripts that compile wasp.
	Version = ""
)

func App() *app.App {
	return app.New(Name, Version,
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
			publisher.CoreComponent,
		}...),
		app.WithPlugins([]*app.Plugin{
			profiling.Plugin,
			profilingrecorder.Plugin,
			prometheus.Plugin,
			webapi.Plugin,
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
