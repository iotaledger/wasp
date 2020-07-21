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
  </head>

  <body>
  	<h1>FairRoulette</h1>
	<div>SC address: <code>{{.SCAddress}}</code></div>
	<div>Status fetched at: <code>{{.Now}}</code></div>
	<div>
		<h2>Next play</h2>
		<div>Next play in: {{.Status.NextPlayIn}}</div>
		<div>Play period: {{.Status.PlayPeriodSeconds}}s</div>
		<div>
			<div>Bets: {{.Status.CurrentBetsAmount}}</div>
			{{if lt (len .Status.CurrentBets) .Status.CurrentBetsAmount}}
				<div>(Showing first {{(len .Status.CurrentBets)}})</div>
			{{end}}
			<ul>
			{{range .Status.CurrentBets}}
				<li>Player <code>{{.Player}}</code> bets {{.Sum}} IOTAs on color {{.Color}}</li>
			{{end}}
			</ul>
		</div>
	</div>

	<div>
		<h2>Stats</h2>
		<div>Last winning color: {{.Status.LastWinningColor}}</div>
		<div>Color stats:<ul>
			{{range $c, $w := .Status.WinsPerColor}}
				<li>Color {{$c}} won {{$w}} times so far</li>
			{{end}}
		</ul></div>
		<div>Player stats:<ul>
			{{range $p, $stats := .Status.PlayerStats}}
				<li>Player <code>{{$p}}</code>: {{$stats}}</li>
			{{end}}
			<form onsubmit="addAddress(document.getElementById('address').value); return false">
				<div>Show player stats for address: <input type="text" maxlength="45" id="address"></input></div>
				<input type="submit" value="Submit">
			</form>
		</ul></div>
	</div>

	<div>
		<h2>CLI usage</h2>
		<h2>Configuration</h2>
		<div><code>fairroulette set goshimmer.api {{.Host}}:8080</code></div>
		<div><code>fairroulette set wasp.api {{.Host}}:9090</code></div>
		<div><code>fairroulette set address {{.SCAddress}}</code></div>
		<div>Initialize a wallet: <code>fairroulette wallet init</code></div>
		<div>Get some funds: <code>fairroulette wallet transfer 1 10000</code></div>
		<h2>Betting</h2>
		<div>Make a bet: <code>fairroulette bet 1 100</code></div>
		<div>Then refresh this page to see the results.</div>
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
