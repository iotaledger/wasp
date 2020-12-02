package client

import (
	"github.com/iotaledger/wasp/packages/coret"
	"net/http"
)

func ActivateChainRoute(chainid string) string {
	return "chain/" + chainid + "/activate"
}

func DeactivateChainRoute(chainid string) string {
	return "chain/" + chainid + "/deactivate"
}

func (c *WaspClient) ActivateChain(chainid *coret.ChainID) error {
	return c.do(http.MethodPost, AdminRoutePrefix+"/"+ActivateChainRoute(chainid.String()), nil, nil)
}

func (c *WaspClient) DeactivateChain(chainid *coret.ChainID) error {
	return c.do(http.MethodPost, AdminRoutePrefix+"/"+DeactivateChainRoute(chainid.String()), nil, nil)
}
