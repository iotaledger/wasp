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
<div class="container">
<div class="row">
<div class="col-sm">
	<h2>Peers</h2>
	<p>Node network ID: <tt>{{.MyNetworkId}}</tt></p>
	<table style="max-width: 50em">
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
</div>
</div>
{{end}}
`
