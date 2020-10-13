package client

import (
	"net/http"

	"github.com/iotaledger/goshimmer/dapps/valuetransfers/packages/address"
)

func ActivateSCRoute(address string) string {
	return "sc/" + address + "/activate"
}

func DeactivateSCRoute(address string) string {
	return "sc/" + address + "/deactivate"
}

func (c *WaspClient) ActivateSC(addr *address.Address) error {
	return c.do(http.MethodPost, AdminRoutePrefix+"/"+ActivateSCRoute(addr.String()), nil, nil)
}

func (c *WaspClient) DeactivateSC(addr *address.Address) error {
	return c.do(http.MethodPost, AdminRoutePrefix+"/"+DeactivateSCRoute(addr.String()), nil, nil)
}
