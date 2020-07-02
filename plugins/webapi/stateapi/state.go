// access to the solid state of the smart contract
package stateapi

import (
	"github.com/iotaledger/goshimmer/dapps/valuetransfers/packages/address"
	"github.com/iotaledger/wasp/packages/state"
	"github.com/iotaledger/wasp/packages/table"
	"github.com/iotaledger/wasp/plugins/webapi/misc"
	"github.com/labstack/echo"
	"github.com/mr-tron/base58"
)

type QueryStateRequest struct {
	Address   string   `json:"address"`
	Variables []string `json:"variables"`
}

type QueryStateResponse struct {
	Values map[string]string `json:"values"` // variable name: base85 encoded binary data
	Error  string            `json:"error"`
}

func HandlerQueryState(c echo.Context) error {
	var req QueryStateRequest

	if err := c.Bind(&req); err != nil {
		return misc.OkJson(c, &QueryStateResponse{
			Error: err.Error(),
		})
	}
	addr, err := address.FromBase58(req.Address)
	if err != nil {
		return misc.OkJson(c, &QueryStateResponse{
			Error: err.Error(),
		})
	}
	// TODO serialize access to solid state
	state, _, exist, err := state.LoadSolidState(&addr)
	if err != nil {
		return misc.OkJson(c, &QueryStateResponse{
			Error: err.Error(),
		})
	}
	if !exist {
		return misc.OkJson(c, &QueryStateResponse{
			Error: "empty state",
		})
	}
	ret := &QueryStateResponse{
		Values: make(map[string]string),
	}
	for _, v := range req.Variables {
		data, _ := state.Variables().Get(table.Key(v))
		ret.Values[v] = base58.Encode(data)
	}
	return misc.OkJson(c, ret)
}
