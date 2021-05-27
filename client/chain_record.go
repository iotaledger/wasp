package client

import (
	"net/http"

	"github.com/iotaledger/wasp/packages/coretypes"
	"github.com/iotaledger/wasp/packages/registry_pkg/chain_record"
	"github.com/iotaledger/wasp/packages/webapi/model"
	"github.com/iotaledger/wasp/packages/webapi/routes"
)

// PutChainRecord sends a request to write a Record
func (c *WaspClient) PutChainRecord(rec *chain_record.ChainRecord) error {
	return c.do(http.MethodPost, routes.PutChainRecord(), model.NewChainRecord(rec), nil)
}

// GetChainRecord fetches a Record by address
func (c *WaspClient) GetChainRecord(chainid coretypes.ChainID) (*chain_record.ChainRecord, error) {
	res := &model.ChainRecord{}
	if err := c.do(http.MethodGet, routes.GetChainRecord(chainid.Base58()), nil, res); err != nil {
		return nil, err
	}
	return res.Record(), nil
}

// GetChainRecordList fetches the list of all chains in the node
func (c *WaspClient) GetChainRecordList() ([]*chain_record.ChainRecord, error) {
	var res []*model.ChainRecord
	if err := c.do(http.MethodGet, routes.ListChainRecords(), nil, &res); err != nil {
		return nil, err
	}
	list := make([]*chain_record.ChainRecord, len(res))
	for i, bd := range res {
		list[i] = bd.Record()
	}
	return list, nil
}
