package dashboard

import (
	_ "embed"
	"fmt"
	"net/http"

	"github.com/iotaledger/wasp/packages/iscp"
	"github.com/iotaledger/wasp/packages/kv/codec"
	"github.com/iotaledger/wasp/packages/vm/core/accounts"
	"github.com/labstack/echo/v4"
)

//go:embed templates/chainaccount.tmpl
var tplChainAccount string

func (d *Dashboard) initChainAccount(e *echo.Echo, r renderer) {
	route := e.GET("/chain/:chainid/account/:agentid", d.handleChainAccount)
	route.Name = "chainAccount"
	r[route.Path] = d.makeTemplate(e, tplChainAccount, tplWebSocket)
}

func (d *Dashboard) handleChainAccount(c echo.Context) error {
	chainID, err := iscp.ChainIDFromString(c.Param("chainid"))
	if err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, err)
	}

	agentID, err := iscp.NewAgentIDFromString(c.Param("agentid"), d.wasp.L1Params().Protocol.Bech32HRP)
	if err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, err)
	}

	result := &ChainAccountTemplateParams{
		BaseTemplateParams: d.BaseParams(c, chainBreadcrumb(c.Echo(), chainID), Tab{
			Path:  c.Path(),
			Title: fmt.Sprintf("Account %.16s…", agentID.String(d.wasp.L1Params().Protocol.Bech32HRP)),
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
	result.Balances, err = iscp.FungibleTokensFromDict(bal)
	if err != nil {
		return err
	}

	return c.Render(http.StatusOK, c.Path(), result)
}

type ChainAccountTemplateParams struct {
	BaseTemplateParams

	ChainID *iscp.ChainID
	AgentID iscp.AgentID

	Balances *iscp.FungibleTokens
}
