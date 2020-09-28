package admapi

import (
	"net/http"

	"github.com/iotaledger/goshimmer/dapps/valuetransfers/packages/address"
	"github.com/iotaledger/wasp/packages/kv"
	"github.com/iotaledger/wasp/packages/state"
	"github.com/labstack/echo"
)

type DumpSCStateResponse struct {
	Err       string            `json:"error"`
	Exists    bool              `json:"exists"`
	Index     uint32            `json:"index"`
	Variables map[kv.Key][]byte `json:"variables"`
}

func HandlerDumpSCState(c echo.Context) error {
	scAddress, err := address.FromBase58(c.Param("scaddress"))
	if err != nil {
		return c.JSON(http.StatusBadRequest, &DumpSCStateResponse{Err: err.Error()})
	}

	virtualState, _, ok, err := state.LoadSolidState(&scAddress)
	if err != nil {
		return c.JSON(http.StatusBadRequest, &DumpSCStateResponse{Err: err.Error()})
	}
	if !ok {
		return c.JSON(http.StatusOK, &DumpSCStateResponse{Exists: false})
	}

	return c.JSON(http.StatusOK, &DumpSCStateResponse{
		Exists:    true,
		Index:     virtualState.StateIndex(),
		Variables: virtualState.Variables().DangerouslyDumpToMap().ToGoMap(),
	})
}
