package client

import (
	"net/http"

	"github.com/iotaledger/wasp/packages/coretypes"
	"github.com/iotaledger/wasp/packages/kv/dict"
)

func StateViewRoute(scID string, fname string) string {
	return "sc/" + scID + "/state/view/" + fname
}

func (c *WaspClient) StateView(scID *coretypes.ContractID, fname string, arguments dict.Dict) (dict.Dict, error) {
	res := dict.New()
	if err := c.do(http.MethodGet, StateViewRoute(scID.Base58(), fname), arguments, res); err != nil {
		return nil, err
	}
	return res, nil
}
