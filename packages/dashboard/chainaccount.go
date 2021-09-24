package dashboard

import (
	_ "embed"
	"fmt"
	"net/http"
	"strings"

	"github.com/iotaledger/wasp/packages/iscp"
	"github.com/iotaledger/wasp/packages/iscp/colored"
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
	chainID, err := iscp.ChainIDFromBase58(c.Param("chainid"))
	if err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, err)
	}

	agentID, err := iscp.NewAgentIDFromString(strings.Replace(c.Param("agentid"), ":", "/", 1))
	if err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, err)
	}

	result := &ChainAccountTemplateParams{
		BaseTemplateParams: d.BaseParams(c, chainBreadcrumb(c.Echo(), *chainID), Tab{
			Path:  c.Path(),
			Title: fmt.Sprintf("Account %.8s…", agentID),
			Href:  "#",
		}),
		ChainID: *chainID,
		AgentID: *agentID,
	}

	bal, err := d.wasp.CallView(chainID, accounts.Contract.Name, accounts.FuncViewBalance.Name, codec.MakeDict(map[string]interface{}{
		accounts.ParamAgentID: codec.EncodeAgentID(agentID),
	}))
	if err != nil {
		return err
	}
	result.Balances, err = accounts.DecodeBalances(bal)
	if err != nil {
		return err
	}

	return c.Render(http.StatusOK, c.Path(), result)
}

type ChainAccountTemplateParams struct {
	BaseTemplateParams

	ChainID iscp.ChainID
	AgentID iscp.AgentID

	Balances colored.Balances
}
