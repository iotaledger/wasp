// Copyright 2020 IOTA Stiftung
// SPDX-License-Identifier: Apache-2.0

package dashboard

import (
	"net/http"

	peering_pkg "github.com/iotaledger/wasp/packages/peering"
	"github.com/iotaledger/wasp/plugins/peering"
	"github.com/labstack/echo/v4"
)

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

const tplPeering = `
{{define "title"}}Peering{{end}}

{{define "body"}}
<div class="card fluid">
	<h2 class="section">Peering</h2>
	<dl>
		<dt>Node network ID</dt><dd><tt>{{.MyNetworkId}}</tt></dd>
	</dl>
</div>
<div class="card fluid">
	<h3 class="section">Peers</h3>
	<table>
		<thead>
			<tr>
				<th>NetID</th>
				<th>Type</th>
				<th>Status</th>
				<th>#Users</th>
			</tr>
		</thead>
		<tbody>
		{{range $_, $ps := .NetworkProvider.PeerStatus}}
			<tr>
				<td data-label="NetID"><tt>{{$ps.NetID}}</tt></td>
				<td data-label="Type">{{if $ps.IsInbound}}inbound{{else}}outbound{{end}}</td>
				<td data-label="Status">{{if $ps.IsAlive}}up{{else}}down{{end}}</td>
				<td data-label="#Users">{{$ps.NumUsers}}</td>
			</tr>
		{{end}}
		</tbody>
	</table>
</div>
{{end}}
`
