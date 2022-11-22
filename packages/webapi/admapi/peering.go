// Copyright 2020 IOTA Stiftung
// SPDX-License-Identifier: Apache-2.0

package admapi

import (
	"net/http"

	"github.com/labstack/echo/v4"
	"github.com/pangpanglabs/echoswagger/v2"

	iotago "github.com/iotaledger/iota.go/v3"
	"github.com/iotaledger/wasp/packages/cryptolib"
	"github.com/iotaledger/wasp/packages/peering"
	"github.com/iotaledger/wasp/packages/webapi/httperrors"
	"github.com/iotaledger/wasp/packages/webapi/model"
	"github.com/iotaledger/wasp/packages/webapi/routes"
)

func addPeeringEndpoints(adm echoswagger.ApiGroup, network peering.NetworkProvider, tnm peering.TrustedNetworkManager) {
	listExample := []*model.PeeringTrustedNode{
		{PubKey: "8mcS4hUaiiedX3jRud41Zuu1ZcRUZZ8zY9SuJJgXHuiQ", NetID: "some-host:9081"},
		{PubKey: "8mcS4hUaiiedX3jRud41Zuu1ZcRUZZ8zY9SuJJgXHuiR", NetID: "some-host:9082"},
	}
	peeringStatusExample := []*model.PeeringNodeStatus{
		{PubKey: "8mcS4hUaiiedX3jRud41Zuu1ZcRUZZ8zY9SuJJgXHuiQ", IsAlive: true, NumUsers: 1, NetID: "some-host:9081"},
		{PubKey: "8mcS4hUaiiedX3jRud41Zuu1ZcRUZZ8zY9SuJJgXHuiR", IsAlive: true, NumUsers: 1, NetID: "some-host:9082"},
	}

	addCtx := func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c echo.Context) error {
			c.Set("net", network)
			c.Set("tnm", tnm)
			return next(c)
		}
	}

	adm.GET(routes.PeeringSelfGet(), handlePeeringSelfGet, addCtx).
		AddResponse(http.StatusOK, "This node as a peer.", listExample[0], nil).
		SetSummary("Basic peer info of the current node.")

	adm.GET(routes.PeeringTrustedList(), handlePeeringTrustedList, addCtx).
		AddResponse(http.StatusOK, "A list of trusted peers.", listExample, nil).
		SetSummary("Get a list of trusted peers.")

	adm.GET(routes.PeeringTrustedGet(":pubKey"), handlePeeringTrustedGet, addCtx).
		AddParamPath(listExample[0].PubKey, "pubKey", "Public key of the trusted peer (hex).").
		AddResponse(http.StatusOK, "Trusted peer info.", listExample[0], nil).
		SetSummary("Get details on a particular trusted peer.")

	adm.PUT(routes.PeeringTrustedPut(":pubKey"), handlePeeringTrustedPut, addCtx).
		AddParamPath(listExample[0].PubKey, "pubKey", "Public key of the trusted peer (hex).").
		AddParamBody(listExample[0], "PeeringTrustedNode", "Info of the peer to trust.", true).
		AddResponse(http.StatusOK, "Trusted peer info.", listExample[0], nil).
		SetSummary("Trust the specified peer, the pub key is passed via the path.")

	adm.GET(routes.PeeringGetStatus(), handlePeeringGetStatus, addCtx).
		AddResponse(http.StatusOK, "A list of all peers.", peeringStatusExample, nil).
		SetSummary("Basic information about all configured peers.")

	adm.POST(routes.PeeringTrustedPost(), handlePeeringTrustedPost, addCtx).
		AddParamBody(listExample[0], "PeeringTrustedNode", "Info of the peer to trust.", true).
		AddResponse(http.StatusOK, "Trusted peer info.", listExample[0], nil).
		SetSummary("Trust the specified peer.")

	adm.DELETE(routes.PeeringTrustedDelete(":pubKey"), handlePeeringTrustedDelete, addCtx).
		AddParamPath(listExample[0].PubKey, "pubKey", "Public key of the trusted peer (hex).").
		SetSummary("Distrust the specified peer.")
}

func handlePeeringSelfGet(c echo.Context) error {
	network := c.Get("net").(peering.NetworkProvider)
	resp := model.PeeringTrustedNode{
		PubKey: iotago.EncodeHex(network.Self().PubKey().AsBytes()),
		NetID:  network.Self().NetID(),
	}
	return c.JSON(http.StatusOK, resp)
}

func handlePeeringGetStatus(c echo.Context) error {
	network := c.Get("net").(peering.NetworkProvider)
	peeringStatus := network.PeerStatus()

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

func handlePeeringTrustedList(c echo.Context) error {
	tnm := c.Get("tnm").(peering.TrustedNetworkManager)
	trustedPeers, err := tnm.TrustedPeers()
	if err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, err)
	}
	response := make([]*model.PeeringTrustedNode, len(trustedPeers))
	for i := range trustedPeers {
		response[i] = model.NewPeeringTrustedNode(trustedPeers[i])
	}
	return c.JSON(http.StatusOK, response)
}

func handlePeeringTrustedPut(c echo.Context) error {
	var err error
	tnm := c.Get("tnm").(peering.TrustedNetworkManager)
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
	pubKey, err := cryptolib.NewPublicKeyFromHexString(req.PubKey)
	if err != nil {
		return httperrors.BadRequest(err.Error())
	}
	tp, err := tnm.TrustPeer(pubKey, req.NetID)
	if err != nil {
		return httperrors.BadRequest(err.Error())
	}
	return c.JSON(http.StatusOK, model.NewPeeringTrustedNode(tp))
}

func handlePeeringTrustedPost(c echo.Context) error {
	var err error
	tnm := c.Get("tnm").(peering.TrustedNetworkManager)
	req := model.PeeringTrustedNode{}
	if err = c.Bind(&req); err != nil {
		return httperrors.BadRequest("Invalid request body.")
	}
	pubKey, err := cryptolib.NewPublicKeyFromHexString(req.PubKey)
	if err != nil {
		return httperrors.BadRequest(err.Error())
	}
	tp, err := tnm.TrustPeer(pubKey, req.NetID)
	if err != nil {
		return httperrors.BadRequest(err.Error())
	}
	return c.JSON(http.StatusOK, model.NewPeeringTrustedNode(tp))
}

func handlePeeringTrustedGet(c echo.Context) error {
	var err error
	tnm := c.Get("tnm").(peering.TrustedNetworkManager)
	pubKeyStr := c.Param("pubKey")
	pubKey, err := cryptolib.NewPublicKeyFromHexString(pubKeyStr)
	if err != nil {
		return httperrors.BadRequest(err.Error())
	}
	tps, err := tnm.TrustedPeers()
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

func handlePeeringTrustedDelete(c echo.Context) error {
	var err error
	tnm := c.Get("tnm").(peering.TrustedNetworkManager)
	pubKeyStr := c.Param("pubKey")
	pubKey, err := cryptolib.NewPublicKeyFromHexString(pubKeyStr)
	if err != nil {
		return httperrors.BadRequest(err.Error())
	}
	tp, err := tnm.DistrustPeer(pubKey)
	if err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, err)
	}
	if tp == nil {
		return c.NoContent(http.StatusOK)
	}
	return c.JSON(http.StatusOK, model.NewPeeringTrustedNode(tp))
}
