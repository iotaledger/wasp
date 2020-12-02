package client

import (
	"github.com/iotaledger/wasp/packages/coret"
	"net/http"

	"github.com/iotaledger/wasp/client/jsonable"
	"github.com/iotaledger/wasp/packages/registry"
)

const (
	PutChainRecordRoute     = "chainrecord"
	GetChainRecordListRoute = "chainrecord"
)

func GetChainRecordRoute(chainid string) string {
	return "chainrecord/" + chainid
}

// PutChainRecord calls node to write ChainRecord record
func (c *WaspClient) PutChainRecord(bd *registry.ChainRecord) error {
	return c.do(http.MethodPost, AdminRoutePrefix+"/"+PutChainRecordRoute, jsonable.NewChainRecord(bd), nil)
}

// GetChainRecord calls node to get ChainRecord record by address
func (c *WaspClient) GetChainRecord(chainid coret.ChainID) (*registry.ChainRecord, error) {
	res := &jsonable.ChainRecord{}
	if err := c.do(http.MethodGet, AdminRoutePrefix+"/"+GetChainRecordRoute(chainid.String()), nil, res); err != nil {
		return nil, err
	}
	return res.ChainRecord(), nil
}

// gets list of all SCs from the node
func (c *WaspClient) GetChainRecordList() ([]*registry.ChainRecord, error) {
	var res []*jsonable.ChainRecord
	if err := c.do(http.MethodGet, AdminRoutePrefix+"/"+GetChainRecordListRoute, nil, &res); err != nil {
		return nil, err
	}
	list := make([]*registry.ChainRecord, len(res))
	for i, bd := range res {
		list[i] = bd.ChainRecord()
	}
	return list, nil
}
