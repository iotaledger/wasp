package admapi

import (
	"fmt"
	"github.com/iotaledger/wasp/packages/coretypes"
	"net/http"

	"github.com/iotaledger/goshimmer/dapps/valuetransfers/packages/address"
	"github.com/iotaledger/wasp/client"
	"github.com/iotaledger/wasp/packages/registry"
	"github.com/iotaledger/wasp/plugins/chains"
	"github.com/iotaledger/wasp/plugins/webapi/httperrors"
	"github.com/labstack/echo"
)

func addSCEndpoints(adm *echo.Group) {
	adm.POST("/"+client.ActivateChainRoute(":chainid"), handleActivateChain)
	adm.POST("/"+client.DeactivateChainRoute(":chainid"), handleDeactivateChain)
}

func handleActivateChain(c echo.Context) error {
	scAddress, err := address.FromBase58(c.Param("chainid"))
	if err != nil {
		return httperrors.BadRequest(fmt.Sprintf("Invalid SC address: %s", c.Param("address")))
	}
	chainID := (coretypes.ChainID)(scAddress)
	bd, err := registry.ActivateBootupData(&chainID)
	if err != nil {
		return err
	}

	log.Debugw("calling committees.ActivateChain", "chainid", bd.ChainID.String())
	if err := chains.ActivateChain(bd); err != nil {
		return err
	}

	return c.NoContent(http.StatusOK)
}

func handleDeactivateChain(c echo.Context) error {
	scAddress, err := address.FromBase58(c.Param("chainid"))
	if err != nil {
		return httperrors.BadRequest(fmt.Sprintf("Invalid chain id: %s", c.Param("chainid")))
	}

	chainID := (coretypes.ChainID)(scAddress)
	bd, err := registry.DeactivateBootupData(&chainID)
	if err != nil {
		return err
	}

	err = chains.DeactivateChain(bd)
	if err != nil {
		return err
	}

	return c.NoContent(http.StatusOK)
}
