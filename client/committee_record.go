package client

import (
	"net/http"

	iotago "github.com/iotaledger/iota.go/v3"
	"github.com/iotaledger/wasp/packages/iscp"
	"github.com/iotaledger/wasp/packages/registry"
	"github.com/iotaledger/wasp/packages/webapi/model"
	"github.com/iotaledger/wasp/packages/webapi/routes"
)

// PutCommitteeRecord sends a request to write a Record
func (c *WaspClient) PutCommitteeRecord(rec *registry.CommitteeRecord) error {
	return c.do(http.MethodPost, routes.PutCommitteeRecord(), model.NewCommitteeRecord(rec), nil)
}

// GetCommitteeRecord fetches a Record by address
func (c *WaspClient) GetCommitteeRecord(addr iotago.Address) (*registry.CommitteeRecord, error) {
	res := &model.CommitteeRecord{}
	if err := c.do(http.MethodGet, routes.GetCommitteeRecord(addr.String()), nil, res); err != nil {
		return nil, err
	}
	return res.Record(), nil
}

// GetCommitteeForChain fetches the CommitteeRecord that manages the given chain
func (c *WaspClient) GetCommitteeForChain(chainID *iscp.ChainID) (*registry.CommitteeRecord, error) {
	res := &model.CommitteeRecord{}
	if err := c.do(http.MethodGet, routes.GetCommitteeForChain(chainID.Hex())+"?includeDeactivated=true", nil, res); err != nil {
		return nil, err
	}
	return res.Record(), nil
}
