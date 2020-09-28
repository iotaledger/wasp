package admapi

import (
	"github.com/iotaledger/goshimmer/dapps/valuetransfers/packages/address"
	"github.com/iotaledger/wasp/packages/registry"
	"github.com/iotaledger/wasp/plugins/committees"
	"github.com/iotaledger/wasp/plugins/webapi/misc"
	"github.com/labstack/echo"
	"net/http"
)

func HandlerDeactivateSC(c echo.Context) error {
	scAddress, err := address.FromBase58(c.Param("scaddress"))
	if err != nil {
		return c.JSON(http.StatusBadRequest, &misc.SimpleResponse{Error: err.Error()})
	}

	bd, err := registry.DeactivateBootupData(&scAddress)
	if err != nil {
		return c.JSON(http.StatusBadRequest, &misc.SimpleResponse{Error: err.Error()})
	}

	err = committees.DeactivateCommittee(bd)
	if err != nil {
		return c.JSON(http.StatusBadRequest, &misc.SimpleResponse{Error: err.Error()})
	}

	return c.JSON(http.StatusOK, &misc.SimpleResponse{})
}
