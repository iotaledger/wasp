// Copyright 2020 IOTA Stiftung
// SPDX-License-Identifier: Apache-2.0

package dashboard

import (
	"context"
	"errors"
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
	registry_pkg "github.com/iotaledger/wasp/packages/registry"
	"github.com/iotaledger/wasp/packages/util/auth"
	"github.com/iotaledger/wasp/packages/vm/viewcontext"
	"github.com/iotaledger/wasp/plugins/chains"
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

func (w *waspServices) PeeringStats() (*dashboard.PeeringStats, error) {
	ret := &dashboard.PeeringStats{}
	peers := peering.DefaultNetworkProvider().PeerStatus()
	ret.Peers = make([]dashboard.Peer, len(peers))
	for i, p := range peers {
		ret.Peers[i] = dashboard.Peer{
			NumUsers: p.NumUsers(),
			NetID:    p.NetID(),
			IsAlive:  p.IsAlive(),
		}
	}
	tpeers, err := peering.DefaultTrustedNetworkManager().TrustedPeers()
	if err != nil {
		return nil, err
	}
	ret.TrustedPeers = make([]dashboard.TrustedPeer, len(tpeers))
	for i, t := range tpeers {
		ret.TrustedPeers[i] = dashboard.TrustedPeer{
			NetID:  t.NetID,
			PubKey: t.PubKey,
		}
	}
	return ret, nil
}

func (w *waspServices) MyNetworkID() string {
	return peering.DefaultNetworkProvider().Self().NetID()
}

func (w *waspServices) GetChainRecords() ([]*registry_pkg.ChainRecord, error) {
	return registry.DefaultRegistry().GetChainRecords()
}

func (w *waspServices) GetChainRecord(chainID *iscp.ChainID) (*registry_pkg.ChainRecord, error) {
	ch, err := registry.DefaultRegistry().GetChainRecordByChainID(chainID)
	if err != nil {
		return nil, err
	}
	if ch == nil {
		return nil, echo.NewHTTPError(http.StatusNotFound, "Chain record not found")
	}
	return ch, nil
}

func (w *waspServices) GetChainCommitteeInfo(chainID *iscp.ChainID) (*chain.CommitteeInfo, error) {
	ch := chains.AllChains().Get(chainID)
	if ch == nil {
		return nil, echo.NewHTTPError(http.StatusNotFound, "Chain not found")
	}
	return ch.GetCommitteeInfo(), nil
}

func (w *waspServices) CallView(chainID *iscp.ChainID, scName, funName string, params dict.Dict) (dict.Dict, error) {
	ch := chains.AllChains().Get(chainID)
	if ch == nil {
		return nil, echo.NewHTTPError(http.StatusNotFound, "Chain not found")
	}
	vctx := viewcontext.NewFromChain(ch)
	var ret dict.Dict
	err := optimism.RetryOnStateInvalidated(func() error {
		var err error
		ret, err = vctx.CallView(iscp.Hn(scName), iscp.Hn(funName), params)
		return err
	})
	return ret, err
}

func exploreAddressURLFromGoshimmerURI(uri string) string {
	url := strings.Split(uri, ":")[0] + ":8081/explorer/address"
	if !strings.HasPrefix(url, "http") {
		return "http://" + url
	}
	return url
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

	d = dashboard.Init(Server, &waspServices{}, log)
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
