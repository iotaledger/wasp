package dashboard

import (
	"fmt"
	"html/template"
	"net/http"

	"github.com/iotaledger/goshimmer/dapps/valuetransfers/packages/balance"
	"github.com/iotaledger/wasp/packages/util"
	"github.com/iotaledger/wasp/packages/vm/examples/tokenregistry"
	"github.com/iotaledger/wasp/packages/vm/examples/tokenregistry/trclient"
	"github.com/iotaledger/wasp/tools/wwallet/sc"
	"github.com/iotaledger/wasp/tools/wwallet/sc/tr"
	"github.com/labstack/echo"
)

func handleTR(c echo.Context) error {
	status, err := tr.Client().FetchStatus()
	if err != nil {
		return err
	}
	return c.Render(http.StatusOK, "tokenregistry", &TRTemplateParams{
		BaseTemplateParams: baseParams(c, "tokenregistry"),
		SC:                 tr.Config,
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
	return c.Render(http.StatusOK, "tokenregistry", &TRTemplateParams{
		BaseTemplateParams: baseParams(c, "tokenregistry"),
		SC:                 tr.Config,
		Color:              &color,
		QueryResult:        tm,
	})
}

type TRTemplateParams struct {
	BaseTemplateParams
	SC          *sc.Config
	Status      *trclient.Status
	Color       *balance.Color
	QueryResult *tokenregistry.TokenMetadata
}

func initTRTemplate() *template.Template {
	return makeTemplate(tplWs, tplSCInfo, tplInstallConfig, tplTokenRegistry)
}

const tplTokenRegistry = `
{{define "title"}}TokenRegistry{{end}}

{{define "tmdetails"}}
	<p>Supply: <code>{{.Supply}}</code></p>
	<p>Minted by: <code>{{.MintedBy}}</code></p>
	<p>Owner: <code>{{.Owner}}</code></p>
	<p>Created: <code>{{formatTimestamp .Created}}</code></p>
	<p>Updated: <code>{{formatTimestamp .Updated}}</code></p>
	<p>UserDefined: <code>{{.UserDefined}}</code></p>
{{end}}

{{define "body"}}
	<h2>TokenRegistry</h2>

	{{if .Status}}
		{{template "sc-info" .}}

		<div>
			<h3>Registry</h3>
			<div>
				{{range $color, $tm := .Status.Registry}}
					<details>
						<summary>{{$tm.Description}}</summary>
						<p>Color: <code>{{$color}}</code></p>
						{{template "tmdetails" $tm}}
					</details>
				{{end}}
			</div>
		</div>
		<hr/>
		<p>Status fetched at: <code>{{.Status.FetchedAt}}</code></p>

		{{template "ws" .}}
	{{end}}

	{{if .Color}}
		{{if .QueryResult}}
			<h3>{{.QueryResult.Description}}</h3>
			<p>Color: <code>{{.Color}}</code></p>
			{{template "tmdetails" .QueryResult}}
		{{else}}
			<p>Registry contains no entry for color <code>{{.Color}}</code></p>
		{{end}}
	{{end}}

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
