// Copyright 2020 IOTA Stiftung
// SPDX-License-Identifier: Apache-2.0

package dashboard

import (
	_ "embed"
	"net/http"

	peering_pkg "github.com/iotaledger/wasp/packages/peering"
	"github.com/iotaledger/wasp/plugins/peering"
	"github.com/labstack/echo/v4"
)

//go:embed templates/peering.tmpl
var tplPeering string

func peeringInit(e *echo.Echo, r renderer) Tab {
	route := e.GET("/peering", handlePeering)
	route.Name = "peering"

	r[route.Path] = makeTemplate(e, tplPeering)

	return Tab{
		Path:  route.Path,
		Title: "Peering",
		Href:  route.Path,
	}
}

func handlePeering(c echo.Context) error {
	return c.Render(http.StatusOK, c.Path(), &PeeringTemplateParams{
		BaseTemplateParams: BaseParams(c),
		NetworkProvider:    peering.DefaultNetworkProvider(),
	})
}

type PeeringTemplateParams struct {
	BaseTemplateParams
	NetworkProvider peering_pkg.NetworkProvider
}
