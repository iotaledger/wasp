// Copyright 2020 IOTA Stiftung
// SPDX-License-Identifier: Apache-2.0

package dashboard

import (
	_ "embed"
	"html/template"
	"strings"

	"github.com/labstack/echo/v4"

	"github.com/iotaledger/hive.go/core/logger"
	iotago "github.com/iotaledger/iota.go/v3"
	"github.com/iotaledger/wasp/packages/authentication"
	"github.com/iotaledger/wasp/packages/chain"
	"github.com/iotaledger/wasp/packages/isc"
	"github.com/iotaledger/wasp/packages/kv/dict"
	"github.com/iotaledger/wasp/packages/metrics/nodeconnmetrics"
	"github.com/iotaledger/wasp/packages/registry"
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
	IsAuthenticated bool
	NavPages        []Tab
	Breadcrumbs     []Tab
	Path            string
	MyNetworkID     string
	Version         string
}

type WaspServicesInterface interface {
	WaspVersion() string
	ConfigDump() map[string]interface{}
	ExploreAddressBaseURL() string
	WebAPIPort() string
	PeeringStats() (*PeeringStats, error)
	MyNetworkID() string
	ChainRecords() ([]*registry.ChainRecord, error)
	GetChainRecord(chainID isc.ChainID) (*registry.ChainRecord, error)
	GetChainCommitteeInfo(chainID isc.ChainID) (*chain.CommitteeInfo, error)
	CallView(chainID isc.ChainID, scName, fname string, params dict.Dict) (dict.Dict, error)
	GetChainNodeConnectionMetrics(isc.ChainID) (nodeconnmetrics.NodeConnectionMessagesMetrics, error)
	GetNodeConnectionMetrics() (nodeconnmetrics.NodeConnectionMetrics, error)
	GetChainConsensusWorkflowStatus(isc.ChainID) (chain.ConsensusWorkflowStatus, error)
	GetChainConsensusPipeMetrics(isc.ChainID) (chain.ConsensusPipeMetrics, error)
}

type Dashboard struct {
	log      *logger.Logger
	wasp     WaspServicesInterface
	navPages []Tab
}

func New(log *logger.Logger, server *echo.Echo, waspServices WaspServicesInterface) *Dashboard {
	r := renderer{}
	server.Renderer = r

	d := &Dashboard{
		log:  log.Named("dashboard"),
		wasp: waspServices,
	}

	d.errorInit(server, r)

	d.navPages = []Tab{
		d.authInit(server, r),
		d.configInit(server, r),
		d.peeringInit(server, r),
		d.chainsInit(server, r),
		d.metricsInit(server, r),
	}

	return d
}

func (d *Dashboard) BaseParams(c echo.Context, breadcrumbs ...Tab) BaseTemplateParams {
	var isAuthenticated bool

	auth, ok := c.Get("auth").(*authentication.AuthContext)

	if !ok {
		isAuthenticated = false
	} else {
		isAuthenticated = auth.IsAuthenticated()
	}

	return BaseTemplateParams{
		IsAuthenticated: isAuthenticated,
		NavPages:        d.navPages,
		Breadcrumbs:     breadcrumbs,
		Path:            c.Path(),
		MyNetworkID:     d.wasp.MyNetworkID(),
		Version:         d.wasp.WaspVersion(),
	}
}

func EVMJSONRPC(chainIDBech32 string) string {
	return "/chains/" + chainIDBech32 + "/evm/jsonrpc"
}

func (d *Dashboard) makeTemplate(e *echo.Echo, parts ...string) *template.Template {
	t := template.New("").Funcs(template.FuncMap{
		"formatTimestamp":        formatTimestamp,
		"formatTimestampOrNever": formatTimestampOrNever,
		"exploreAddressUrl":      d.exploreAddressURL,
		"args":                   args,
		"hashref":                hashref,
		"chainIDBech32":          chainIDBech32,
		"assedID":                assetID,
		"trim":                   trim,
		"incUint32":              incUint32,
		"decUint32":              decUint32,
		"bytesToString":          bytesToString,
		"addressToString":        d.addressToString,
		"agentIDToString":        d.agentIDToString,
		"addressFromAgentID":     d.addressFromAgentID,
		"getETHAddress":          d.getETHAddress,
		"isETHAddress":           d.isETHAddress,
		"isValidAddress":         d.isValidAddress,
		"keyToString":            keyToString,
		"anythingToString":       anythingToString,
		"hex":                    iotago.EncodeHex,
		"replace":                strings.Replace,
		"webapiPort":             d.wasp.WebAPIPort,
		"evmJSONRPCEndpoint":     EVMJSONRPC,
		"uri":                    func(s string, p ...interface{}) string { return e.Reverse(s, p...) },
		"href":                   func(s string) string { return s },
	})
	t = template.Must(t.Parse(tplBase))
	for _, part := range parts {
		t = template.Must(t.Parse(part))
	}
	return t
}
