package admapi

import (
	"fmt"
	"net/http"

	"github.com/iotaledger/goshimmer/packages/ledgerstate"
	"github.com/iotaledger/wasp/packages/chains"
	"github.com/iotaledger/wasp/packages/iscp"
	"github.com/iotaledger/wasp/packages/registry"
	"github.com/iotaledger/wasp/packages/webapi/httperrors"
	"github.com/iotaledger/wasp/packages/webapi/routes"
	"github.com/labstack/echo/v4"
	"github.com/pangpanglabs/echoswagger/v2"
)

func addChainEndpoints(adm echoswagger.ApiGroup, registryProvider registry.Provider, chainsProvider chains.Provider) {
	c := &chainWebAPI{registryProvider, chainsProvider}

	adm.POST(routes.ActivateChain(":chainID"), c.handleActivateChain).
		AddParamPath("", "chainID", "ChainID (base58)").
		SetSummary("Activate a chain")

	adm.POST(routes.DeactivateChain(":chainID"), c.handleDeactivateChain).
		AddParamPath("", "chainID", "ChainID (base58)").
		SetSummary("Deactivate a chain")
}

type chainWebAPI struct {
	registry registry.Provider
	chains   chains.Provider
}

func (w *chainWebAPI) handleActivateChain(c echo.Context) error {
	aliasAddress, err := ledgerstate.AliasAddressFromBase58EncodedString(c.Param("chainID"))
	if err != nil {
		return httperrors.BadRequest(fmt.Sprintf("Invalid alias address: %s", c.Param("chainID")))
	}
	chainID, err := iscp.ChainIDFromAddress(aliasAddress)
	if err != nil {
		return err
	}
	rec, err := w.registry().ActivateChainRecord(chainID)
	if err != nil {
		return err
	}

	log.Debugw("calling Chains.Activate", "chainID", rec.ChainID.String())
	if err := w.chains().Activate(rec, w.registry, nil); err != nil {
		return err
	}

	return c.NoContent(http.StatusOK)
}

func (w *chainWebAPI) handleDeactivateChain(c echo.Context) error {
	scAddress, err := ledgerstate.AddressFromBase58EncodedString(c.Param("chainID"))
	if err != nil {
		return httperrors.BadRequest(fmt.Sprintf("Invalid chain id: %s", c.Param("chainID")))
	}
	chainID, err := iscp.ChainIDFromAddress(scAddress)
	if err != nil {
		return err
	}
	bd, err := w.registry().DeactivateChainRecord(chainID)
	if err != nil {
		return err
	}

	err = w.chains().Deactivate(bd)
	if err != nil {
		return err
	}

	return c.NoContent(http.StatusOK)
}
