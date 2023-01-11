// Copyright 2020 IOTA Stiftung
// SPDX-License-Identifier: Apache-2.0

package client

import (
	"net/http"

	"github.com/iotaledger/wasp/packages/isc"
	"github.com/iotaledger/wasp/packages/webapi/v1/routes"
)

func (c *WaspClient) AddAccessNode(chainID isc.ChainID, pubKey string) error {
	return c.do(http.MethodPut, routes.AdmAddAccessNode(chainID.String(), pubKey), nil, nil)
}

func (c *WaspClient) RemoveAccessNode(chainID isc.ChainID, pubKey string) error {
	return c.do(http.MethodDelete, routes.AdmRemoveAccessNode(chainID.String(), pubKey), nil, nil)
}
