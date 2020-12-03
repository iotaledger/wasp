package trdashboard

import (
	"fmt"
	"net/http"

	"github.com/iotaledger/goshimmer/dapps/valuetransfers/packages/balance"
	"github.com/iotaledger/wasp/packages/util"
	"github.com/iotaledger/wasp/packages/vm/examples/tokenregistry"
	"github.com/iotaledger/wasp/packages/vm/examples/tokenregistry/trclient"
	"github.com/iotaledger/wasp/tools/wasp-cli/dashboard"
	"github.com/iotaledger/wasp/tools/wasp-cli/sc"
	"github.com/iotaledger/wasp/tools/wasp-cli/sc/tr"
	"github.com/labstack/echo"
)

type trdashboard struct{}

func Dashboard() dashboard.SCDashboard {
	return &trdashboard{}
}

func (d *trdashboard) Config() *sc.Config {
	return tr.Config
}

func (d *trdashboard) AddEndpoints(e *echo.Echo) {
	e.GET(tr.Config.Href(), handleTR)
	e.GET(tr.Config.Href()+"/:color", handleTRQuery)
}

func (d *trdashboard) AddTemplates(r dashboard.Renderer) {
	r[tr.Config.ShortName] = dashboard.MakeTemplate(
		dashboard.TplWs,
		dashboard.TplSCInfo,
		dashboard.TplInstallConfig,
		tplTokenRegistry,
	)
}

func handleTR(c echo.Context) error {
	status, err := tr.Client().FetchStatus(true)
	if err != nil {
		return err
	}
	return c.Render(http.StatusOK, tr.Config.ShortName, &TRTemplateParams{
		BaseTemplateParams: dashboard.BaseParams(c, tr.Config.Href()),
		Config:             tr.Config,
		Status:             status,
	})
}

func handleTRQuery(c echo.Context) error {
	color, err := util.ColorFromString(c.Param("color"))
	if err != nil {
		return err
	}

	tm, err := tr.Client().Query(&color)
	if err != nil {
		return fmt.Errorf("query error: %v", err)
	}
	return c.Render(http.StatusOK, tr.Config.ShortName, &TRTemplateParams{
		BaseTemplateParams: dashboard.BaseParams(c, tr.Config.Href()),
		Config:             tr.Config,
		Color:              &color,
		QueryResult:        tm,
	})
}

type TRTemplateParams struct {
	dashboard.BaseTemplateParams
	Config      *sc.Config
	Status      *trclient.Status
	Color       *balance.Color
	QueryResult *tokenregistry.TokenMetadata
}

const tplTokenRegistry = `
{{define "title"}}{{.Config.Name}}{{end}}

{{define "tmdetails"}}
	<p>Description: {{trim .Description}}</p>
	<p>Supply: <code>{{.Supply}}</code></p>
	<p>Minted by: {{template "address" .MintedBy}}</p>
	<p>Minted when: <code>{{formatTimestamp .Created}}</code></p>
	<p>Metadata owner: <code>{{.Owner}}</code></p>
	<p>Additional metadata: <code>{{.UserDefined}}</code></p>
	<p>Metadata updated when: <code>{{formatTimestamp .Updated}}</code></p>
{{end}}

{{define "body"}}
	<h2>{{.Config.Name}}</h1>

	{{if .Status}}
		{{template "sc-info" .}}

		<div>
			<h3>Registry (chronologically sorted)</h3>
			<div>
				{{range $_, $tm := .Status.RegistrySortedByMintTimeDesc}}
					<details>
						<summary>{{$tm.Supply}} token(s) of color <code>{{$tm.Color}}</code></summary>
						{{template "tmdetails" $tm}}
					</details>
				{{end}}
			</div>
		</div>
		<hr/>
		<p>Status fetched at: <code>{{formatTimestamp .Status.FetchedAt}}</code></p>

		{{template "ws" .}}
	{{end}}

	{{if .Color}}
		{{if .QueryResult}}
			<h3>{{trim .QueryResult.Description}}</h3>
			<p>Color: <code>{{.Color}}</code></p>
			{{template "tmdetails" .QueryResult}}
		{{else}}
			<p>Registry contains no entry for color <code>{{.Color}}</code></p>
		{{end}}
	{{end}}

	<hr/>
	<div>
		<h3>CLI usage</h3>
		{{template "install-config" .}}
		<details>
			<summary>3. Mint a new color</summary>
			<p><code>{{waspClientCmd}} tr mint <i>description</i> <i>amount-tokens</i></code>
			<br/>(e.g.: <code>{{waspClientCmd}} tr mint "My first coin" 1</code>)</p>
		</details>
	</div>
{{end}}
`
