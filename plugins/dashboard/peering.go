package dashboard

import (
	"net/http"

	"github.com/iotaledger/wasp/plugins/peering"
	"github.com/labstack/echo"
)

type peeringNavPage struct{}

func initPeering() NavPage {
	return &peeringNavPage{}
}

const peeringRoute = "/"
const peeringTplName = "peering"

func (n *peeringNavPage) Title() string { return "Peering" }
func (n *peeringNavPage) Href() string  { return peeringRoute }

func (n *peeringNavPage) AddTemplates(renderer Renderer) {
	renderer[peeringTplName] = MakeTemplate(tplPeering)
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
	<p>Node network ID: <code>{{.Status.MyNetworkId}}</code></p>
	<h2>Peers</h2>
	<ul>
	{{range $_, $peer := .Status.Peers}}
		<li>
			<p>Remote location: <code>{{$peer.RemoteLocation}}</code></p>
			<p>Is inbound?: <code>{{$peer.IsInbound}}</code></p>
			<p>Is alive?: <code>{{$peer.IsAlive}}</code></p>
		</li>
	{{else}}
		<p>(empty list)</p>
	{{end}}
	</ul>
{{end}}
`
