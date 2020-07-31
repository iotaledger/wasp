package dashboard

import (
	"html/template"
	"io"
	"time"

	"github.com/iotaledger/goshimmer/dapps/valuetransfers/packages/address"
	"github.com/iotaledger/wasp/tools/fairroulette/client"
	"github.com/labstack/echo"
)

type IndexTemplateParams struct {
	Host      string
	SCAddress address.Address
	Status    *client.Status
}

func (p IndexTemplateParams) FormatNextPlayTime() string {
	return p.Status.NextPlayTimestamp.Format(time.RFC3339)
}

type Template struct {
	templates *template.Template
}

func (t *Template) Render(w io.Writer, name string, data interface{}, c echo.Context) error {
	return t.templates.ExecuteTemplate(w, name, data)
}

var renderer = &Template{
	templates: template.Must(template.New("index").Parse(`
<!doctype html>
<html lang="en">
  <head>
  	<meta charset="utf-8" />
	<meta http-equiv="x-ua-compatible" content="ie=edge" />
	<meta name="viewport" content="width=device-width, initial-scale=1" />

	<title>FairRoulette dashboard</title>

	<link rel="stylesheet" href="https://cdn.jsdelivr.net/npm/@xz/fonts@1/serve/inter.css">
	<link rel="stylesheet" href="https://cdn.jsdelivr.net/npm/@exampledev/new.css@1.1.2/new.min.css">
  </head>

  <body>
  	<style>
		details {background: #EEF9FF}
	</style>
	<header>
		<h1>FairRoulette</h1>
	</header>
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
			<p>Grab the latest <code>fairroulette</code> binary from the
			<a href="https://github.com/iotaledger/wasp/releases">Releases</a> page.</p>
			<p>-- OR --</p>
			<p>Build from source:</p>
<pre>$ git clone --branch develop https://github.com/iotaledger/wasp.git
$ cd wasp
$ go install ./tools/fairroulette
</pre>
		</details>
		<details>
			<summary>2. Configure</summary>
<pre>$ fairroulette set goshimmer.api {{.Host}}:8080
$ fairroulette set wasp.api {{.Host}}:9090
$ fairroulette set address {{.SCAddress}}</pre>
			<p>Initialize a wallet: <code>fairroulette wallet init</code></p>
			<p>Get some funds: <code>fairroulette wallet request-funds</code></p>
		</details>
		<details>
			<summary>3. Place bets</summary>
			<p><code>fairroulette bet <i>color</i> <i>amount</i></code>
			(e.g.: <code>fairroulette bet 1 100</code>)</p>
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

	<script>
		const url = 'ws://' +  location.host + '/ws';
		console.log('opening WebSocket to ' + url);
		const ws = new WebSocket(url);

		ws.addEventListener('error', function (event) {
			console.error('WebSocket error!', event);
		});

		const connectedAt = new Date();
		ws.addEventListener('message', function (event) {
			console.log('Message from server: ', event.data);
			ws.close();
			if (new Date() - connectedAt > 5000) {
				location.reload();
			} else {
				setTimeout(() => location.reload(), 5000);
			}
		});
	</script>
  </body>
</html>
`)),
}
