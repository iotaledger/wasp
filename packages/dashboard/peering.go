// Copyright 2020 IOTA Stiftung
// SPDX-License-Identifier: Apache-2.0

package dashboard

import (
	_ "embed"
	"net/http"

	"github.com/iotaledger/wasp/packages/cryptolib"
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
	p, err := d.wasp.PeeringStats()
	if err != nil {
		return err
	}
	return c.Render(http.StatusOK, c.Path(), &PeeringTemplateParams{
		BaseTemplateParams: d.BaseParams(c),
		PeeringStats:       p,
	})
}

type PeeringTemplateParams struct {
	BaseTemplateParams
	*PeeringStats
}

type PeeringStats struct {
	Peers        []Peer
	TrustedPeers []TrustedPeer
}

type Peer struct {
	NumUsers int
	NetID    string
	IsAlive  bool
}

type TrustedPeer struct {
	NetID  string
	PubKey cryptolib.PublicKey
}
