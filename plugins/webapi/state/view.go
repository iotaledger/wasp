package state

import (
	"fmt"
	"net/http"

	"github.com/iotaledger/wasp/client"
	"github.com/iotaledger/wasp/packages/coretypes"
	"github.com/iotaledger/wasp/packages/kv/codec"
	"github.com/iotaledger/wasp/packages/kv/dict"
	"github.com/iotaledger/wasp/packages/vm/viewcontext"
	"github.com/iotaledger/wasp/plugins/chains"
	"github.com/iotaledger/wasp/plugins/webapi/httperrors"
	"github.com/labstack/echo"
)

func AddEndpoints(server *echo.Echo) {
	addStateQueryEndpoint(server)
	server.GET("/"+client.StateViewRoute(":contractID", ":fname"), handleStateView)
}

func handleStateView(c echo.Context) error {
	contractID, err := coretypes.NewContractIDFromBase58(c.Param("contractID"))
	if err != nil {
		return httperrors.BadRequest(fmt.Sprintf("Invalid contract ID: %+v", c.Param("contractID")))
	}

	fname := c.Param("fname")

	params := dict.New()
	if err = c.Bind(&params); err != nil {
		return httperrors.BadRequest("Invalid request body")
	}

	chain := chains.GetChain(contractID.ChainID())
	if chain == nil {
		return httperrors.NotFound(fmt.Sprintf("Chain not found: %s", contractID.ChainID()))
	}

	vctx, err := viewcontext.New(chain)
	if err != nil {
		return fmt.Errorf(fmt.Sprintf("Failed to create context: %v", err))
	}

	ret, err := vctx.CallView(contractID.Hname(), coretypes.Hn(fname), codec.NewCodec(params))
	if err != nil {
		return httperrors.BadRequest(fmt.Sprintf("View call failed: %v", err))
	}

	// convert return value to Dict which can be marshalled into json
	var d dict.Dict
	if ret == nil {
		d = dict.New()
	} else {
		d, err = dict.FromKVStore(ret.KVStore())
		if err != nil {
			return err
		}
	}

	return c.JSON(http.StatusOK, d)
}
