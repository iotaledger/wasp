// Copyright 2020 IOTA Stiftung
// SPDX-License-Identifier: Apache-2.0

package admapi

import (
	"fmt"
	"net/http"

	"github.com/labstack/echo/v4"
	"github.com/pangpanglabs/echoswagger/v2"
	"github.com/samber/lo"

	iotago "github.com/iotaledger/iota.go/v3"
	"github.com/iotaledger/wasp/packages/cryptolib"
	"github.com/iotaledger/wasp/packages/peering"
	"github.com/iotaledger/wasp/packages/registry"
	"github.com/iotaledger/wasp/packages/webapi/v1/httperrors"
	"github.com/iotaledger/wasp/packages/webapi/v1/model"
	"github.com/iotaledger/wasp/packages/webapi/v1/routes"
)

type peeringService struct {
	chainRecordRegistryProvider registry.ChainRecordRegistryProvider
	network                     peering.NetworkProvider
	networkMgr                  peering.TrustedNetworkManager
}

func addPeeringEndpoints(adm echoswagger.ApiGroup, chainRecordRegistryProvider registry.ChainRecordRegistryProvider, network peering.NetworkProvider, tnm peering.TrustedNetworkManager) {
	listExample := []*model.PeeringTrustedNode{
		{PubKey: "8mcS4hUaiiedX3jRud41Zuu1ZcRUZZ8zY9SuJJgXHuiQ", NetID: "some-host:9081"},
		{PubKey: "8mcS4hUaiiedX3jRud41Zuu1ZcRUZZ8zY9SuJJgXHuiR", NetID: "some-host:9082"},
	}
	peeringStatusExample := []*model.PeeringNodeStatus{
		{PubKey: "8mcS4hUaiiedX3jRud41Zuu1ZcRUZZ8zY9SuJJgXHuiQ", IsAlive: true, NumUsers: 1, NetID: "some-host:9081"},
		{PubKey: "8mcS4hUaiiedX3jRud41Zuu1ZcRUZZ8zY9SuJJgXHuiR", IsAlive: true, NumUsers: 1, NetID: "some-host:9082"},
	}
	p := &peeringService{
		chainRecordRegistryProvider: chainRecordRegistryProvider,
		network:                     network,
		networkMgr:                  tnm,
	}

	adm.GET(routes.PeeringSelfGet(), p.handlePeeringSelfGet).
		AddResponse(http.StatusOK, "This node as a peer.", listExample[0], nil).
		SetSummary("Basic peer info of the current node.")

	adm.GET(routes.PeeringTrustedList(), p.handlePeeringTrustedList).
		AddResponse(http.StatusOK, "A list of trusted peers.", listExample, nil).
		SetSummary("Get a list of trusted peers.")

	adm.GET(routes.PeeringTrustedGet(":pubKey"), p.handlePeeringTrustedGet).
		AddParamPath(listExample[0].PubKey, "pubKey", "Public key of the trusted peer (hex).").
		AddResponse(http.StatusOK, "Trusted peer info.", listExample[0], nil).
		SetSummary("Get details on a particular trusted peer.")

	adm.PUT(routes.PeeringTrustedPut(":pubKey"), p.handlePeeringTrustedPut).
		AddParamPath(listExample[0].PubKey, "pubKey", "Public key of the trusted peer (hex).").
		AddParamBody(listExample[0], "PeeringTrustedNode", "Info of the peer to trust.", true).
		AddResponse(http.StatusOK, "Trusted peer info.", listExample[0], nil).
		SetSummary("Trust the specified peer, the pub key is passed via the path.")

	adm.GET(routes.PeeringGetStatus(), p.handlePeeringGetStatus).
		AddResponse(http.StatusOK, "A list of all peers.", peeringStatusExample, nil).
		SetSummary("Basic information about all configured peers.")

	adm.POST(routes.PeeringTrustedPost(), p.handlePeeringTrustedPost).
		AddParamBody(listExample[0], "PeeringTrustedNode", "Info of the peer to trust.", true).
		AddResponse(http.StatusOK, "Trusted peer info.", listExample[0], nil).
		SetSummary("Trust the specified peer.")

	adm.DELETE(routes.PeeringTrustedDelete(":pubKey"), p.handlePeeringTrustedDelete).
		AddParamPath(listExample[0].PubKey, "pubKey", "Public key of the trusted peer (hex).").
		SetSummary("Distrust the specified peer.")
}

func (p *peeringService) handlePeeringSelfGet(c echo.Context) error {
	resp := model.PeeringTrustedNode{
		PubKey: iotago.EncodeHex(p.network.Self().PubKey().AsBytes()),
		NetID:  p.network.Self().NetID(),
	}
	return c.JSON(http.StatusOK, resp)
}

