// Copyright 2020 IOTA Stiftung
// SPDX-License-Identifier: Apache-2.0

package admapi

import (
	"net/http"

	"github.com/iotaledger/hive.go/crypto/ed25519"
	"github.com/iotaledger/wasp/packages/peering"
	"github.com/iotaledger/wasp/packages/webapi/httperrors"
	"github.com/iotaledger/wasp/packages/webapi/model"
	"github.com/iotaledger/wasp/packages/webapi/routes"
	"github.com/labstack/echo/v4"
	"github.com/pangpanglabs/echoswagger/v2"
)

func addPeeringEndpoints(adm echoswagger.ApiGroup, tnm peering.TrustedNetworkManager) {
	listExample := []*model.PeeringTrustedNode{
		{PubKey: "8mcS4hUaiiedX3jRud41Zuu1ZcRUZZ8zY9SuJJgXHuiQ", NetID: "some-host:9081"},
		{PubKey: "8mcS4hUaiiedX3jRud41Zuu1ZcRUZZ8zY9SuJJgXHuiR", NetID: "some-host:9082"},
	}
	addTnm := func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c echo.Context) error {
			c.Set("tnm", tnm)
			return next(c)
		}
	}
	adm.GET(routes.PeeringTrustedList(), handlePeeringTrustedList, addTnm).
		AddResponse(http.StatusOK, "A list of trusted peers.", listExample, nil).
		SetSummary("Get a list of trusted peers.")

	adm.GET(routes.PeeringTrustedGet(":pubKey"), handlePeeringTrustedGet, addTnm).
		AddParamPath(listExample[0].PubKey, "pubKey", "Public key of the trusted peer (base58).").
		AddResponse(http.StatusOK, "Trusted peer info.", listExample[0], nil).
		SetSummary("Get details on a particular trusted peer.")

	adm.PUT(routes.PeeringTrustedPut(":pubKey"), handlePeeringTrustedPut, addTnm).
		AddParamPath(listExample[0].PubKey, "pubKey", "Public key of the trusted peer (base58).").
		AddParamBody(listExample[0], "PeeringTrustedNode", "Info of the peer to trust.", true).
		AddResponse(http.StatusOK, "Trusted peer info.", listExample[0], nil).
		SetSummary("Trust the specified peer, the pub key is passed via the path.")

	adm.POST(routes.PeeringTrustedPost(), handlePeeringTrustedPost, addTnm).
		AddParamBody(listExample[0], "PeeringTrustedNode", "Info of the peer to trust.", true).
		AddResponse(http.StatusOK, "Trusted peer info.", listExample[0], nil).
		SetSummary("Trust the specified peer.")

	adm.DELETE(routes.PeeringTrustedDelete(":pubKey"), handlePeeringTrustedDelete, addTnm).
		AddParamPath(listExample[0].PubKey, "pubKey", "Public key of the trusted peer (base58).").
		SetSummary("Distrust the specified peer.")
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
	pubKey, err := ed25519.PublicKeyFromString(req.PubKey)
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
	pubKey, err := ed25519.PublicKeyFromString(req.PubKey)
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
	pubKey, err := ed25519.PublicKeyFromString(pubKeyStr)
	if err != nil {
		return httperrors.BadRequest(err.Error())
	}
	tps, err := tnm.TrustedPeers()
	if err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, err)
	}
	for _, tp := range tps {
		if tp.PubKey == pubKey {
			return c.JSON(http.StatusOK, model.NewPeeringTrustedNode(tp))
		}
	}
	return httperrors.NotFound("peer not trusted")
}

func handlePeeringTrustedDelete(c echo.Context) error {
	var err error
	tnm := c.Get("tnm").(peering.TrustedNetworkManager)
	pubKeyStr := c.Param("pubKey")
	pubKey, err := ed25519.PublicKeyFromString(pubKeyStr)
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
