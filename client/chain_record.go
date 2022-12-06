// Copyright 2020 IOTA Stiftung
// SPDX-License-Identifier: Apache-2.0

package client

import (
	"net/http"

	"github.com/iotaledger/wasp/packages/webapi/v1/model"
	"github.com/iotaledger/wasp/packages/webapi/v1/routes"

	"github.com/iotaledger/wasp/packages/isc"
	"github.com/iotaledger/wasp/packages/registry"
)

// PutChainRecord sends a request to write a Record
func (c *WaspClient) PutChainRecord(rec *registry.ChainRecord) error {
	return c.do(http.MethodPost, routes.PutChainRecord(), model.NewChainRecord(rec), nil)
}

// GetChainRecord fetches a Record by address
func (c *WaspClient) GetChainRecord(chainID isc.ChainID) (*registry.ChainRecord, error) {
	res := &model.ChainRecord{}
	if err := c.do(http.MethodGet, routes.GetChainRecord(chainID.String()), nil, res); err != nil {
		return nil, err
	}
	return res.Record()
}

// GetChainRecordList fetches the list of all chains in the node
func (c *WaspClient) GetChainRecordList() ([]*registry.ChainRecord, error) {
	var res []*model.ChainRecord
	if err := c.do(http.MethodGet, routes.ListChainRecords(), nil, &res); err != nil {
		return nil, err
	}
	list := make([]*registry.ChainRecord, len(res))
	for i, bd := range res {
		rec, err := bd.Record()
		if err != nil {
			return nil, err
		}
		list[i] = rec
	}
	return list, nil
}
