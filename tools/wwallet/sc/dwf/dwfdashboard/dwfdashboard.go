package dwfdashboard

import (
	"net/http"

	"github.com/iotaledger/wasp/packages/vm/examples/donatewithfeedback/dwfclient"
	"github.com/iotaledger/wasp/tools/wwallet/dashboard"
	"github.com/iotaledger/wasp/tools/wwallet/sc"
	"github.com/iotaledger/wasp/tools/wwallet/sc/dwf"
	"github.com/labstack/echo"
)

type dwfdashboard struct{}

func Dashboard() dashboard.SCDashboard {
	return &dwfdashboard{}
}

func (d *dwfdashboard) Config() *sc.Config {
	return dwf.Config
}

func (d *dwfdashboard) AddEndpoints(e *echo.Echo) {
	e.GET(dwf.Config.Href(), handleDwf)
}

func (d *dwfdashboard) AddTemplates(r dashboard.Renderer) {
	r[dwf.Config.ShortName] = dashboard.MakeTemplate(
		dashboard.TplWs,
		dashboard.TplSCInfo,
		dashboard.TplInstallConfig,
		tplDwf,
	)
}

func handleDwf(c echo.Context) error {
	status, err := dwf.Client().FetchStatus()
	if err != nil {
		return err
	}
	return c.Render(http.StatusOK, dwf.Config.ShortName, &DwfTemplateParams{
		BaseTemplateParams: dashboard.BaseParams(c, dwf.Config.Href()),
		Config:             dwf.Config,
		Status:             status,
	})
}

type DwfTemplateParams struct {
	dashboard.BaseTemplateParams
	Config *sc.Config
	Status *dwfclient.Status
}

const tplDwf = `
{{define "title"}}{{.Config.Name}}{{end}}

{{define "body"}}
	<h2>{{.Config.Name}}</h2>

	{{template "sc-info" .}}

	<div>
		<h3>Donations</h3>
		<p>Number of donations: <code>{{.Status.NumRecords}}</code></p>
		<p>Total received: <code>{{.Status.TotalDonations}} IOTAs</code></p>
		<p>Largest donation so far: <code>{{.Status.MaxDonation}} IOTAs</code></p>
		<div>
			<h4>Log (latest first)</h4>
			{{range $i, $di := .Status.LastRecordsDesc}}
				<details>
					<summary>{{$di.Seq}}: {{$di.Feedback}}</summary>
					<p>Sender: <code>{{$di.Sender}}</code></p>
					<p>Amount: <code>{{$di.Amount}} IOTAs</code></p>
					<p>When: <code>{{formatTimestamp $di.When}}</code></p>
		            <p>Request Id: <code>{{$di.Id}}</code></p>
				</details>
			{{end}}
		</div>
	</div>
	<hr/>
	<p>Status fetched at: <code>{{formatTimestamp .Status.FetchedAt}}</code></p>

	{{template "ws" .}}

	<div>
		<h3>CLI usage</h3>
		{{template "install-config" .}}
		<details>
			<summary>3. Donate</summary>
			<p><code>{{waspClientCmd}} dwf donate <i>amount-iotas</i> <i>feedback</i></code>
			<br/>(e.g.: <code>{{waspClientCmd}} dwf donate 100 "Nice app :)"</code>)</p>
			<summary>3. Withdraw (if you are owner of Config)</summary>
			<p><code>{{waspClientCmd}} dwf withdraw <i>amount-iotas</i></code>
			<br/>(e.g.: <code>{{waspClientCmd}} dwf withdraw 100</code>)</p>
		</details>
	</div>
{{end}}
`
