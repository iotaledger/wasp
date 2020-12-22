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

type Breadcrumb struct {
	Title  string
	Href   string
	Active bool
}

type BaseTemplateParams struct {
	NavPages    []NavPage
	Breadcrumbs []Breadcrumb
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

func BaseParams(c echo.Context, activePage string, breadcrumbs ...Breadcrumb) BaseTemplateParams {
	b := BaseTemplateParams{
		NavPages:    navPages,
		Breadcrumbs: breadcrumbs,
		ActivePage:  activePage,
		MyNetworkId: peering.DefaultNetworkProvider().Self().NetID(),
	}
	if len(b.Breadcrumbs) > 0 {
		b.Breadcrumbs[len(b.Breadcrumbs)-1].Active = true
	}
	return b
}

func MakeTemplate(parts ...string) *template.Template {
	t := template.New("").Funcs(template.FuncMap{
		"formatTimestamp":   formatTimestamp,
		"exploreAddressUrl": exploreAddressUrl(exploreAddressBaseUrl()),
		"args":              args,
		"hashref":           hashref,
		"quoted":            quoted,
		"bytesToString":     bytesToString,
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
	<tt>{{.}}</tt> {{ template "exploreAddressInTangle" . }}
{{end}}

{{define "agentid"}}
	{{ $chainid := index . 0 }}
	{{ $agentid := index . 1 }}
	<tt>{{ $agentid }}</tt>
	<a href="/chain/{{ $chainid }}/account/{{ $agentid }}" class="linkbtn">Balance</a>
	{{if $agentid.IsAddress}} {{ template "exploreAddressInTangle" $agentid.MustAddress }} {{end}}
{{end}}

{{define "balances"}}
	<dl>
		{{range $color, $bal := .}}
			<dt><tt>{{ $color }}</tt></dt>
			<dd>{{ $bal }}</dd>
		{{end}}
	</dl>
{{end}}

{{define "navitem"}}
	{{ $title := index . 0 }}
	{{ $href := index . 1 }}
	{{ $active := index . 2 }}
	<a href="{{$href}}" class="button"
		{{- if $active -}}
			style="background-color: var(--button-hover-back-color)"
		{{- end -}}
	>
		{{- $title -}}
	</a>
{{end}}

{{define "base"}}
	<!doctype html>
	<html lang="en">
	<head>
		<meta charset="utf-8" />
		<meta http-equiv="x-ua-compatible" content="ie=edge" />
		<meta name="viewport" content="width=device-width, initial-scale=1" />

		<title>{{template "title"}} - Wasp node dashboard</title>

		<link rel="stylesheet" href="https://cdnjs.cloudflare.com/ajax/libs/mini.css/3.0.1/mini-default.min.css">
	</head>

	<body>
		<style>
			tt {
				font-family: Menlo, Consolas, monospace;
			}
			table {
				max-height: none !important;
			}
			dl {
				display: flex;
				flex-wrap: wrap;
				padding: var(--universal-padding);
				align-items: baseline;
			}
			dt {
				width: 33%;
				font-weight: bold;
				text-align: right;
			}
			dt:after {
				content: ":";
			}
			dd {
				margin-left: auto;
				width: 66%;
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
			body {
				--back-color: #eee;
			}
			table th, table td {
				padding: var(--universal-padding);
			}
		</style>

		<header>
				<a class="logo" href="#">Wasp</a>
				{{$activePage := .ActivePage}}
				{{range $_, $p := .NavPages}}
					{{ template "navitem" (args $p.Title $p.Href (eq $activePage $p.Href)) }}
				{{end}}
				{{range $_, $p := .Breadcrumbs}}
					{{ template "navitem" (args (printf "🢒 %s" $p.Title) $p.Href $p.Active) }}
				{{end}}
		</header>
		<main>
			<div class="container">
			<div class="row" style="justify-content: center">
			<div class="col-sm" style="max-width: 65em">
				{{template "body" .}}
			</div>
			</div>
			</div>
		</main>
	</body>
	</html>
{{end}}`
