package client

import (
	"net/http"

	"github.com/iotaledger/wasp/packages/coretypes"
	"github.com/iotaledger/wasp/packages/webapi/routes"
)

// ActivateChain sends a request to activate a chain in the wasp node
func (c *WaspClient) ActivateChain(chID coretypes.ChainID) error {
	return c.do(http.MethodPost, routes.ActivateChain(chID.Base58()), nil, nil)
}

// DeactivateChain sends a request to deactivate a chain in the wasp node
func (c *WaspClient) DeactivateChain(chID coretypes.ChainID) error {
	return c.do(http.MethodPost, routes.DeactivateChain(chID.Base58()), nil, nil)
}
