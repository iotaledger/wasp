// Copyright 2020 IOTA Stiftung
// SPDX-License-Identifier: Apache-2.0

package client

import (
	"net/http"

	"github.com/iotaledger/wasp/packages/isc"
	"github.com/iotaledger/wasp/packages/registry"
	"github.com/iotaledger/wasp/packages/webapi/model"
	"github.com/iotaledger/wasp/packages/webapi/routes"
)

// PutChainRecord sends a request to write a Record
func (c *WaspClient) PutChainRecord(rec *registry.ChainRecord) error {
	return c.do(http.MethodPost, routes.PutChainRecord(), model.NewChainRecord(rec), nil)
}

// GetChainRecord fetches a Record by address
func (c *WaspClient) GetChainRecord(chID *isc.ChainID) (*registry.ChainRecord, error) {
	res := &model.ChainRecord{}
	if err := c.do(http.MethodGet, routes.GetChainRecord(chID.String()), nil, res); err != nil {
		return nil, err
	}
	return res.Record(), nil
}

// GetChainRecordList fetches the list of all chains in the node
func (c *WaspClient) GetChainRecordList() ([]*registry.ChainRecord, error) {
	var res []*model.ChainRecord
	if err := c.do(http.MethodGet, routes.ListChainRecords(), nil, &res); err != nil {
		return nil, err
	}
	list := make([]*registry.ChainRecord, len(res))
	for i, bd := range res {
		list[i] = bd.Record()
	}
	return list, nil
}
