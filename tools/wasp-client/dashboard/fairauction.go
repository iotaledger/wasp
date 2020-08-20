package dashboard

import (
	"html/template"
	"net/http"

	"github.com/iotaledger/goshimmer/dapps/valuetransfers/packages/address"
	"github.com/iotaledger/wasp/packages/vm/examples/fairauction/faclient"
	"github.com/iotaledger/wasp/tools/wasp-client/config/fa"
	"github.com/labstack/echo"
)

func handleFA(c echo.Context) error {
	scAddress := fa.Config.Address()
	status, err := fa.Client().FetchStatus()
	if err != nil {
		return err
	}
	return c.Render(http.StatusOK, "fairauction", &FATemplateParams{
		BaseTemplateParams: baseParams(c, "fairauction"),
		SCAddress:          *scAddress,
		Status:             status,
	})
}

type FATemplateParams struct {
	BaseTemplateParams
	SCAddress address.Address
	Status    *faclient.Status
}

func initFATemplate() *template.Template {
	t := template.Must(template.New("").Parse(tplBase))
	t = template.Must(t.Parse(tplWs))
	t = template.Must(t.Parse(tplInstallConfig))
	t = template.Must(t.Parse(tplFairAuction))
	return t
}

const tplFairAuction = `
{{define "title"}}FairAuction{{end}}

{{define "body"}}
	<p>SC address: <code>{{.SCAddress}}</code></p>
	<p>Balance: <code>{{.Status.SCBalance}} IOTAs</code></p>

	<hr/>
	<p>Status fetched at: <code>{{.Status.FetchedAt}}</code></p>
	<div>
		<h2>CLI usage</h2>
		{{template "install-config" .}}
		<details>
			<summary>3. Mint a new color</summary>
			<p><code>wasp-client wallet mint <i>amount-tokens</i></code>
			(e.g.: <code>wasp-client wallet mint 1</code>)</p>
		</details>
		<details>
			<summary>4. Start an auction</summary>
			<p><code>wasp-client fa start-auction <i>description</i> <i>color</i> <i>amount-tokens</i> <i>minimum-bid</i> <i>duration</i></code>
			(e.g.: <code>wasp-client fa start-auction gHw2r... 1 100 10</code>)</p>
		</details>
		<details>
			<summary>5. Place a bid</summary>
			<p><code>wasp-client fa place-bid <i>color</i> <i>amount-iotas</i></code>
			(e.g.: <code>wasp-client fa place-bid gHw2r... 110</code>)</p>
		</details>
	</div>
	{{template "ws" .}}
{{end}}
`
