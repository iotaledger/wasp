package client

import (
	"fmt"
	"net/http"

	"github.com/iotaledger/wasp/packages/iscp"
	"github.com/iotaledger/wasp/packages/webapi/model"
	"github.com/iotaledger/wasp/packages/webapi/routes"
)

const maxMessageLen = 80

// GetNodeConnectionMetrics fetches a connection to L1 metrics for all addresses
func (c *WaspClient) GetNodeConnectionMetrics() ([]string, [][]string, error) {
	ncm := &model.NodeConnectionMetrics{}
	if err := c.do(http.MethodGet, routes.GetChainsNodeConnectionMetrics(), nil, ncm); err != nil {
		return nil, nil, err
	}
	subscribed := make([]string, len(ncm.Subscribed))
	for i := range subscribed {
		subscribed[i] = string(ncm.Subscribed[i])
	}
	return subscribed, makeNodeConnMetricsTable(&(ncm.NodeConnectionMessagesMetrics)), nil
}

// GetNodeConnectionMetrics fetches a connection to L1 metrics by address
func (c *WaspClient) GetChainNodeConnectionMetrics(chID *iscp.ChainID) ([][]string, error) {
	ncmm := &model.NodeConnectionMessagesMetrics{}
	if err := c.do(http.MethodGet, routes.GetChainNodeConnectionMetrics(chID.Base58()), nil, ncmm); err != nil {
		return nil, err
	}
	return makeNodeConnMetricsTable(ncmm), nil
}

func makeNodeConnMetricsTable(ncmm *model.NodeConnectionMessagesMetrics) [][]string {
	res := make([][]string, 8)
	res[0] = makeNodeConnMetricsTableRow("Pull state", false, ncmm.OutPullState)
	res[1] = makeNodeConnMetricsTableRow("Pull tx inclusion state", false, ncmm.OutPullTransactionInclusionState)
	res[2] = makeNodeConnMetricsTableRow("Pull confirmed output", false, ncmm.OutPullConfirmedOutput)
	res[3] = makeNodeConnMetricsTableRow("Post transaction", false, ncmm.OutPostTransaction)
	res[4] = makeNodeConnMetricsTableRow("Transaction", true, ncmm.InTransaction)
	res[5] = makeNodeConnMetricsTableRow("Inclusion state", true, ncmm.InInclusionState)
	res[6] = makeNodeConnMetricsTableRow("Output", true, ncmm.InOutput)
	res[7] = makeNodeConnMetricsTableRow("Unspent alias output", true, ncmm.InUnspentAliasOutput)
	return res
}

func makeNodeConnMetricsTableRow(name string, isIn bool, ncmm *model.NodeConnectionMessageMetrics) []string {
	res := make([]string, 5)
	res[0] = name
	if isIn {
		res[1] = "IN"
	} else {
		res[1] = "OUT"
	}
	res[2] = fmt.Sprintf("%v", ncmm.Total)
	res[3] = fmt.Sprintf("%v", ncmm.LastEvent)
	res[4] = ncmm.LastMessage
	if len(res[4]) > maxMessageLen {
		res[4] = res[4][:maxMessageLen]
	}
	return res
}
