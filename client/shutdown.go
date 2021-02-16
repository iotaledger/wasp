package client

import (
	"net/http"

	"github.com/iotaledger/wasp/packages/webapi/routes"
)

// Shutdown shuts down the node
func (c *WaspClient) Shutdown() error {
	return c.do(http.MethodGet, routes.Shutdown(), nil, nil)
}
