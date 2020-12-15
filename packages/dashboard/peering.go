package dashboard

import (
	"net/http"

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
			Status:             peering.GetStatus(),
		})
	})
}

type PeeringTemplateParams struct {
	BaseTemplateParams
	Status *peering.Status
}

const tplPeering = `
{{define "title"}}Peering{{end}}

{{define "body"}}
	<h2>Peers</h2>
	<table>
		<thead>
			<tr>
				<th>Location</th>
				<th>Type</th>
				<th>Status</th>
				<th>#Users</th>
			</tr>
		</thead>
		<tbody>
		{{range $_, $peer := .Status.Peers}}
			<tr>
				<td><code>{{$peer.RemoteLocation}}</code></td>
				<td>{{if $peer.IsInbound}}inbound{{else}}outbound{{end}}</td>
				<td>{{if $peer.IsAlive}}up{{else}}down{{end}}</td>
				<td>{{$peer.NumUsers}}</td>
			</tr>
		{{end}}
		</tbody>
	</table>
{{end}}
`