func (p *peeringService) handlePeeringGetStatus(c echo.Context) error {
	peeringStatus := p.network.PeerStatus()

	peers := make([]model.PeeringNodeStatus, len(peeringStatus))

	for k, v := range peeringStatus {
		peers[k] = model.PeeringNodeStatus{
			PubKey:   iotago.EncodeHex(v.PubKey().AsBytes()),
			NetID:    v.NetID(),
			IsAlive:  v.IsAlive(),
			NumUsers: v.NumUsers(),
		}
	}

	return c.JSON(http.StatusOK, peers)
}

func (p *peeringService) handlePeeringTrustedList(c echo.Context) error {
	trustedPeers, err := p.networkMgr.TrustedPeers()
	if err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, err)
	}
	response := make([]*model.PeeringTrustedNode, len(trustedPeers))
	for i := range trustedPeers {
		response[i] = model.NewPeeringTrustedNode(trustedPeers[i])
	}
	return c.JSON(http.StatusOK, response)
}

func (p *peeringService) handlePeeringTrustedPut(c echo.Context) error {
	var err error
	pubKeyStr := c.Param("pubKey")
	req := model.PeeringTrustedNode{}
	if err = c.Bind(&req); err != nil {
		return httperrors.BadRequest("Invalid request body.")
	}
	if req.PubKey == "" {
		req.PubKey = pubKeyStr
	}
	if req.PubKey != pubKeyStr {
		return httperrors.BadRequest("Pub keys do not match.")
	}
	pubKey, err := cryptolib.NewPublicKeyFromString(req.PubKey)
	if err != nil {
		return httperrors.BadRequest(err.Error())
	}
	tp, err := p.networkMgr.TrustPeer(pubKey, req.NetID)
	if err != nil {
		return httperrors.BadRequest(err.Error())
	}
	return c.JSON(http.StatusOK, model.NewPeeringTrustedNode(tp))
}

func (p *peeringService) handlePeeringTrustedPost(c echo.Context) error {
	var err error
	req := model.PeeringTrustedNode{}
	if err = c.Bind(&req); err != nil {
		return httperrors.BadRequest("Invalid request body.")
	}
	pubKey, err := cryptolib.NewPublicKeyFromString(req.PubKey)
	if err != nil {
		return httperrors.BadRequest(err.Error())
	}
	tp, err := p.networkMgr.TrustPeer(pubKey, req.NetID)
	if err != nil {
		return httperrors.BadRequest(err.Error())
	}
	return c.JSON(http.StatusOK, model.NewPeeringTrustedNode(tp))
}

func (p *peeringService) handlePeeringTrustedGet(c echo.Context) error {
	var err error
	pubKeyStr := c.Param("pubKey")
	pubKey, err := cryptolib.NewPublicKeyFromString(pubKeyStr)
	if err != nil {
		return httperrors.BadRequest(err.Error())
	}
	tps, err := p.networkMgr.TrustedPeers()
	if err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, err)
	}
	for _, tp := range tps {
		if tp.PubKey().Equals(pubKey) {
			return c.JSON(http.StatusOK, model.NewPeeringTrustedNode(tp))
		}
	}
	return httperrors.NotFound("peer not trusted")
}

func (p *peeringService) handlePeeringTrustedDelete(c echo.Context) error {
	var err error
	pubKeyStr := c.Param("pubKey")
	pubKey, err := cryptolib.NewPublicKeyFromString(pubKeyStr)
	if err != nil {
		return httperrors.BadRequest(err.Error())
	}
	tp, err := p.networkMgr.DistrustPeer(pubKey)
	if err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, err)
	}
	if tp == nil {
		return c.NoContent(http.StatusOK)
	}
	// remove any access nodes for the distrusted peer
	chainRecs, err := p.chainRecordRegistryProvider.ChainRecords()
	if err != nil {
		return httperrors.ServerError("Peer trust removed, but errored when trying to get chain list from registry")
	}
	chainRecsToModify := lo.Filter(chainRecs, func(r *registry.ChainRecord, _ int) bool {
		return lo.ContainsBy(r.AccessNodes, func(p *cryptolib.PublicKey) bool {
			return p.Equals(pubKey)
		})
	})
	for _, r := range chainRecsToModify {
		_, err = p.chainRecordRegistryProvider.UpdateChainRecord(r.ID(), func(rec *registry.ChainRecord) bool {
			return rec.RemoveAccessNode(pubKey)
		})
		if err != nil {
			return httperrors.ServerError(fmt.Sprintf("Peer trust removed, but errored when trying to save chain record %s", r.ChainID()))
		}
	}
	return c.JSON(http.StatusOK, model.NewPeeringTrustedNode(tp))
}
