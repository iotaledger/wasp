// Copyright 2020 IOTA Stiftung
// SPDX-License-Identifier: Apache-2.0

package client

import (
	"net/http"

	"github.com/iotaledger/wasp/packages/isc"
	"github.com/iotaledger/wasp/packages/webapi/model"
	"github.com/iotaledger/wasp/packages/webapi/routes"
)

// GetChainRecord fetches ChainInfo by address
func (c *WaspClient) GetChainInfo(chainID isc.ChainID) (*model.ChainInfo, error) {
	res := &model.ChainInfo{}
	if err := c.do(http.MethodGet, routes.GetChainInfo(chainID.String()), nil, res); err != nil {
		return nil, err
	}
	return res, nil
}
