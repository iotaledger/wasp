package dashboard

import (
	"net"

	"github.com/labstack/echo"
)

type BaseTemplateParams struct {
	Host     string
	NavPages []NavPage
}

func baseParams(c echo.Context, page string) BaseTemplateParams {
	host, _, err := net.SplitHostPort(c.Request().Host)
	if err != nil {
		panic(err)
	}
	return BaseTemplateParams{
		Host: host,
		NavPages: []NavPage{
			NavPage{Title: "FairRoulette", Active: page == "fairroulette", Href: "/fairroulette"},
			NavPage{Title: "FairAuction", Active: page == "fairauction", Href: "/fairauction"},
		},
	}
}

type NavPage struct {
	Title  string
	Active bool
	Href   string
}

const tplBase = `
{{define "base"}}
	<!doctype html>
	<html lang="en">
	<head>
		<meta charset="utf-8" />
		<meta http-equiv="x-ua-compatible" content="ie=edge" />
		<meta name="viewport" content="width=device-width, initial-scale=1" />

		<title>Wasp dashboard - {{template "title"}}</title>

		<link rel="stylesheet" href="https://cdn.jsdelivr.net/npm/@xz/fonts@1/serve/inter.css">
		<link rel="stylesheet" href="https://cdn.jsdelivr.net/npm/@exampledev/new.css@1.1.2/new.min.css">
	</head>

	<body>
		<style>
			details {background: #EEF9FF}
		</style>
		<header>
			<h1>Wasp dashboard</h1>
			<nav>
				{{range $i, $p := .NavPages}}
					{{if $i}} | {{end}}
					{{if $p.Active}}
						{{$p.Title}}
					{{else}}
						<a href="{{$p.Href}}">{{$p.Title}}</a>
					{{end}}
				{{end}}
			</nav>
		</header>
		{{template "body" .}}
	</body>
	</html>
{{end}}`

const tplWs = `
{{define "ws"}}
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
{{end}}
`

const tplInstallConfig = `
{{define "install-config"}}
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
$ wasp-client {{.SC.ShortName}} set address {{.SC.Address}}</pre>
		<p>Initialize a wallet: <code>wasp-client wallet init</code></p>
		<p>Get some funds: <code>wasp-client wallet request-funds</code></p>
	</details>
{{end}}
`
