package admapi

import (
	"fmt"
	"net/http"

	"github.com/iotaledger/wasp/packages/coretypes"
	"github.com/iotaledger/wasp/packages/registry"
	"github.com/iotaledger/wasp/packages/webapi/httperrors"
	"github.com/iotaledger/wasp/packages/webapi/routes"
	"github.com/iotaledger/wasp/plugins/chains"
	"github.com/labstack/echo/v4"
	"github.com/pangpanglabs/echoswagger/v2"
)

func addChainEndpoints(adm echoswagger.ApiGroup) {
	adm.POST(routes.ActivateChain(":chainID"), handleActivateChain).
		AddParamPath("", "chainID", "ChainID (base58)").
		SetSummary("Activate a chain")

	adm.POST(routes.DeactivateChain(":chainID"), handleDeactivateChain).
		AddParamPath("", "chainID", "ChainID (base58)").
		SetSummary("Deactivate a chain")
}

func handleActivateChain(c echo.Context) error {
	scAddress, err := address.FromBase58(c.Param("chainID"))
	if err != nil {
		return httperrors.BadRequest(fmt.Sprintf("Invalid SC address: %s", c.Param("address")))
	}
	chainID := (coretypes.ChainID)(scAddress)
	bd, err := registry.ActivateChainRecord(&chainID)
	if err != nil {
		return err
	}

	log.Debugw("calling committees.Activate", "chainID", bd.ChainID.String())
	if err := chains.ActivateChain(bd); err != nil {
		return err
	}

	return c.NoContent(http.StatusOK)
}

func handleDeactivateChain(c echo.Context) error {
	scAddress, err := address.FromBase58(c.Param("chainID"))
	if err != nil {
		return httperrors.BadRequest(fmt.Sprintf("Invalid chain id: %s", c.Param("chainID")))
	}

	chainID := (coretypes.ChainID)(scAddress)
	bd, err := registry.DeactivateChainRecord(&chainID)
	if err != nil {
		return err
	}

	err = chains.DeactivateChain(bd)
	if err != nil {
		return err
	}

	return c.NoContent(http.StatusOK)
}
