// Copyright 2020 IOTA Stiftung
// SPDX-License-Identifier: Apache-2.0

package dashboard

import (
	"context"
	"errors"
	"net/http"
	"strings"
	"time"

	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"
	"github.com/labstack/gommon/log"
	"go.uber.org/dig"

	"github.com/iotaledger/hive.go/core/app"
	"github.com/iotaledger/wasp/core/chains"
	"github.com/iotaledger/wasp/core/peering"
	"github.com/iotaledger/wasp/core/registry"
	"github.com/iotaledger/wasp/packages/authentication"
	"github.com/iotaledger/wasp/packages/authentication/shared/permissions"
	"github.com/iotaledger/wasp/packages/chain"
	"github.com/iotaledger/wasp/packages/dashboard"
	"github.com/iotaledger/wasp/packages/isc"
	"github.com/iotaledger/wasp/packages/kv/dict"
	"github.com/iotaledger/wasp/packages/kv/optimism"
	"github.com/iotaledger/wasp/packages/metrics/nodeconnmetrics"
	registry_pkg "github.com/iotaledger/wasp/packages/registry"
	"github.com/iotaledger/wasp/packages/users"
	"github.com/iotaledger/wasp/packages/vm/viewcontext"
	"github.com/iotaledger/wasp/plugins/webapi"
)

func init() {
	Plugin = &app.Plugin{
		Component: &app.Component{
			Name:      "Dashboard",
			DepsFunc:  func(cDeps dependencies) { deps = cDeps },
			Params:    params,
			Configure: configure,
			Run:       run,
		},
		IsEnabled: func() bool {
			return ParamsDashboard.Enabled
		},
	}
}

var (
	Plugin *app.Plugin
	deps   dependencies

	Server = echo.New()
	d      *dashboard.Dashboard
)

type dependencies struct {
	dig.In

	UserManager *users.UserManager
}

func configure() error {
	Server.HideBanner = true
	Server.HidePort = true
	Server.Use(middleware.LoggerWithConfig(middleware.LoggerConfig{
		Format: `${time_rfc3339_nano} ${remote_ip} ${method} ${uri} ${status} error="${error}"` + "\n",
	}))
	Server.Use(middleware.Recover())

	claimValidator := func(claims *authentication.WaspClaims) bool {
		// The Dashboard will be accessible if the token has a 'Dashboard' claim
		return claims.HasPermission(permissions.Dashboard)
	}

	authentication.AddAuthentication(Server, deps.UserManager, registry.DefaultRegistry, ParamsDashboard.Auth, claimValidator)

	d = dashboard.Init(Server, &waspServices{}, Plugin.Logger())

	return nil
}

func run() error {
	log.Infof("Starting %s ...", Plugin.Name)
	if err := Plugin.Daemon().BackgroundWorker(Plugin.Name, worker); err != nil {
		log.Errorf("error starting as daemon: %s", err)
	}

	return nil
}

type waspServices struct{}

var _ dashboard.WaspServices = &waspServices{}

func (w *waspServices) ConfigDump() map[string]interface{} {
	return Plugin.App().Config().Koanf().All()
}

func (*waspServices) WebAPIPort() string {
	port := "80"
	parts := strings.Split(webapi.ParamsWebAPI.BindAddress, ":")
	if len(parts) == 2 {
		port = parts[1]
	}
	return port
}

func (w *waspServices) ExploreAddressBaseURL() string {
	return ParamsDashboard.ExploreAddressURL
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
			PubKey: *t.PubKey,
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

func (w *waspServices) GetChainRecord(chainID *isc.ChainID) (*registry_pkg.ChainRecord, error) {
	ch, err := registry.DefaultRegistry().GetChainRecordByChainID(chainID)
	if err != nil {
		return nil, err
	}
	if ch == nil {
		return nil, echo.NewHTTPError(http.StatusNotFound, "Chain record not found")
	}
	return ch, nil
}

func (w *waspServices) GetChainCommitteeInfo(chainID *isc.ChainID) (*chain.CommitteeInfo, error) {
	ch := chains.AllChains().Get(chainID)
	if ch == nil {
		return nil, echo.NewHTTPError(http.StatusNotFound, "Chain not found")
	}
	return ch.GetCommitteeInfo(), nil
}

func (w *waspServices) GetChainNodeConnectionMetrics(chainID *isc.ChainID) (nodeconnmetrics.NodeConnectionMessagesMetrics, error) {
	ch := chains.AllChains().Get(chainID)
	if ch == nil {
		return nil, echo.NewHTTPError(http.StatusNotFound, "Chain not found")
	}
	return ch.GetNodeConnectionMetrics(), nil
}

func (w *waspServices) GetNodeConnectionMetrics() (nodeconnmetrics.NodeConnectionMetrics, error) {
	chs := chains.AllChains()
	return chs.GetNodeConnectionMetrics(), nil
}

func (w *waspServices) GetChainConsensusWorkflowStatus(chainID *isc.ChainID) (chain.ConsensusWorkflowStatus, error) {
	ch := chains.AllChains().Get(chainID)
	if ch == nil {
		return nil, echo.NewHTTPError(http.StatusNotFound, "Chain not found")
	}
	return ch.GetConsensusWorkflowStatus(), nil
}

func (w *waspServices) GetChainConsensusPipeMetrics(chainID *isc.ChainID) (chain.ConsensusPipeMetrics, error) {
	ch := chains.AllChains().Get(chainID)
	if ch == nil {
		return nil, echo.NewHTTPError(http.StatusNotFound, "Chain not found")
	}
	return ch.GetConsensusPipeMetrics(), nil
}

func (w *waspServices) CallView(chainID *isc.ChainID, scName, funName string, params dict.Dict) (dict.Dict, error) {
	ch := chains.AllChains().Get(chainID)
	if ch == nil {
		return nil, echo.NewHTTPError(http.StatusNotFound, "Chain not found")
	}
	vctx := viewcontext.New(ch)
	var ret dict.Dict
	err := optimism.RetryOnStateInvalidated(func() error {
		var err error
		ret, err = vctx.CallViewExternal(isc.Hn(scName), isc.Hn(funName), params)
		return err
	})
	return ret, err
}

func worker(ctx context.Context) {
	stopped := make(chan struct{})
	go func() {
		defer close(stopped)
		bindAddr := ParamsDashboard.BindAddress
		log.Infof("%s started, bind address=%s", Plugin.Name, bindAddr)
		if err := Server.Start(bindAddr); err != nil {
			if !errors.Is(err, http.ErrServerClosed) {
				log.Errorf("error serving: %s", err)
			}
		}
	}()

	select {
	case <-ctx.Done():
	case <-stopped:
	}

	log.Infof("Stopping %s ...", Plugin.Name)
	defer log.Infof("Stopping %s ... done", Plugin.Name)

	d.Stop()

	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()
	if err := Server.Shutdown(ctx); err != nil { //nolint:contextcheck // false positive
		log.Errorf("error stopping: %s", err)
	}
}
