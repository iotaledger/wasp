// Copyright 2020 IOTA Stiftung
// SPDX-License-Identifier: Apache-2.0

package client

// This API is used to maintain the distributed key shares.
// The Golang API in this file tries to follow the REST conventions.

import (
	"net/http"

	iotago "github.com/iotaledger/iota.go/v3"
	"github.com/iotaledger/wasp/packages/webapi/model"
	"github.com/iotaledger/wasp/packages/webapi/routes"
)

// DKSharesPost creates a new DKShare and returns its state.
func (c *WaspClient) DKSharesPost(request *model.DKSharesPostRequest) (*model.DKSharesInfo, error) {
	var response model.DKSharesInfo
	err := c.do(http.MethodPost, routes.DKSharesPost(), request, &response)
	return &response, err
}

// DKSharesGet retrieves the representation of an existing DKShare.
func (c *WaspClient) DKSharesGet(addr iotago.Address) (*model.DKSharesInfo, error) {
	var response model.DKSharesInfo
	err := c.do(http.MethodGet, routes.DKSharesGet(addr.String()), nil, &response)
	return &response, err
}
