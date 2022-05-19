package client

import (
	"net/http"

	"github.com/iotaledger/wasp/packages/iscp"
	"github.com/iotaledger/wasp/packages/webapi/routes"
)

// ActivateChain sends a request to activate a chain in the wasp node
func (c *WaspClient) ActivateChain(chID *iscp.ChainID) error {
	return c.do(http.MethodPost, routes.ActivateChain(chID.String()), nil, nil)
}

// DeactivateChain sends a request to deactivate a chain in the wasp node
func (c *WaspClient) DeactivateChain(chID *iscp.ChainID) error {
	return c.do(http.MethodPost, routes.DeactivateChain(chID.String()), nil, nil)
}
