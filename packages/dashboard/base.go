// Copyright 2020 IOTA Stiftung
// SPDX-License-Identifier: Apache-2.0

package dashboard

import (
	"html/template"
	"strings"

	"github.com/iotaledger/wasp/plugins/peering"
	"github.com/labstack/echo/v4"
	"github.com/mr-tron/base58"
)

type Tab struct {
	Path       string
	Title      string
	Href       string
	Breadcrumb bool
}

type BaseTemplateParams struct {
	NavPages    []Tab
	Breadcrumbs []Tab
	Path        string
	MyNetworkId string
}

var navPages []Tab

func Init(server *echo.Echo) {
	r := renderer{}
	server.Renderer = r

	navPages = []Tab{
		configInit(server, r),
		peeringInit(server, r),
		chainsInit(server, r),
	}

	addWsEndpoints(server)
	startWsForwarder()

	useHTMLErrorHandler(server)
}

func BaseParams(c echo.Context, breadcrumbs ...Tab) BaseTemplateParams {
	b := BaseTemplateParams{
		NavPages:    navPages,
		Breadcrumbs: breadcrumbs,
		Path:        c.Path(),
		MyNetworkId: peering.DefaultNetworkProvider().Self().NetID(),
	}
	return b
}

func makeTemplate(e *echo.Echo, parts ...string) *template.Template {
	t := template.New("").Funcs(template.FuncMap{
		"formatTimestamp":   formatTimestamp,
		"exploreAddressUrl": exploreAddressUrl(exploreAddressBaseUrl()),
		"args":              args,
		"hashref":           hashref,
		"trim":              trim,
		"bytesToString":     bytesToString,
		"base58":            base58.Encode,
		"replace":           strings.Replace,
		"uri":               func(s string, p ...interface{}) string { return e.Reverse(s, p...) },
	})
	t = template.Must(t.Parse(tplBase))
	for _, part := range parts {
		t = template.Must(t.Parse(part))
	}
	return t
}

const tplBase = `
{{define "externalLink"}}
	<a href="{{ index . 0 }}" style="font-size: small">🡽 {{ index . 1 }}</a>
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
	<a href="{{ uri "chainAccount" $chainid (replace $agentid.String "/" ":" 1) }}"><tt>{{ $agentid }}</tt></a>
	{{if $agentid.IsNonAliasAddress}} {{ template "exploreAddressInTangle" $agentid.MustAddress }} {{end}}
{{end}}

{{define "balances"}}
	<dl>
		{{range $color, $bal := .}}
			<dt><tt>{{ $color }}</tt></dt>
			<dd>{{ $bal }}</dd>
		{{end}}
	</dl>
{{end}}

{{define "tab"}}
	{{ $title := index . 0 }}
	{{ $href := index . 1 }}
	{{ $active := index . 2 }}
	<a href="{{$href}}" class="button"
		{{- if $active -}}
			style="background-color: var(--header-hover-back-color)"
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
			body {
				--gray2: #EEF2FA;
				--gray3: #D9E1EF;
				--main-green: #0FC1B7;
				--light-green: #00E0CA;
				--dark-gray: #131F37;

				--a-link-color: var(--main-green);
				--a-visited-color: var(--main-green);
				--pre-color: var(--main-green);
				--card-border-color: var(--gray3);
				--card-back-color: white;
				--header-back-color: var(--dark-gray);
				--header-hover-back-color: var(--main-green);
				--header-fore-color: white;
				--back-color: var(--gray2);
				--fore-color: var(--dark-gray);
				--card-fore-color: var(--dark-gray);
				--table-head-back-color: var(--gray2);
				--table-body-back-color: white;
				--table-head-fore-color: var(--dark-gray);
				--table-body-fore-color: var(--dark-gray);
				--table-border-color: var(--gray3);
				--table-border-separator-color: var(--gray3);
			}
			header {
				border: none;
			}
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
			.align-right {
				text-align: right;
			}
			table {
				margin-top: var(--universal-margin);
			}
			table th, table td {
				padding: var(--universal-padding);
			}
			pre {
				margin: 0;
				padding: var(--universal-padding);
			}
			.card {
				padding: 1em;
			}
		</style>

		<header>
				<a class="logo" href="#" style="color: var(--light-green)">Wasp</a>
				{{$path := .Path}}
				{{range $_, $tab := .NavPages}}
					{{ template "tab" (args $tab.Title $tab.Href (eq $path $tab.Path)) }}
				{{end}}
				{{range $_, $tab := .Breadcrumbs}}
					{{ template "tab" (args (printf "🢒 %s" $tab.Title) $tab.Href (eq $path $tab.Path)) }}
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
