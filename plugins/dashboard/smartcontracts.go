package dashboard

import (
	"net/http"

	"github.com/iotaledger/wasp/packages/registry"
	"github.com/labstack/echo"
)

type scNavPage struct{}

func initSc() NavPage {
	return &scNavPage{}
}

const scRoute = "/smart-contracts"
const scTplName = "sc"

func (n *scNavPage) Title() string { return "Smart Contracts" }
func (n *scNavPage) Href() string  { return scRoute }

func (n *scNavPage) AddTemplates(renderer Renderer) {
	renderer[scTplName] = MakeTemplate(tplSc)
}

func (n *scNavPage) AddEndpoints(e *echo.Echo) {
	e.GET(scRoute, func(c echo.Context) error {
		brs, err := registry.GetBootupRecords()
		if err != nil {
			return err
		}
		return c.Render(http.StatusOK, scTplName, &ScTemplateParams{
			BaseTemplateParams: BaseParams(c, scRoute),
			BootupRecords:      brs,
		})
	})
}

type ScTemplateParams struct {
	BaseTemplateParams
	BootupRecords []*registry.BootupData
}

const tplSc = `
{{define "title"}}Smart Contracts{{end}}

{{define "body"}}
	<h2>Smart Contracts</h2>
	<div>
	{{range $_, $r := .BootupRecords}}
		<details>
			<summary><code>{{$r.Address}}</code></summary>
			<p>Owner address:   <code>{{$r.OwnerAddress}}</code></p>
			<p>Color:           <code>{{$r.Color}}</code></p>
			<p>Committee Nodes: <code>{{$r.CommitteeNodes}}</code></p>
			<p>Access Nodes:    <code>{{$r.AccessNodes}}</code></p>
			<p>Active:          <code>{{$r.Active}}</code></p>
		</details>
	{{else}}
		<p>(empty list)</p>
	{{end}}
	</div>
{{end}}
`
