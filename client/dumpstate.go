package client

import (
	"net/http"

	"github.com/iotaledger/wasp/packages/coretypes"
	"github.com/iotaledger/wasp/packages/kv"
)

type SCStateDump struct {
	Index     uint32            `json:"index"`
	Variables map[kv.Key][]byte `json:"variables"`
}

func DumpSCStateRoute(scid string) string {
	return "sc/" + scid + "/dumpstate"
}

func (c *WaspClient) DumpSCState(scid *coretypes.ContractID) (*SCStateDump, error) {
	res := &SCStateDump{}
	if err := c.do(http.MethodGet, AdminRoutePrefix+"/"+DumpSCStateRoute(scid.String()), nil, res); err != nil {
		return nil, err
	}
	return res, nil
}
