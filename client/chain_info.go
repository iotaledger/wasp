// Copyright 2020 IOTA Stiftung
// SPDX-License-Identifier: Apache-2.0

package client

import (
	"github.com/iotaledger/wasp/packages/webapi/v1/model"
	"github.com/iotaledger/wasp/packages/webapi/v1/routes"
	"net/http"

	"github.com/iotaledger/wasp/packages/isc"
)

// GetChainRecord fetches ChainInfo by address
func (c *WaspClient) GetChainInfo(chainID isc.ChainID) (*model.ChainInfo, error) {
	res := &model.ChainInfo{}
	if err := c.do(http.MethodGet, routes.GetChainInfo(chainID.String()), nil, res); err != nil {
		return nil, err
	}
	return res, nil
}
