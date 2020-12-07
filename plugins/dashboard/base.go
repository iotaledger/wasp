package dashboard

// Copyright 2020 IOTA Stiftung
// SPDX-License-Identifier: Apache-2.0

import (
	"html/template"

	"github.com/iotaledger/wasp/packages/dashboard"
	"github.com/iotaledger/wasp/packages/parameters"
	"github.com/iotaledger/wasp/plugins/peering"
	"github.com/labstack/echo"
)

type BaseTemplateParams struct {
	NavPages    []NavPage
	ActivePage  string
	MyNetworkId string
}

var navPages = []NavPage{}

func BaseParams(c echo.Context, activePage string) BaseTemplateParams {
	return BaseTemplateParams{
		NavPages:    navPages,
		ActivePage:  activePage,
		MyNetworkId: peering.DefaultNetworkProvider().Self().Location(),
	}
}

type NavPage interface {
	Title() string
	Href() string

	AddTemplates(renderer Renderer)
	AddEndpoints(e *echo.Echo)
}

func MakeTemplate(parts ...string) *template.Template {
	t := template.New("").Funcs(template.FuncMap{
		"formatTimestamp":   dashboard.FormatTimestamp,
		"exploreAddressUrl": dashboard.ExploreAddressUrl(exploreAddressBaseUrl()),
	})
	t = template.Must(t.Parse(tplBase))
	t = template.Must(t.Parse(dashboard.TplExploreAddress))
	for _, part := range parts {
		t = template.Must(t.Parse(part))
	}
	return t
}

func exploreAddressBaseUrl() string {
	baseUrl := parameters.GetString(parameters.DashboardExploreAddressUrl)
	if baseUrl != "" {
		return baseUrl
	}
	return dashboard.ExploreAddressUrlFromGoshimmerUri(parameters.GetString(parameters.NodeAddress))
}

const tplBase = `
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
			details {background: #EEF9FF}
		</style>
		<header>
			<h1>Wasp node dashboard</h1>
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
		<p>Node network ID: <code>{{.MyNetworkId}}</code></p>
		{{template "body" .}}
	</body>
	</html>
{{end}}`
