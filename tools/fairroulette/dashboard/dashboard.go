package dashboard

import (
	"fmt"
	"html/template"
	"net/http"
	"os"
	"time"

	"github.com/iotaledger/wasp/tools/fairroulette/client"
)

func check(err error) {
	if err != nil {
		panic(err)
	}
}

func Cmd(args []string) {
	indexTemplate := template.Must(template.New("index").Parse(indexTemplate))
	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		params := &IndexParams{
			Now:    time.Now().UTC(),
			Status: client.FetchStatus(args),
		}
		check(indexTemplate.Execute(w, params))
	})

	listenAddr := ":10000"
	if len(args) > 0 {
		if len(args) != 1 {
			fmt.Printf("Usage: %s dashboard [listen-address]\n", os.Args[0])
			os.Exit(1)
		}
		listenAddr = args[0]
	}
	fmt.Printf("Serving dashboard on %s\n", listenAddr)
	http.ListenAndServe(listenAddr, nil)
}

type IndexParams struct {
	Now        time.Time
	NextPlayIn string
	Status     *client.Status
}

const indexTemplate = `
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
	<div>Status fetched at: {{.Now}}</div>
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
				<li>Player <code>{{slice (.Player.String) 0 6}}</code> bets {{.Sum}} IOTAs on color {{.Color}}</li>
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
		{{if gt (len .Status.PlayerStats) 0}}
			<div>Player stats:<ul>
				{{range $p, $stats := .Status.PlayerStats}}
					<li>Player <code>{{slice ($p.String) 0 6}}</code>: {{$stats}}</li>
				{{end}}
			</ul></div>
		{{end}}
	</div>
  </body>
</html>
`
