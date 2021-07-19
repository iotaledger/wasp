// +build ignore

package dashboard

import (
	"html/template"
	"os"

	"github.com/iotaledger/wasp/packages/dashboard"
	"github.com/iotaledger/wasp/tools/wasp-cli/config"
	"github.com/labstack/echo/v4"
)

type BaseTemplateParams struct {
	NavPages   []NavPage
	ActivePage string
}

var navPages = []NavPage{}

func BaseParams(c echo.Context, page string) BaseTemplateParams {
	return BaseTemplateParams{NavPages: navPages, ActivePage: page}
}

type NavPage struct {
	Title string
	Href  string
}

func MakeTemplate(parts ...string) *template.Template {
	t := template.New("").Funcs(template.FuncMap{
		"formatTimestamp": dashboard.FormatTimestamp,
		"trim":            dashboard.Trim,
		"exploreAddressUrl": dashboard.ExploreAddressUrl(
			dashboard.ExploreAddressUrlFromGoshimmerUri(config.GoshimmerApi()),
		),
		"waspClientCmd": func() string {
			if config.Utxodb {
				return os.Args[0] + " -u"
			}
			return os.Args[0]
		},
	})
	t = template.Must(t.Parse(tplBase))
	t = template.Must(t.Parse(dashboard.TplExploreAddress))
	for _, part := range parts {
		t = template.Must(t.Parse(part))
	}
	return t
}

const tplBase = `
{{define "base"}}
	<!doctype html>
	<html lang="en">
	<head>
		<meta charset="utf-8" />
		<meta http-equiv="x-ua-compatible" content="ie=edge" />
		<meta name="viewport" content="width=device-width, initial-scale=1" />

		<title>{{template "title"}} - IOTA Smart Contracts PoC</title>

		<link rel="stylesheet" href="https://cdn.jsdelivr.net/npm/@xz/fonts@1/serve/inter.css">
		<link rel="stylesheet" href="https://cdn.jsdelivr.net/npm/@exampledev/new.css@1.1.2/new.min.css">
	</head>

	<body>
		<style>
			details {background: #EEF9FF}
		</style>
		<header>
			<h1>IOTA Smart Contracts PoC</h1>
			<nav>
				{{$activePage := .ActivePage}}
				{{range $i, $p := .NavPages}}
					{{if $i}} | {{end}}
					{{if eq $activePage $p.Href}}
						<strong>{{$p.Title}}</strong>
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

const TplSCInfo = `
{{define "sc-info"}}
	<details>
		<summary>Smart contract details</summary>
		<p>SC address: {{template "address" .Status.TargetContract}}</p>
		<p>Program hash: <code>{{.Status.ProgramHash}}</code></p>
		<p>Description of the instance: <code>{{trim .Status.Description}}</code></p>
		<p>Owner address: {{template "address" .Status.OriginatorAddress}}</p>
		<p>Minimum node reward (fee): <code>{{.Status.MinimumReward}}</code></p>
		<p>Color: <code>{{.Config.Record.Color}}</code></p>
		<p>Committee nodes:
			<ul>
				{{range $_, $node := .Config.Record.CommitteeNodes}}
				   <li><code>{{$node}}</code></li>
				{{end}}
			</ul>
		</p>
	</details>
	<h4>State details</h4>
    <p>
	  <ul>
		<li>Index: <code>{{.Status.BlockIndex}}</code></li>
		<li>Timestamp: <code>{{.Status.OutputTimestamp}}</code></li>
		<li>Anchor transaction: <code>{{.Status.StateTxId}}</code></li>
		<li>State hash: <code>{{.Status.StateHash}}</code></li>
		<li>Batched requests:</li>
		<ul>
			{{range $_, $reqid := .Status.Requests}}
			   <li><code>{{$reqid}}</code></li>
			{{end}}
		</ul>
	  </ul>
	</p>

	<h4>Balance:</h4>
	<p>
		<ul>
			{{range $color, $amount := .Status.Balance}}
				<li><code>{{$color}}</code>: <code>{{$amount}} </code></li>
			{{end}}
		</ul>
	</p>
{{end}}`

const TplWs = `
{{define "ws"}}
	<script>
		const url = 'ws://' +  location.host + '/ws/{{.Config.ShortName}}';
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

const TplInstallConfig = `
{{define "install-config"}}
	<details>
		<summary>1. Install wasp-cli</summary>
		<p>Grab the latest <code>wasp-cli</code> binary from the
		<a href="https://github.com/iotaledger/wasp/releases">Releases</a> page.</p>
		<p>-- OR --</p>
		<p>Build from source:</p>
<pre>$ git clone --branch develop https://github.com/iotaledger/wasp.git
$ cd wasp
$ go install ./tools/wasp-cli
</pre>
	</details>
	<details>
		<summary>2. Configure wasp-cli</summary>
		<p>Download <a href="/wasp-cli.json"><code>wasp-cli.json</code></a>. Make
		sure you always run the <code>wasp-cli</code> command in the same folder
		as <code>wasp-cli.json</code>.</p>
		<p>Create an address + private/public keys for your wallet:</p>
		<pre>{{waspClientCmd}} init</pre>
		<p>Get some funds:</p>
		<pre>{{waspClientCmd}} request-funds</pre>
	</details>
{{end}}
`
