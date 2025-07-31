// Package app provides core application functionality and lifecycle management.
package app

import (
	_ "net/http/pprof"

	"github.com/iotaledger/hive.go/app"
	"github.com/iotaledger/hive.go/app/components/profiling"
	"github.com/iotaledger/hive.go/app/components/shutdown"
	"github.com/iotaledger/wasp/v2/components/cache"
	"github.com/iotaledger/wasp/v2/components/chains"
	"github.com/iotaledger/wasp/v2/components/database"
	"github.com/iotaledger/wasp/v2/components/dkg"
	"github.com/iotaledger/wasp/v2/components/logger"
	"github.com/iotaledger/wasp/v2/components/nodeconn"
	"github.com/iotaledger/wasp/v2/components/peering"
	"github.com/iotaledger/wasp/v2/components/processors"
	"github.com/iotaledger/wasp/v2/components/profilingrecorder"
	"github.com/iotaledger/wasp/v2/components/prometheus"
	"github.com/iotaledger/wasp/v2/components/publisher"
	"github.com/iotaledger/wasp/v2/components/registry"
	"github.com/iotaledger/wasp/v2/components/users"
	"github.com/iotaledger/wasp/v2/components/webapi"
	"github.com/iotaledger/wasp/v2/packages/toolset"
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
		app.WithComponents(
			shutdown.Component,
			nodeconn.Component,
			users.Component,
			logger.Component,
			cache.Component,
			database.Component,
			registry.Component,
			peering.Component,
			dkg.Component,
			processors.Component,
			chains.Component,
			publisher.Component,
			webapi.Component,
			profiling.Component,
			profilingrecorder.Component,
			prometheus.Component,
		),
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
		Init: initialize,
	}
}

func initialize(_ *app.App) error {
	if toolset.ShouldHandleTools() {
		toolset.HandleTools()
		// HandleTools will call os.Exit
	}

	return nil
}
