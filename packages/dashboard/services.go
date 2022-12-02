// Copyright 2020 IOTA Stiftung
// SPDX-License-Identifier: Apache-2.0

package dashboard

import (
	"net/http"
	"strings"

	"github.com/labstack/echo/v4"

	"github.com/iotaledger/hive.go/core/configuration"
	"github.com/iotaledger/wasp/packages/chain"
	"github.com/iotaledger/wasp/packages/chains"
	"github.com/iotaledger/wasp/packages/isc"
	"github.com/iotaledger/wasp/packages/kv/dict"
	"github.com/iotaledger/wasp/packages/metrics/nodeconnmetrics"
	"github.com/iotaledger/wasp/packages/peering"
	"github.com/iotaledger/wasp/packages/registry"
	"github.com/iotaledger/wasp/packages/vm/viewcontext"
)

type WaspServices struct {
	webAPIBindAddress           string
	exploreAddressURL           string
	config                      *configuration.Configuration
	chains                      *chains.Chains
	chainRecordRegistryProvider registry.ChainRecordRegistryProvider
	networkProvider             peering.NetworkProvider
	trustedNetworkManager       peering.TrustedNetworkManager
}

func NewWaspServices(
	webAPIBindAddress string,
	exploreAddressURL string,
	config *configuration.Configuration,
	chains *chains.Chains,
	chainRecordRegistryProvider registry.ChainRecordRegistryProvider,
	networkProvider peering.NetworkProvider,
	trustedNetworkManager peering.TrustedNetworkManager,
) *WaspServices {
	return &WaspServices{
		webAPIBindAddress:           webAPIBindAddress,
		exploreAddressURL:           exploreAddressURL,
		config:                      config,
		chains:                      chains,
		chainRecordRegistryProvider: chainRecordRegistryProvider,
		networkProvider:             networkProvider,
		trustedNetworkManager:       trustedNetworkManager,
	}
}

func (w *WaspServices) ConfigDump() map[string]interface{} {
	return w.config.Koanf().All()
}

func (w *WaspServices) WebAPIPort() string {
	port := "80"
	parts := strings.Split(w.webAPIBindAddress, ":")
	if len(parts) == 2 {
		port = parts[1]
	}
	return port
}

func (w *WaspServices) ExploreAddressBaseURL() string {
	return w.exploreAddressURL
}

func (w *WaspServices) PeeringStats() (*PeeringStats, error) {
	ret := &PeeringStats{}
	peers := w.networkProvider.PeerStatus()
	ret.Peers = make([]Peer, len(peers))
	for i, p := range peers {
		ret.Peers[i] = Peer{
			NumUsers: p.NumUsers(),
			NetID:    p.NetID(),
			IsAlive:  p.IsAlive(),
		}
	}
	tpeers, err := w.trustedNetworkManager.TrustedPeers()
	if err != nil {
		return nil, err
	}
	ret.TrustedPeers = make([]TrustedPeer, len(tpeers))
	for i, t := range tpeers {
		ret.TrustedPeers[i] = TrustedPeer{
			NetID:  t.NetID,
			PubKey: *t.PubKey(),
		}
	}
	return ret, nil
}

func (w *WaspServices) MyNetworkID() string {
	return w.networkProvider.Self().NetID()
}

func (w *WaspServices) ChainRecords() ([]*registry.ChainRecord, error) {
	return w.chainRecordRegistryProvider.ChainRecords()
}

func (w *WaspServices) GetChainRecord(chainID *isc.ChainID) (*registry.ChainRecord, error) {
	ch, err := w.chainRecordRegistryProvider.ChainRecord(*chainID)
	if err != nil {
		return nil, err
	}
	if ch == nil {
		return nil, echo.NewHTTPError(http.StatusNotFound, "Chain record not found")
	}
	return ch, nil
}

func (w *WaspServices) GetChainCommitteeInfo(chainID *isc.ChainID) (*chain.CommitteeInfo, error) {
	ch := w.chains.Get(chainID)
	if ch == nil {
		return nil, echo.NewHTTPError(http.StatusNotFound, "Chain not found")
	}
	return ch.GetCommitteeInfo(), nil
}

func (w *WaspServices) GetChainNodeConnectionMetrics(chainID *isc.ChainID) (nodeconnmetrics.NodeConnectionMessagesMetrics, error) {
	ch := w.chains.Get(chainID)
	if ch == nil {
		return nil, echo.NewHTTPError(http.StatusNotFound, "Chain not found")
	}
	return ch.GetNodeConnectionMetrics(), nil
}

func (w *WaspServices) GetNodeConnectionMetrics() (nodeconnmetrics.NodeConnectionMetrics, error) {
	return w.chains.GetNodeConnectionMetrics(), nil
}

func (w *WaspServices) GetChainConsensusWorkflowStatus(chainID *isc.ChainID) (chain.ConsensusWorkflowStatus, error) {
	ch := w.chains.Get(chainID)
	if ch == nil {
		return nil, echo.NewHTTPError(http.StatusNotFound, "Chain not found")
	}
	return ch.GetConsensusWorkflowStatus(), nil
}

func (w *WaspServices) GetChainConsensusPipeMetrics(chainID *isc.ChainID) (chain.ConsensusPipeMetrics, error) {
	ch := w.chains.Get(chainID)
	if ch == nil {
		return nil, echo.NewHTTPError(http.StatusNotFound, "Chain not found")
	}
	return ch.GetConsensusPipeMetrics(), nil
}

func (w *WaspServices) CallView(chainID *isc.ChainID, scName, funName string, params dict.Dict) (dict.Dict, error) {
	ch := w.chains.Get(chainID)
	if ch == nil {
		return nil, echo.NewHTTPError(http.StatusNotFound, "Chain not found")
	}
	vctx := viewcontext.New(ch, ch.LatestBlockIndex())
	return vctx.CallViewExternal(isc.Hn(scName), isc.Hn(funName), params)
}

var _ WaspServicesInterface = &WaspServices{}
