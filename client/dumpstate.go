package client

import (
	"net/http"

	"github.com/iotaledger/wasp/packages/coret"
	"github.com/iotaledger/wasp/packages/kv/dict"
)

type SCStateDump struct {
	Index     uint32    `json:"index"`
	Variables dict.Dict `json:"variables"`
}

func DumpSCStateRoute(scid string) string {
	return "sc/" + scid + "/dumpstate"
}

func (c *WaspClient) DumpSCState(scid *coret.ContractID) (*SCStateDump, error) {
	res := &SCStateDump{}
	if err := c.do(http.MethodGet, AdminRoutePrefix+"/"+DumpSCStateRoute(scid.Base58()), nil, res); err != nil {
		return nil, err
	}
	return res, nil
}
