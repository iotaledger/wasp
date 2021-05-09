// Copyright 2020 IOTA Stiftung
// SPDX-License-Identifier: Apache-2.0

package dashboard

import (
	_ "embed"
	"html/template"
	"strings"

	"github.com/iotaledger/wasp/packages/chain"
	"github.com/iotaledger/wasp/packages/coretypes"
	"github.com/iotaledger/wasp/packages/kv/dict"
	"github.com/iotaledger/wasp/packages/peering"
	"github.com/iotaledger/wasp/packages/registry_pkg"
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

type WaspServices interface {
	ConfigDump() map[string]interface{}
	ExploreAddressBaseURL() string
	NetworkProvider() peering.NetworkProvider
	GetChainRecords() ([]*registry_pkg.ChainRecord, error)
	GetChainRecord(chainID *coretypes.ChainID) (*registry_pkg.ChainRecord, error)
	GetChainState(chainID *coretypes.ChainID) (*ChainState, error)
	GetChain(chainID *coretypes.ChainID) chain.Chain
	CallView(chain chain.Chain, hname coretypes.Hname, fname string, params dict.Dict) (dict.Dict, error)
}

type Dashboard struct {
	navPages []Tab
	stop     chan bool
	wasp     WaspServices
}

func Init(server *echo.Echo, waspServices WaspServices) *Dashboard {
	r := renderer{}
	server.Renderer = r

	d := &Dashboard{
		stop: make(chan bool),
		wasp: waspServices,
	}
	d.navPages = []Tab{
		d.configInit(server, r),
		d.peeringInit(server, r),
		d.chainsInit(server, r),
	}

	addWsEndpoints(server)
	d.startWsForwarder()

	useHTMLErrorHandler(server)

	return d
}

func (d *Dashboard) Stop() {
	close(d.stop)
}

func (d *Dashboard) BaseParams(c echo.Context, breadcrumbs ...Tab) BaseTemplateParams {
	return BaseTemplateParams{
		NavPages:    d.navPages,
		Breadcrumbs: breadcrumbs,
		Path:        c.Path(),
		MyNetworkId: d.wasp.NetworkProvider().Self().NetID(),
	}
}

func (d *Dashboard) makeTemplate(e *echo.Echo, parts ...string) *template.Template {
	t := template.New("").Funcs(template.FuncMap{
		"formatTimestamp":   formatTimestamp,
		"exploreAddressUrl": exploreAddressUrl(d.wasp.ExploreAddressBaseURL()),
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
