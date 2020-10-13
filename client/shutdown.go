package client

import (
	"net/http"
)

const (
	ShutdownRoute = "shutdown"
)

func (c *WaspClient) Shutdown() error {
	return c.do(http.MethodGet, AdminRoutePrefix+"/"+ShutdownRoute, nil, nil)
}
