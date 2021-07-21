// Copyright 2020 IOTA Stiftung
// SPDX-License-Identifier: Apache-2.0

package dashboard

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"strings"
	"time"

	"github.com/iotaledger/hive.go/daemon"
	"github.com/iotaledger/hive.go/logger"
	"github.com/iotaledger/hive.go/node"
	"github.com/iotaledger/wasp/packages/chain"
	"github.com/iotaledger/wasp/packages/dashboard"
	"github.com/iotaledger/wasp/packages/iscp"
	"github.com/iotaledger/wasp/packages/kv/dict"
	"github.com/iotaledger/wasp/packages/kv/optimism"
	"github.com/iotaledger/wasp/packages/parameters"
	peering_pkg "github.com/iotaledger/wasp/packages/peering"
	registry_pkg "github.com/iotaledger/wasp/packages/registry"
	"github.com/iotaledger/wasp/packages/state"
	"github.com/iotaledger/wasp/packages/util/auth"
	"github.com/iotaledger/wasp/packages/vm/viewcontext"
	"github.com/iotaledger/wasp/plugins/chains"
	"github.com/iotaledger/wasp/plugins/database"
	"github.com/iotaledger/wasp/plugins/peering"
	"github.com/iotaledger/wasp/plugins/registry"
	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"
)

const PluginName = "Dashboard"

var (
	Server = echo.New()

	log *logger.Logger

	d *dashboard.Dashboard
)

func Init() *node.Plugin {
	return node.NewPlugin(PluginName, node.Enabled, configure, run)
}

type waspServices struct{}

func (w *waspServices) ConfigDump() map[string]interface{} {
	return parameters.Dump()
}

func (w *waspServices) ExploreAddressBaseURL() string {
	baseURL := parameters.GetString(parameters.DashboardExploreAddressURL)
	if baseURL != "" {
		return baseURL
	}
	return exploreAddressURLFromGoshimmerURI(parameters.GetString(parameters.NodeAddress))
}

func (w *waspServices) GetChainRecords() ([]*registry_pkg.ChainRecord, error) {
	return registry.DefaultRegistry().GetChainRecords()
}

func (w *waspServices) GetChainRecord(chainID *iscp.ChainID) (*registry_pkg.ChainRecord, error) {
	return registry.DefaultRegistry().GetChainRecordByChainID(chainID)
}

func (w *waspServices) GetChainState(chainID *iscp.ChainID) (*dashboard.ChainState, error) {
	chainStore := database.GetKVStore(chainID)
	virtualState, _, err := state.LoadSolidState(chainStore, chainID)
	if err != nil {
		return nil, err
	}
	block, err := state.LoadBlock(chainStore, virtualState.BlockIndex())
	if err != nil {
		return nil, err
	}
	return &dashboard.ChainState{
		Index:             virtualState.BlockIndex(),
		Hash:              virtualState.Hash(),
		Timestamp:         virtualState.Timestamp().UnixNano(),
		ApprovingOutputID: block.ApprovingOutputID(),
	}, nil
}

func (w *waspServices) GetChain(chainID *iscp.ChainID) chain.ChainCore {
	return chains.AllChains().Get(chainID)
}

const retryOnStateInvalidatedRetry = 100 * time.Millisecond //nolint:gofumpt
const retryOnStateInvalidatedTimeout = 5 * time.Minute

func (w *waspServices) CallView(ch chain.ChainCore, hname iscp.Hname, funName string, params dict.Dict) (dict.Dict, error) {
	vctx := viewcontext.NewFromChain(ch)
	var ret dict.Dict
	err := optimism.RetryOnStateInvalidated(func() error {
		var err error
		ret, err = vctx.CallView(hname, iscp.Hn(funName), params)
		return err
	}, retryOnStateInvalidatedRetry, time.Now().Add(retryOnStateInvalidatedTimeout))
	if err != nil {
		return nil, fmt.Errorf("root view call failed: %w", err)
	}
	return ret, nil
}

func exploreAddressURLFromGoshimmerURI(uri string) string {
	url := strings.Split(uri, ":")[0] + ":8081/explorer/address"
	if !strings.HasPrefix(url, "http") {
		return "http://" + url
	}
	return url
}

func (w *waspServices) NetworkProvider() peering_pkg.NetworkProvider {
	return peering.DefaultNetworkProvider()
}

func (w *waspServices) TrustedNetworkManager() peering_pkg.TrustedNetworkManager {
	return peering.DefaultTrustedNetworkManager()
}

func configure(*node.Plugin) {
	log = logger.NewLogger(PluginName)

	Server.HideBanner = true
	Server.HidePort = true
	Server.Use(middleware.LoggerWithConfig(middleware.LoggerConfig{
		Format: `${time_rfc3339_nano} ${remote_ip} ${method} ${uri} ${status} error="${error}"` + "\n",
	}))
	Server.Use(middleware.Recover())
	auth.AddAuthentication(Server, parameters.GetStringToString(parameters.DashboardAuth))

	d = dashboard.Init(Server, &waspServices{})
}

func run(_ *node.Plugin) {
	log.Infof("Starting %s ...", PluginName)
	if err := daemon.BackgroundWorker(PluginName, worker); err != nil {
		log.Errorf("Error starting as daemon: %s", err)
	}
}

func worker(shutdownSignal <-chan struct{}) {
	stopped := make(chan struct{})
	go func() {
		defer close(stopped)
		bindAddr := parameters.GetString(parameters.DashboardBindAddress)
		log.Infof("%s started, bind address=%s", PluginName, bindAddr)
		if err := Server.Start(bindAddr); err != nil {
			if !errors.Is(err, http.ErrServerClosed) {
				log.Errorf("Error serving: %s", err)
			}
		}
	}()

	select {
	case <-shutdownSignal:
	case <-stopped:
	}

	log.Infof("Stopping %s ...", PluginName)
	defer log.Infof("Stopping %s ... done", PluginName)

	d.Stop()

	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()
	if err := Server.Shutdown(ctx); err != nil {
		log.Errorf("Error stopping: %s", err)
	}
}
