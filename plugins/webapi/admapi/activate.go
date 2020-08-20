package admapi

import (
	"github.com/iotaledger/goshimmer/dapps/valuetransfers/packages/address"
	"github.com/iotaledger/wasp/packages/registry"
	"github.com/iotaledger/wasp/plugins/committees"
	"github.com/iotaledger/wasp/plugins/webapi/misc"
	"github.com/labstack/echo"
	"net/http"
)

type ActivateSCRequest struct {
	Address string `json:"address"` //base58
}

func HandlerActivateSC(c echo.Context) error {
	var req ActivateSCRequest
	if err := c.Bind(&req); err != nil {
		return c.JSON(http.StatusOK, &misc.SimpleResponse{Error: err.Error()})
	}
	addr, err := address.FromBase58(req.Address)
	if err != nil {
		return c.JSON(http.StatusOK, &misc.SimpleResponse{Error: err.Error()})
	}

	bd, exist, err := registry.GetBootupData(&addr)

	if err != nil {
		return c.JSON(http.StatusOK, &misc.SimpleResponse{Error: err.Error()})
	}
	if !exist {
		return c.JSON(http.StatusOK, &misc.SimpleResponse{Error: "address not found"})
	}

	log.Debugw("calling committees.ActivateCommittee", "addr", bd.Address.String())

	committees.ActivateCommittee(bd)

	return c.JSON(http.StatusOK, &misc.SimpleResponse{})
}
