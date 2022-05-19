package client

import (
	"net/http"

	"github.com/iotaledger/wasp/packages/iscp"
	"github.com/iotaledger/wasp/packages/webapi/model"
	"github.com/iotaledger/wasp/packages/webapi/routes"
)

// GetNodeConnectionMetrics fetches a connection to L1 metrics for all addresses
func (c *WaspClient) GetNodeConnectionMetrics() (*model.NodeConnectionMetrics, error) {
	ncm := &model.NodeConnectionMetrics{}
	if err := c.do(http.MethodGet, routes.GetChainsNodeConnectionMetrics(), nil, ncm); err != nil {
		return nil, err
	}
	return ncm, nil
}

// GetNodeConnectionMetrics fetches a connection to L1 metrics by address
func (c *WaspClient) GetChainNodeConnectionMetrics(chID *iscp.ChainID) (*model.NodeConnectionMessagesMetrics, error) {
	ncmm := &model.NodeConnectionMessagesMetrics{}
	if err := c.do(http.MethodGet, routes.GetChainNodeConnectionMetrics(chID.String()), nil, ncmm); err != nil {
		return nil, err
	}
	return ncmm, nil
}

// GetNodeConnectionMetrics fetches a consensus workflow status by address
func (c *WaspClient) GetChainConsensusWorkflowStatus(chID *iscp.ChainID) (*model.ConsensusWorkflowStatus, error) {
	ncmm := &model.ConsensusWorkflowStatus{}
	if err := c.do(http.MethodGet, routes.GetChainConsensusWorkflowStatus(chID.String()), nil, ncmm); err != nil {
		return nil, err
	}
	return ncmm, nil
}

func (c *WaspClient) GetChainConsensusPipeMetrics(chID *iscp.ChainID) (*model.ConsensusPipeMetrics, error) {
	ncmm := &model.ConsensusPipeMetrics{}
	if err := c.do(http.MethodGet, routes.GetChainConsensusPipeMetrics(chID.Base58()), nil, ncmm); err != nil {
		return nil, err
	}
	return ncmm, nil
}
