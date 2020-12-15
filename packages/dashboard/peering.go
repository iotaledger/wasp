// Copyright 2020 IOTA Stiftung
// SPDX-License-Identifier: Apache-2.0

package dashboard

import (
	"net/http"

	peering_pkg "github.com/iotaledger/wasp/packages/peering"
	"github.com/iotaledger/wasp/plugins/peering"
	"github.com/labstack/echo"
)

type peeringNavPage struct{}

const peeringRoute = "/peering"
const peeringTplName = "peering"

func (n *peeringNavPage) Title() string { return "Peering" }
func (n *peeringNavPage) Href() string  { return peeringRoute }

func (n *peeringNavPage) AddTemplates(r renderer) {
	r[peeringTplName] = MakeTemplate(tplPeering)
}

func (n *peeringNavPage) AddEndpoints(e *echo.Echo) {
	e.GET(peeringRoute, func(c echo.Context) error {
		return c.Render(http.StatusOK, peeringTplName, &PeeringTemplateParams{
			BaseTemplateParams: BaseParams(c, peeringRoute),
			NetworkProvider:    peering.DefaultNetworkProvider(),
		})
	})
}

// PeeringTemplateParams holts template params.
type PeeringTemplateParams struct {
	BaseTemplateParams
	NetworkProvider peering_pkg.NetworkProvider
}

const tplPeering = `
{{define "title"}}Peering{{end}}

{{define "body"}}
	<h2>Peers</h2>
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
				<td><code>{{$ps.NetID}}</code></td>
				<td>{{if $ps.IsInbound}}inbound{{else}}outbound{{end}}</td>
				<td>{{if $ps.IsAlive}}up{{else}}down{{end}}</td>
				<td>{{$ps.NumUsers}}</td>
			</tr>
		{{end}}
		</tbody>
	</table>
{{end}}
`
