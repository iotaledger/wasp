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
	"github.com/iotaledger/wasp/packages/coretypes"
	"github.com/iotaledger/wasp/packages/dashboard"
	"github.com/iotaledger/wasp/packages/kv/dict"
	"github.com/iotaledger/wasp/packages/parameters"
	peering_pkg "github.com/iotaledger/wasp/packages/peering"
	"github.com/iotaledger/wasp/packages/registry"
	"github.com/iotaledger/wasp/packages/state"
	"github.com/iotaledger/wasp/packages/util/auth"
	"github.com/iotaledger/wasp/packages/vm/viewcontext"
	"github.com/iotaledger/wasp/plugins/chains"
	"github.com/iotaledger/wasp/plugins/config"
	"github.com/iotaledger/wasp/plugins/database"
	"github.com/iotaledger/wasp/plugins/peering"
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
	return config.Dump()
}

func (w *waspServices) ExploreAddressBaseURL() string {
	baseUrl := parameters.GetString(parameters.DashboardExploreAddressUrl)
	if baseUrl != "" {
		return baseUrl
	}
	return exploreAddressUrlFromGoshimmerUri(parameters.GetString(parameters.NodeAddress))
}

func (w *waspServices) GetChainRecords() ([]*registry.ChainRecord, error) {
	return registry.GetChainRecords()
}

func (w *waspServices) GetChainRecord(chainID *coretypes.ChainID) (*registry.ChainRecord, error) {
	return registry.ChainRecordFromRegistry(chainID)
}

func (w *waspServices) GetChainState(chainID *coretypes.ChainID) (*dashboard.ChainState, error) {
	// TODO get rid of database.GetInstance(), pass dbprovider as parameter
	virtualState, _, err := state.LoadSolidState(database.GetInstance(), chainID)
	if err != nil {
		return nil, err
	}
	block, err := state.LoadBlock(database.GetInstance(), chainID, virtualState.BlockIndex())
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

func (w *waspServices) GetChain(chainID *coretypes.ChainID) chain.Chain {
	return chains.AllChains().Get(chainID)
}

func (w *waspServices) CallView(chain chain.Chain, hname coretypes.Hname, fname string, params dict.Dict) (dict.Dict, error) {
	vctx, err := viewcontext.NewFromDB(database.GetInstance(), *chain.ID(), chain.Processors())
	if err != nil {
		return nil, fmt.Errorf(fmt.Sprintf("Failed to create context: %v", err))
	}

	ret, err := vctx.CallView(hname, coretypes.Hn(fname), params)
	if err != nil {
		return nil, fmt.Errorf("root view call failed: %v", err)
	}
	return ret, nil
}

func exploreAddressUrlFromGoshimmerUri(uri string) string {
	url := strings.Split(uri, ":")[0] + ":8081/explorer/address"
	if !strings.HasPrefix(url, "http") {
		return "http://" + url
	}
	return url
}

func (w *waspServices) NetworkProvider() peering_pkg.NetworkProvider {
	return peering.DefaultNetworkProvider()
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
