package admapi

import (
	"github.com/iotaledger/goshimmer/dapps/valuetransfers/packages/address"
	"github.com/iotaledger/wasp/packages/registry"
	"github.com/iotaledger/wasp/plugins/committees"
	"github.com/iotaledger/wasp/plugins/webapi/misc"
	"github.com/labstack/echo"
	"net/http"
)

func HandlerActivateSC(c echo.Context) error {
	scAddress, err := address.FromBase58(c.Param("scaddress"))
	if err != nil {
		return c.JSON(http.StatusBadRequest, &misc.SimpleResponse{Error: err.Error()})
	}

	bd, err := registry.ActivateBootupData(&scAddress)
	if err != nil {
		return c.JSON(http.StatusBadRequest, &misc.SimpleResponse{Error: err.Error()})
	}

	log.Debugw("calling committees.ActivateCommittee", "addr", bd.Address.String())

	if err := committees.ActivateCommittee(bd); err != nil {
		return c.JSON(http.StatusBadRequest, &misc.SimpleResponse{Error: err.Error()})
	}

	return c.JSON(http.StatusOK, &misc.SimpleResponse{})
}
