// Copyright 2020 IOTA Stiftung
// SPDX-License-Identifier: Apache-2.0

package dashboard

import (
	_ "embed"
	"html/template"
	"strings"

	"github.com/iotaledger/hive.go/logger"
	"github.com/iotaledger/wasp/packages/chain"
	"github.com/iotaledger/wasp/packages/iscp"
	"github.com/iotaledger/wasp/packages/kv/dict"
	"github.com/iotaledger/wasp/packages/registry"
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
	MyNetworkID string
}

type WaspServices interface {
	ConfigDump() map[string]interface{}
	ExploreAddressBaseURL() string
	PeeringStats() (*PeeringStats, error)
	MyNetworkID() string
	GetChainRecords() ([]*registry.ChainRecord, error)
	GetChainRecord(chainID *iscp.ChainID) (*registry.ChainRecord, error)
	GetChainCommitteeInfo(chainID *iscp.ChainID) (*chain.CommitteeInfo, error)
	CallView(chainID *iscp.ChainID, scName, fname string, params dict.Dict) (dict.Dict, error)
}

type Dashboard struct {
	navPages []Tab
	stop     chan bool
	wasp     WaspServices
	log      *logger.Logger
}

func Init(server *echo.Echo, waspServices WaspServices, log *logger.Logger) *Dashboard {
	r := renderer{}
	server.Renderer = r

	d := &Dashboard{
		stop: make(chan bool),
		wasp: waspServices,
		log:  log.Named("dashboard"),
	}

	d.errorInit(server, r)

	d.navPages = []Tab{
		d.configInit(server, r),
		d.peeringInit(server, r),
		d.chainsInit(server, r),
	}

	d.webSocketInit(server)

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
		MyNetworkID: d.wasp.MyNetworkID(),
	}
}

func (d *Dashboard) makeTemplate(e *echo.Echo, parts ...string) *template.Template {
	t := template.New("").Funcs(template.FuncMap{
		"formatTimestamp":   formatTimestamp,
		"exploreAddressUrl": exploreAddressURL(d.wasp.ExploreAddressBaseURL()),
		"args":              args,
		"hashref":           hashref,
		"colorref":          colorref,
		"trim":              trim,
		"incUint32":         incUint32,
		"decUint32":         decUint32,
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
