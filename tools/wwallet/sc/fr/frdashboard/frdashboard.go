package frdashboard

import (
	"net/http"

	"github.com/iotaledger/wasp/packages/vm/examples/fairroulette/frclient"
	"github.com/iotaledger/wasp/tools/wwallet/dashboard"
	"github.com/iotaledger/wasp/tools/wwallet/sc"
	"github.com/iotaledger/wasp/tools/wwallet/sc/fr"
	"github.com/labstack/echo"
)

type frdashboard struct{}

func Dashboard() dashboard.SCDashboard {
	return &frdashboard{}
}

func (d *frdashboard) Config() *sc.Config {
	return fr.Config
}

func (d *frdashboard) AddEndpoints(e *echo.Echo) {
	e.GET(fr.Config.Href(), handleFR)
}

func (d *frdashboard) AddTemplates(r dashboard.Renderer) {
	r[fr.Config.ShortName] = dashboard.MakeTemplate(
		dashboard.TplWs,
		dashboard.TplSCInfo,
		dashboard.TplInstallConfig,
		tplFairRoulette,
	)
}

func handleFR(c echo.Context) error {
	status, err := fr.Client().FetchStatus()
	if err != nil {
		return err
	}
	return c.Render(http.StatusOK, fr.Config.ShortName, &FRTemplateParams{
		BaseTemplateParams: dashboard.BaseParams(c, fr.Config.Href()),
		Config:             fr.Config,
		Status:             status,
	})
}

type FRTemplateParams struct {
	dashboard.BaseTemplateParams
	Config *sc.Config
	Status *frclient.Status
}

const tplFairRoulette = `
{{define "title"}}{{.Config.Name}}{{end}}

{{define "body"}}
	<h2>{{.Config.Name}}</h1>
	{{template "sc-info" .}}

	<div>
		<h3>Next play</h3>
		<p>Next play: <code id="nextPlayIn"></code></p>
		<p>Play period: <code>{{.Status.PlayPeriodSeconds}}s</code></p>
		<div>
			<p>Bets: <code>{{.Status.CurrentBetsAmount}}</code></p>
			{{if lt (len .Status.CurrentBets) .Status.CurrentBetsAmount}}
				<p>(Showing first {{(len .Status.CurrentBets)}})</p>
			{{end}}
			<ul>
			{{range .Status.CurrentBets}}
				<li>Player <code>{{.Player}}</code> bets <code>{{.Sum}} IOTAs</code> on <code>color {{.Color}}</code></li>
			{{end}}
			</ul>
		</div>
	</div>

	<div>
		<h3>Stats</h3>
		<p>Last winning color: <code>{{.Status.LastWinningColor}}</code></p>
		<div>
			<p>Color stats:</p>
			<ul>
				{{range $c, $w := .Status.WinsPerColor}}
					<li><b>Color {{$c}}</b> won <code>{{$w}} times</code> so far</li>
				{{end}}
			</ul>
		</div>
		<div>
			<p>Player stats:</p>
			<ul>
				{{range $p, $stats := .Status.PlayerStats}}
					<li>Player <code>{{$p}}</code>: Bets: <code>{{$stats.Bets}}</code> - Wins: <code>{{$stats.Wins}}</code></li>
				{{end}}
			</ul>
		</div>
	</div>
	<hr/>
	<p>Status fetched at: <code>{{.Status.FetchedAt}}</code></p>
	<div>
		<h3>CLI usage</h3>
		{{template "install-config" .}}
		<details>
			<summary>3. Place bets</summary>
			<p><code>{{waspClientCmd}} fr bet <i>color</i> <i>amount</i></code>
			<br/>(e.g.: <code>{{waspClientCmd}} fr bet 1 100</code>)</p>
			<p>Then refresh this page to see the results.</p>
		</details>
	</div>

	<script>
		const nextPlayAt = new Date({{formatTimestamp .Status.NextPlayTimestamp}});

		const nextPlayIn = document.getElementById("nextPlayIn");

		function update() {
			const diff = nextPlayAt - new Date();
			if (diff > 0) {
				var date = new Date(0);
				date.setSeconds(diff / 1000);
				nextPlayIn.innerText = date.toISOString().substr(11, 8);
			} else {
				nextPlayIn.innerText = "not scheduled";
			}
		}

		update()
		setInterval(update, 1000);
	</script>

	{{template "ws" .}}
{{end}}
`
