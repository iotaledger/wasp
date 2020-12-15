// Copyright 2020 IOTA Stiftung
// SPDX-License-Identifier: Apache-2.0

package dashboard

import (
	"html/template"

	"github.com/iotaledger/wasp/plugins/peering"
	"github.com/labstack/echo"
)

type NavPage interface {
	Title() string
	Href() string

	AddTemplates(r renderer)
	AddEndpoints(e *echo.Echo)
}

type BaseTemplateParams struct {
	NavPages    []NavPage
	ActivePage  string
	MyNetworkId string
}

var navPages = []NavPage{
	&configNavPage{},
	&peeringNavPage{},
	&chainsNavPage{},
}

func Init(server *echo.Echo) {
	r := renderer{}
	server.Renderer = r

	for _, navPage := range navPages {
		navPage.AddTemplates(r)
		navPage.AddEndpoints(server)
	}

	useHTMLErrorHandler(server)
}

func BaseParams(c echo.Context, activePage string) BaseTemplateParams {
	return BaseTemplateParams{
		NavPages:    navPages,
		ActivePage:  activePage,
		MyNetworkId: peering.DefaultNetworkProvider().Self().NetID(),
	}
}

func MakeTemplate(parts ...string) *template.Template {
	t := template.New("").Funcs(template.FuncMap{
		"formatTimestamp":   formatTimestamp,
		"exploreAddressUrl": exploreAddressUrl(exploreAddressBaseUrl()),
		"args":              args,
	})
	t = template.Must(t.Parse(tplBase))
	for _, part := range parts {
		t = template.Must(t.Parse(part))
	}
	return t
}

const tplBase = `
{{define "externalLink"}}
	<a href="{{ index . 0 }}" class="linkbtn">🡽 {{ index . 1 }}</a>
{{end}}

{{define "exploreAddressInTangle"}}
	{{ template "externalLink" (args (exploreAddressUrl .) "Tangle") }}
{{end}}

{{define "address"}}
	<code>{{.}}</code> {{ template "exploreAddressInTangle" . }}
{{end}}

{{define "agentid"}}
	{{ $chainid := index . 0 }}
	{{ $agentid := index . 1 }}
	<code>{{ $agentid }}</code>
	<a href="/chains/{{ $chainid }}/account/{{ $agentid }}" class="linkbtn">Balance</a>
	{{if $agentid.IsAddress}} {{ template "exploreAddressInTangle" $agentid.MustAddress }} {{end}}
{{end}}

{{define "balances"}}
	<table>
		<thead>
			<tr>
				<th class="align-right">Color</th>
				<th class="align-right">Balance</th>
			</tr>
		</thead>
		<tbody>
		{{range $color, $bal := .}}
			<tr>
				<td class="align-right"><tt>{{ $color }}</tt></td>
				<td class="align-right"><tt>{{ $bal }}</tt></td>
			</tr>
		{{end}}
		</tbody>
	</table>
{{end}}

{{define "base"}}
	<!doctype html>
	<html lang="en">
	<head>
		<meta charset="utf-8" />
		<meta http-equiv="x-ua-compatible" content="ie=edge" />
		<meta name="viewport" content="width=device-width, initial-scale=1" />

		<title>{{template "title"}} - Wasp node dashboard</title>

		<link rel="stylesheet" href="https://cdn.jsdelivr.net/npm/@xz/fonts@1/serve/inter.css">
		<link rel="stylesheet" href="https://cdn.jsdelivr.net/npm/@exampledev/new.css@1.1.2/new.min.css">
	</head>

	<body>
		<style>
			details {
				background: #EEF9FF;
			}
			.linkbtn {
				font-size: small;
				border: 1px solid var(--nc-lk-1);
				padding: 2px;
				text-decoration: none;
			}
			.align-right {
				text-align: right;
			}
		</style>
		<header>
			<h1>Wasp node dashboard</h1>
			<nav>
				{{$activePage := .ActivePage}}
				{{range $i, $p := .NavPages}}
					{{if $i}} | {{end}}
					<a href="{{$p.Href}}">
						{{- if eq $activePage $p.Href -}}
							<strong>{{$p.Title}}</strong>
						{{- else -}}
							{{$p.Title}}
						{{- end -}}
					</a>
				{{end}}
			</nav>
		</header>
		{{template "body" .}}
		<hr/>
		<footer>
		<p>Node network ID: <code>{{.MyNetworkId}}</code></p>
		</footer>
	</body>
	</html>
{{end}}`
