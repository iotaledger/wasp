// Copyright 2020 IOTA Stiftung
// SPDX-License-Identifier: Apache-2.0

package admapi

import (
	"net/http"

	"github.com/iotaledger/wasp/packages/webapi/model"
	"github.com/iotaledger/wasp/packages/webapi/routes"
	"github.com/labstack/echo/v4"
	"github.com/pangpanglabs/echoswagger/v2"
)

func addPeeringEndpoints(adm echoswagger.ApiGroup) {
	listExample := []*model.PeeringTrustedNode{
		{PubKey: "8mcS4hUaiiedX3jRud41Zuu1ZcRUZZ8zY9SuJJgXHuiQ", NetID: "some-host:9081"},
		{PubKey: "8mcS4hUaiiedX3jRud41Zuu1ZcRUZZ8zY9SuJJgXHuiR", NetID: "some-host:9082"},
	}

	adm.GET(routes.PeeringTrustedList(), handlePeeringTrustedList).
		AddResponse(http.StatusOK, "A list of trusted peers.", listExample, nil).
		SetSummary("Get a list of trusted peers.")

	adm.GET(routes.PeeringTrustedGet(":pubKey"), handlePeeringTrustedGet).
		AddParamPath(listExample[0].PubKey, "pubKey", "Public key of the trusted peer (base58).").
		AddResponse(http.StatusOK, "Trusted peer info.", listExample[0], nil).
		SetSummary("Get details on a particular trusted peer.")

	adm.PUT(routes.PeeringTrustedPut(":pubKey"), handlePeeringTrustedPut).
		AddParamPath(listExample[0].PubKey, "pubKey", "Public key of the trusted peer (base58).").
		AddParamBody(listExample[0], "PeeringTrustedNode", "Info of the peer to trust.", true).
		AddResponse(http.StatusOK, "Trusted peer info.", listExample[0], nil).
		SetSummary("Trust the specified peer, the pub key is passed via the path.")

	adm.POST(routes.PeeringTrustedPost(), handlePeeringTrustedPost).
		AddParamBody(listExample[0], "PeeringTrustedNode", "Info of the peer to trust.", true).
		AddResponse(http.StatusOK, "Trusted peer info.", listExample[0], nil).
		SetSummary("Trust the specified peer.")

	adm.DELETE(routes.PeeringTrustedDelete(":pubKey"), handlePeeringTrustedDelete).
		AddParamPath(listExample[0].PubKey, "pubKey", "Public key of the trusted peer (base58).").
		SetSummary("Distrust the specified peer.")
}

func handlePeeringTrustedList(c echo.Context) error {
	response := []*model.PeeringTrustedNode{ // TODO: Implement.
		{PubKey: "8mcS4hUaiiedX3jRud41Zuu1ZcRUZZ8zY9SuJJgXHuiQ", NetID: "some-host:9081"},
		{PubKey: "8mcS4hUaiiedX3jRud41Zuu1ZcRUZZ8zY9SuJJgXHuiR", NetID: "some-host:9082"},
	}
	return c.JSON(http.StatusOK, response)
}

func handlePeeringTrustedPut(c echo.Context) error {
	response := &model.PeeringTrustedNode{ // TODO: Implement.
		PubKey: "8mcS4hUaiiedX3jRud41Zuu1ZcRUZZ8zY9SuJJgXHuiQ",
		NetID:  "some-host:9081",
	}
	return c.JSON(http.StatusOK, response)
}

func handlePeeringTrustedPost(c echo.Context) error {
	response := &model.PeeringTrustedNode{ // TODO: Implement.
		PubKey: "8mcS4hUaiiedX3jRud41Zuu1ZcRUZZ8zY9SuJJgXHuiQ",
		NetID:  "some-host:9081",
	}
	return c.JSON(http.StatusOK, response)
}

func handlePeeringTrustedGet(c echo.Context) error {
	response := &model.PeeringTrustedNode{ // TODO: Implement.
		PubKey: "8mcS4hUaiiedX3jRud41Zuu1ZcRUZZ8zY9SuJJgXHuiQ",
		NetID:  "some-host:9081",
	}
	return c.JSON(http.StatusOK, response)
}

func handlePeeringTrustedDelete(c echo.Context) error {
	// TODO: Implement.
	return c.NoContent(http.StatusOK)
}
