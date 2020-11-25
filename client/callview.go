package client

import (
	"net/http"

	"github.com/iotaledger/wasp/packages/coretypes"
	"github.com/iotaledger/wasp/packages/kv/dict"
)

func CallViewRoute(scID string, fname string) string {
	return "sc/" + scID + "/state/view/" + fname
}

func (c *WaspClient) CallView(scID coretypes.ContractID, fname string, arguments dict.Dict) (dict.Dict, error) {
	var res dict.Dict
	if err := c.do(http.MethodGet, CallViewRoute(scID.Base58(), fname), arguments, &res); err != nil {
		return nil, err
	}
	return res, nil
}
