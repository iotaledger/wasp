package dashboard

import (
	"html/template"
	"net/http"
	"time"

	"github.com/iotaledger/goshimmer/dapps/valuetransfers/packages/address"
	"github.com/iotaledger/wasp/packages/vm/examples/fairroulette"
	"github.com/iotaledger/wasp/tools/wasp-client/config/fr"
	"github.com/labstack/echo"
)

func handleFR(c echo.Context) error {
	scAddress := fr.Config.Address()
	status, err := fr.Client().FetchStatus()
	if err != nil {
		return err
	}
	return c.Render(http.StatusOK, "fairroulette", &FRTemplateParams{
		BaseTemplateParams: baseParams(c, "fairroulette"),
		SCAddress:          *scAddress,
		Status:             status,
	})
}

type FRTemplateParams struct {
	BaseTemplateParams
	SCAddress address.Address
	Status    *fairroulette.Status
}

func (p FRTemplateParams) FormatNextPlayTime() string {
	return p.Status.NextPlayTimestamp.Format(time.RFC3339)
}

func initFRTemplate() *template.Template {
	t := template.Must(template.New("").Parse(tplBase))
	t = template.Must(t.Parse(tplWs))
	t = template.Must(t.Parse(tplFairRoulette))
	return t
}

const tplFairRoulette = `
{{define "title"}}FairRoulette{{end}}

{{define "body"}}
	<p>SC address: <code>{{.SCAddress}}</code></p>
	<p>Balance: <code>{{.Status.SCBalance}} IOTAs</code></p>
	<div>
		<h2>Next play</h2>
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
		<h2>Stats</h2>
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
		<h2>CLI usage</h2>
		<details>
			<summary>1. Install</summary>
			<p>Grab the latest <code>wasp-client</code> binary from the
			<a href="https://github.com/iotaledger/wasp/releases">Releases</a> page.</p>
			<p>-- OR --</p>
			<p>Build from source:</p>
	<pre>$ git clone --branch develop https://github.com/iotaledger/wasp.git
	$ cd wasp
	$ go install ./tools/wallet
	</pre>
		</details>
		<details>
			<summary>2. Configure</summary>
	<pre>$ wasp-client set goshimmer.api {{.Host}}:8080
	$ wasp-client set wasp.api {{.Host}}:9090
	$ wasp-client fr set address {{.SCAddress}}</pre>
			<p>Initialize a wallet: <code>wasp-client wallet init</code></p>
			<p>Get some funds: <code>wasp-client wallet request-funds</code></p>
		</details>
		<details>
			<summary>3. Place bets</summary>
			<p><code>wasp-client fr bet <i>color</i> <i>amount</i></code>
			(e.g.: <code>wasp-client fr bet 1 100</code>)</p>
			<p>Then refresh this page to see the results.</p>
		</details>
	</div>

	<script>
		const nextPlayAt = new Date({{.FormatNextPlayTime}});

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
