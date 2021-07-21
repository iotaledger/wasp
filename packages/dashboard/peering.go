// Copyright 2020 IOTA Stiftung
// SPDX-License-Identifier: Apache-2.0

package dashboard

import (
	_ "embed"
	"net/http"

	"github.com/iotaledger/wasp/packages/peering"
	"github.com/labstack/echo/v4"
)

//go:embed templates/peering.tmpl
var tplPeering string

func (d *Dashboard) peeringInit(e *echo.Echo, r renderer) Tab {
	route := e.GET("/peering", d.handlePeering)
	route.Name = "peering"

	r[route.Path] = d.makeTemplate(e, tplPeering)

	return Tab{
		Path:  route.Path,
		Title: "Peering",
		Href:  route.Path,
	}
}

func (d *Dashboard) handlePeering(c echo.Context) error {
	return c.Render(http.StatusOK, c.Path(), &PeeringTemplateParams{
		BaseTemplateParams:    d.BaseParams(c),
		NetworkProvider:       d.wasp.NetworkProvider(),
		TrustedNetworkManager: d.wasp.TrustedNetworkManager(),
	})
}

type PeeringTemplateParams struct {
	BaseTemplateParams
	NetworkProvider       peering.NetworkProvider
	TrustedNetworkManager peering.TrustedNetworkManager
}
