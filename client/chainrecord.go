package client

import (
	"net/http"

	"github.com/iotaledger/wasp/packages/coretypes"
	"github.com/iotaledger/wasp/packages/registry"
	"github.com/iotaledger/wasp/packages/webapi/model"
	"github.com/iotaledger/wasp/packages/webapi/routes"
)

// PutChainRecord sends a request to write a ChainRecord
func (c *WaspClient) PutChainRecord(bd *registry.ChainRecord) error {
	return c.do(http.MethodPost, routes.PutChainRecord(), model.NewChainRecord(bd), nil)
}

// GetChainRecord fetches a ChainRecord by address
func (c *WaspClient) GetChainRecord(chainid coretypes.ChainID) (*registry.ChainRecord, error) {
	res := &model.ChainRecord{}
	if err := c.do(http.MethodGet, routes.GetChainRecord(chainid.String()), nil, res); err != nil {
		return nil, err
	}
	return res.ChainRecord(), nil
}

// GetChainRecordList fetches the list of all chains in the node
func (c *WaspClient) GetChainRecordList() ([]*registry.ChainRecord, error) {
	var res []*model.ChainRecord
	if err := c.do(http.MethodGet, routes.ListChainRecords(), nil, &res); err != nil {
		return nil, err
	}
	list := make([]*registry.ChainRecord, len(res))
	for i, bd := range res {
		list[i] = bd.ChainRecord()
	}
	return list, nil
}
