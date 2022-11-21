// Copyright 2020 IOTA Stiftung
// SPDX-License-Identifier: Apache-2.0

package dashboard

import (
	"context"
	"errors"
	"net/http"
	"time"

	"github.com/labstack/echo/v4"
	"go.uber.org/dig"

	"github.com/iotaledger/hive.go/core/app"
	"github.com/iotaledger/inx-app/pkg/httpserver"
	"github.com/iotaledger/wasp/packages/authentication"
	"github.com/iotaledger/wasp/packages/authentication/shared/permissions"
	"github.com/iotaledger/wasp/packages/chains"
	"github.com/iotaledger/wasp/packages/dashboard"
	"github.com/iotaledger/wasp/packages/parameters"
	"github.com/iotaledger/wasp/packages/peering"
	"github.com/iotaledger/wasp/packages/registry"
	"github.com/iotaledger/wasp/packages/users"
)

func init() {
	Plugin = &app.Plugin{
		Component: &app.Component{
			Name:     "Dashboard",
			DepsFunc: func(cDeps dependencies) { deps = cDeps },
			Params:   params,
			Provide:  provide,
			Run:      run,
		},
		IsEnabled: func() bool {
			return ParamsDashboard.Enabled
		},
	}
}

var (
	Plugin *app.Plugin
	deps   dependencies
)

type dependencies struct {
	dig.In

	Echo *echo.Echo `name:"dashboardServer"`
}

func provide(c *dig.Container) error {
	type dashboardDeps struct {
		dig.In

		WebAPIBindAddress            string `name:"webAPIBindAddress"`
		Chains                       *chains.Chains
		DefaultRegistry              registry.Registry
		DefaultNetworkProvider       peering.NetworkProvider       `name:"defaultNetworkProvider"`
		DefaultTrustedNetworkManager peering.TrustedNetworkManager `name:"defaultTrustedNetworkManager"`
		UserManager                  *users.UserManager
	}

	type dashboardResult struct {
		dig.Out

		Echo      *echo.Echo `name:"dashboardServer"`
		Dashboard *dashboard.Dashboard
	}

	if err := c.Provide(func(deps dashboardDeps) dashboardResult {
		e := httpserver.NewEcho(
			Plugin.Logger(),
			nil,
			ParamsDashboard.DebugRequestLoggerEnabled,
		)
		e.HidePort = true

		claimValidator := func(claims *authentication.WaspClaims) bool {
			// The Dashboard will be accessible if the token has a 'Dashboard' claim
			return claims.HasPermission(permissions.Dashboard)
		}

		authentication.AddAuthentication(
			e,
			deps.UserManager,
			func() registry.Registry {
				return deps.DefaultRegistry
			},
			ParamsDashboard.Auth,
			claimValidator,
		)

		waspServices := dashboard.NewWaspServices(
			deps.WebAPIBindAddress,
			ParamsDashboard.ExploreAddressURL,
			Plugin.App().Config(),
			deps.Chains,
			deps.DefaultRegistry,
			deps.DefaultNetworkProvider,
			deps.DefaultTrustedNetworkManager,
		)

		return dashboardResult{
			Echo: e,
			Dashboard: dashboard.New(
				Plugin.Logger(),
				e,
				waspServices,
			),
		}
	}); err != nil {
		Plugin.LogPanic(err)
	}

	return nil
}

func run() error {
	Plugin.LogInfof("Starting %s server ...", Plugin.Name)
	if err := Plugin.Daemon().BackgroundWorker(Plugin.Name, func(ctx context.Context) {
		Plugin.LogInfof("Starting %s server ... done", Plugin.Name)

		go func() {
			Plugin.LogInfof("You can now access the dashboard using: http://%s", ParamsDashboard.BindAddress)
			if err := deps.Echo.Start(ParamsDashboard.BindAddress); err != nil && !errors.Is(err, http.ErrServerClosed) {
				Plugin.LogWarnf("Stopped %s server due to an error (%s)", Plugin.Name, err)
			}
		}()

		<-ctx.Done()
		Plugin.LogInfof("Stopping %s server ...", Plugin.Name)

		shutdownCtx, shutdownCtxCancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer shutdownCtxCancel()

		//nolint:contextcheck // false positive
		if err := deps.Echo.Shutdown(shutdownCtx); err != nil {
			Plugin.LogWarn(err)
		}

		Plugin.LogInfof("Stopping %s server ... done", Plugin.Name)
	}, parameters.PriorityDashboard); err != nil {
		Plugin.LogPanicf("failed to start worker: %s", err)
	}

	return nil
}
