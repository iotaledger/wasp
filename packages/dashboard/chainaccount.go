package dashboard

import (
	_ "embed"
	"fmt"
	"net/http"

	"github.com/iotaledger/wasp/packages/isc"
	"github.com/iotaledger/wasp/packages/kv/codec"
	"github.com/iotaledger/wasp/packages/vm/core/accounts"
	"github.com/labstack/echo/v4"
)

//go:embed templates/chainaccount.tmpl
var tplChainAccount string

func (d *Dashboard) initChainAccount(e *echo.Echo, r renderer) {
	route := e.GET("/chain/:chainid/account/:agentid", d.handleChainAccount)
	route.Name = "chainAccount"
	r[route.Path] = d.makeTemplate(e, tplChainAccount)
}

func (d *Dashboard) handleChainAccount(c echo.Context) error {
	chainID, err := isc.ChainIDFromString(c.Param("chainid"))
	if err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, err)
	}

	agentID, err := isc.NewAgentIDFromString(c.Param("agentid"))
	if err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, err)
	}

	result := &ChainAccountTemplateParams{
		BaseTemplateParams: d.BaseParams(c, chainBreadcrumb(c.Echo(), chainID), Tab{
			Path:  c.Path(),
			Title: fmt.Sprintf("Account %.16s…", agentID.String()),
			Href:  "#",
		}),
		ChainID: chainID,
		AgentID: agentID,
	}

	bal, err := d.wasp.CallView(chainID, accounts.Contract.Name, accounts.ViewBalance.Name, codec.MakeDict(map[string]interface{}{
		accounts.ParamAgentID: codec.EncodeAgentID(agentID),
	}))
	if err != nil {
		return err
	}
	result.Balances, err = isc.FungibleTokensFromDict(bal)
	if err != nil {
		return err
	}

	return c.Render(http.StatusOK, c.Path(), result)
}

type ChainAccountTemplateParams struct {
	BaseTemplateParams

	ChainID *isc.ChainID
	AgentID isc.AgentID

	Balances *isc.FungibleTokens
}
