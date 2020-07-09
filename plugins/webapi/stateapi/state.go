// access to the solid state of the smart contract
package stateapi

import (
	"fmt"
	"net/http"

	"github.com/iotaledger/goshimmer/dapps/valuetransfers/packages/address"
	"github.com/iotaledger/wasp/packages/kv"
	"github.com/iotaledger/wasp/packages/state"
	"github.com/iotaledger/wasp/plugins/webapi/misc"
	"github.com/labstack/echo"
)

type QueryStateRequest struct {
	Address string   `json:"address"`
	Keys    [][]byte `json:"keys"`
}

type KeyValuePair struct {
	Key   []byte `json:"k"`
	Value []byte `json:"v"`
}

type QueryStateResponse struct {
	Values []KeyValuePair `json:"values"`
	Error  string         `json:"error"`
}

func HandlerQueryState(c echo.Context) error {
	var req QueryStateRequest

	if err := c.Bind(&req); err != nil {
		return c.JSON(http.StatusBadRequest, &QueryStateResponse{Error: err.Error()})
	}
	addr, err := address.FromBase58(req.Address)
	if err != nil {
		return c.JSON(http.StatusBadRequest, &QueryStateResponse{Error: err.Error()})
	}
	// TODO serialize access to solid state
	state, _, exist, err := state.LoadSolidState(&addr)
	if err != nil {
		return c.JSON(http.StatusInternalServerError, &QueryStateResponse{Error: err.Error()})
	}
	if !exist {
		return c.JSON(http.StatusNotFound, &QueryStateResponse{
			Error: fmt.Sprintf("State not found with address %s", addr),
		})
	}
	ret := &QueryStateResponse{
		Values: make([]KeyValuePair, 0),
	}
	for _, k := range req.Keys {
		data, err := state.Variables().Get(kv.Key(k))
		if err != nil {
			return c.JSON(http.StatusInternalServerError, &QueryStateResponse{Error: err.Error()})
		}
		ret.Values = append(ret.Values, KeyValuePair{Key: k, Value: data})
	}
	return misc.OkJson(c, ret)
}
