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
	<h2>Peers</h2>
	<div>
	{{range $_, $peer := .Status.Peers}}
		<details>
			<summary><code>{{$peer.RemoteLocation}}</code></summary>
			<p>Is inbound?: <code>{{$peer.IsInbound}}</code></p>
			<p>Is alive?: <code>{{$peer.IsAlive}}</code></p>
		</details>
	{{else}}
		<p>(empty list)</p>
	{{end}}
	</div>
{{end}}
`
