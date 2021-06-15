package client

import (
	"net/http"

	"github.com/iotaledger/wasp/packages/coretypes/chainid"

	"github.com/iotaledger/wasp/packages/webapi/routes"
)

// ActivateChain sends a request to activate a chain in the wasp node
func (c *WaspClient) ActivateChain(chainid chainid.ChainID) error {
	return c.do(http.MethodPost, routes.ActivateChain(chainid.Base58()), nil, nil)
}

// DeactivateChain sends a request to deactivate a chain in the wasp node
func (c *WaspClient) DeactivateChain(chainid chainid.ChainID) error {
	return c.do(http.MethodPost, routes.DeactivateChain(chainid.Base58()), nil, nil)
}
