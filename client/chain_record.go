package client

import (
	"net/http"

	"github.com/iotaledger/wasp/packages/coretypes/chainid"
	"github.com/iotaledger/wasp/packages/registry/chainrecord"
	"github.com/iotaledger/wasp/packages/webapi/model"
	"github.com/iotaledger/wasp/packages/webapi/routes"
)

// PutChainRecord sends a request to write a Record
func (c *WaspClient) PutChainRecord(rec *chainrecord.ChainRecord) error {
	return c.do(http.MethodPost, routes.PutChainRecord(), model.NewChainRecord(rec), nil)
}

// GetChainRecord fetches a Record by address
func (c *WaspClient) GetChainRecord(chID chainid.ChainID) (*chainrecord.ChainRecord, error) {
	res := &model.ChainRecord{}
	if err := c.do(http.MethodGet, routes.GetChainRecord(chID.Base58()), nil, res); err != nil {
		return nil, err
	}
	return res.Record(), nil
}

// GetChainRecordList fetches the list of all chains in the node
func (c *WaspClient) GetChainRecordList() ([]*chainrecord.ChainRecord, error) {
	var res []*model.ChainRecord
	if err := c.do(http.MethodGet, routes.ListChainRecords(), nil, &res); err != nil {
		return nil, err
	}
	list := make([]*chainrecord.ChainRecord, len(res))
	for i, bd := range res {
		list[i] = bd.Record()
	}
	return list, nil
}
