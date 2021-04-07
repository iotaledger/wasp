// Copyright 2020 IOTA Stiftung
// SPDX-License-Identifier: Apache-2.0

package dashboard

import (
	_ "embed"
	"html/template"
	"strings"

	"github.com/iotaledger/wasp/plugins/peering"
	"github.com/labstack/echo/v4"
	"github.com/mr-tron/base58"
)

//go:embed templates/base.tmpl
var tplBase string

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
