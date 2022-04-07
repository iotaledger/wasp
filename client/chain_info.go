// Copyright 2020 IOTA Stiftung
// SPDX-License-Identifier: Apache-2.0

package client

import (
	"net/http"

	"github.com/iotaledger/wasp/packages/iscp"
	"github.com/iotaledger/wasp/packages/webapi/model"
	"github.com/iotaledger/wasp/packages/webapi/routes"
)

// GetChainRecord fetches ChainInfo by address
func (c *WaspClient) GetChainInfo(chID *iscp.ChainID) (*model.ChainInfo, error) {
	res := &model.ChainInfo{}
	if err := c.do(http.MethodGet, routes.GetChainInfo(chID.String()), nil, res); err != nil {
		return nil, err
	}
	return res, nil
}
