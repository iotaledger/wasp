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

	"github.com/iotaledger/wasp/packages/kv/optimism"

	"github.com/iotaledger/hive.go/daemon"
	"github.com/iotaledger/hive.go/logger"
	"github.com/iotaledger/hive.go/node"
	"github.com/iotaledger/wasp/packages/chain"
	"github.com/iotaledger/wasp/packages/coretypes"
	"github.com/iotaledger/wasp/packages/dashboard"
	"github.com/iotaledger/wasp/packages/kv/dict"
	"github.com/iotaledger/wasp/packages/parameters"
	peering_pkg "github.com/iotaledger/wasp/packages/peering"
	"github.com/iotaledger/wasp/packages/registry/chainrecord"
	"github.com/iotaledger/wasp/packages/state"
	"github.com/iotaledger/wasp/packages/util/auth"
	"github.com/iotaledger/wasp/packages/vm/viewcontext"
	"github.com/iotaledger/wasp/plugins/chains"
	"github.com/iotaledger/wasp/plugins/config"
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
	return config.Dump()
}

func (w *waspServices) ExploreAddressBaseURL() string {
	baseUrl := parameters.GetString(parameters.DashboardExploreAddressUrl)
	if baseUrl != "" {
		return baseUrl
	}
	return exploreAddressUrlFromGoshimmerUri(parameters.GetString(parameters.NodeAddress))
}

func (w *waspServices) GetChainRecords() ([]*chainrecord.ChainRecord, error) {
	return registry.DefaultRegistry().GetChainRecords()
}

func (w *waspServices) GetChainRecord(chainID *coretypes.ChainID) (*chainrecord.ChainRecord, error) {
	return registry.DefaultRegistry().GetChainRecordByChainID(chainID)
}

func (w *waspServices) GetChainState(chainID *coretypes.ChainID) (*dashboard.ChainState, error) {
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

func (w *waspServices) GetChain(chainID *coretypes.ChainID) chain.Chain {
	return chains.AllChains().Get(chainID)
}

func (w *waspServices) CallView(chain chain.Chain, hname coretypes.Hname, funName string, params dict.Dict) (dict.Dict, error) {
	vctx := viewcontext.NewFromChain(chain)
	var err error
	var ret dict.Dict
	_ = optimism.RepeatOnceIfUnlucky(func() error {
		ret, err = vctx.CallView(hname, coretypes.Hn(funName), params)
		return err
	})
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
