package client

import (
	"github.com/iotaledger/goshimmer/packages/ledgerstate"
	"net/http"

	"github.com/iotaledger/wasp/packages/registry"
	"github.com/iotaledger/wasp/packages/webapi/model"
	"github.com/iotaledger/wasp/packages/webapi/routes"
)

// PutChainRecord sends a request to write a Record
func (c *WaspClient) PutCommitteeRecord(rec *registry.CommitteeRecord) error {
	return c.do(http.MethodPost, routes.PutCommitteeRecord(), model.NewCommitteeRecord(rec), nil)
}

// GetChainRecord fetches a Record by address
func (c *WaspClient) GetCommitteeRecord(addr ledgerstate.Address) (*registry.ChainRecord, error) {
	res := &model.ChainRecord{}
	if err := c.do(http.MethodGet, routes.GetCommitteeRecord(addr.Base58()), nil, res); err != nil {
		return nil, err
	}
	return res.Record(), nil
}
