package dashboard

import (
	"fmt"
	"html/template"
	"io"
	"net"
	"net/http"
	"os"
	"strings"
	"time"

	"github.com/iotaledger/goshimmer/dapps/valuetransfers/packages/address"
	"github.com/iotaledger/wasp/tools/fairroulette/client"
	"github.com/iotaledger/wasp/tools/fairroulette/config"
	"github.com/labstack/echo"
	"github.com/labstack/echo/middleware"
)

func check(err error) {
	if err != nil {
		panic(err)
	}
}

func Cmd(args []string) {
	listenAddr := ":10000"
	if len(args) > 0 {
		if len(args) != 1 {
			fmt.Printf("Usage: %s dashboard [listen-address]\n", os.Args[0])
			os.Exit(1)
		}
		listenAddr = args[0]
	}

	e := echo.New()
	e.Use(middleware.Logger())
	e.Use(middleware.Recover())
	e.Debug = true
	e.HideBanner = true
	e.Renderer = renderer

	e.GET("/", func(c echo.Context) error {
		var players []string
		if len(c.QueryParam("players")) > 0 {
			players = strings.Split(c.QueryParam("players"), ",")
		}

		status, err := client.FetchStatus(players)
		if err != nil {
			return err
		}
		host, _, err := net.SplitHostPort(c.Request().Host)
		if err != nil {
			return err
		}
		return c.Render(http.StatusOK, "index", &IndexTemplateParams{
			Host:      host,
			SCAddress: config.GetSCAddress(),
			Now:       time.Now().UTC(),
			Status:    status,
		})
	})

	fmt.Printf("Serving dashboard on %s\n", listenAddr)
	e.Logger.Fatal(e.Start(listenAddr))
}

type IndexTemplateParams struct {
	Host       string
	SCAddress  address.Address
	Now        time.Time
	NextPlayIn string
	Status     *client.Status
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

	<link rel="stylesheet" href="https://fonts.xz.style/serve/inter.css">
	<link rel="stylesheet" href="https://cdn.jsdelivr.net/npm/@exampledev/new.css@1.1.2/new.min.css">
  </head>

  <body>
	<header>
		<h1>FairRoulette</h1>
	</header>
	<p>SC address: <code>{{.SCAddress}}</code></p>
	<p>Status fetched at: <code>{{.Now}}</code></p>
	<div>
		<h2>Next play</h2>
		<p>Next play in: <code>{{.Status.NextPlayIn}}</code></p>
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
			<form onsubmit="addAddress(document.getElementById('address').value); return false">
				<fieldset>
					<legend>Player stats:</legend>
					<ul>
						{{range $p, $stats := .Status.PlayerStats}}
							<li>Player <code>{{$p}}</code>: Bets: <code>{{$stats.Bets}}</code> - Wins: <code>{{$stats.Wins}}</code></li>
						{{end}}
					</ul>
					Show address <input type="text" maxlength="45" id="address"></input>
					<input type="submit" value="+">
				</fieldset>
			</form>
		</div>
	</div>
	<hr/>
	<div>
		<h2>CLI usage</h2>
		<h3>Configuration</h3>
		<p><code>fairroulette set goshimmer.api {{.Host}}:8080</code></p>
		<p><code>fairroulette set wasp.api {{.Host}}:9090</code></p>
		<p><code>fairroulette set address {{.SCAddress}}</code></p>
		<p>Initialize a wallet: <code>fairroulette wallet init</code></p>
		<p>Get some funds: <code>fairroulette wallet transfer 1 10000</code></p>

		<h3>Betting</h3>
		<p>Make a bet: <code>fairroulette bet <i>color</i> <i>amount</i></code>
		(e.g.: <code>fairroulette bet 1 100</code>)</p>
		<p>Then refresh this page to see the results.</p>
	</div>
	<script>
		function addAddress(address) {
			const url = new URL(document.location);
			const players = (url.searchParams.get('players') || '').split(',').filter(s => s.length > 0);
			players.push(address)
			document.location.href = document.location.href.split('?')[0] + '?players=' + players.join(',');
		}
	</script>
  </body>
</html>
`)),
}
