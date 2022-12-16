package client

import (
	"net/http"

	"github.com/iotaledger/wasp/packages/isc"
	"github.com/iotaledger/wasp/packages/webapi/v1/model"
	"github.com/iotaledger/wasp/packages/webapi/v1/routes"
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
func (c *WaspClient) GetChainNodeConnectionMetrics(chainID isc.ChainID) (*model.NodeConnectionMessagesMetrics, error) {
	ncmm := &model.NodeConnectionMessagesMetrics{}
	if err := c.do(http.MethodGet, routes.GetChainNodeConnectionMetrics(chainID.String()), nil, ncmm); err != nil {
		return nil, err
	}
	return ncmm, nil
}

// GetNodeConnectionMetrics fetches a consensus workflow status by address
func (c *WaspClient) GetChainConsensusWorkflowStatus(chainID isc.ChainID) (*model.ConsensusWorkflowStatus, error) {
	ncmm := &model.ConsensusWorkflowStatus{}
	if err := c.do(http.MethodGet, routes.GetChainConsensusWorkflowStatus(chainID.String()), nil, ncmm); err != nil {
		return nil, err
	}
	return ncmm, nil
}

func (c *WaspClient) GetChainConsensusPipeMetrics(chainID isc.ChainID) (*model.ConsensusPipeMetrics, error) {
	ncmm := &model.ConsensusPipeMetrics{}
	if err := c.do(http.MethodGet, routes.GetChainConsensusPipeMetrics(chainID.String()), nil, ncmm); err != nil {
		return nil, err
	}
	return ncmm, nil
}
