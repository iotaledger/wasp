package dashboard

import (
	"fmt"
	"github.com/iotaledger/goshimmer/packages/ledgerstate"
	"net/http"
	"strings"

	"github.com/iotaledger/wasp/packages/coretypes"
	"github.com/iotaledger/wasp/packages/kv/codec"
	"github.com/iotaledger/wasp/packages/vm/core/accounts"
	"github.com/iotaledger/wasp/plugins/chains"
	"github.com/labstack/echo/v4"
)

func initChainAccount(e *echo.Echo, r renderer) {
	route := e.GET("/chain/:chainid/account/:agentid", handleChainAccount)
	route.Name = "chainAccount"
	r[route.Path] = makeTemplate(e, tplChainAccount, tplWs)
}

func handleChainAccount(c echo.Context) error {
	chainID, err := coretypes.NewChainIDFromBase58(c.Param("chainid"))
	if err != nil {
		return err
	}

	agentID, err := coretypes.NewAgentIDFromString(strings.Replace(c.Param("agentid"), ":", "/", 1))
	if err != nil {
		return err
	}

	result := &ChainAccountTemplateParams{
		BaseTemplateParams: BaseParams(c, chainBreadcrumb(c.Echo(), chainID), Tab{
			Path:  c.Path(),
			Title: fmt.Sprintf("Account %.8s…", agentID),
			Href:  "#",
		}),
		ChainID: chainID,
		AgentID: agentID,
	}

	chain := chains.GetChain(chainID)
	if chain != nil {
		bal, err := callView(chain, accounts.Interface.Hname(), accounts.FuncBalance, codec.MakeDict(map[string]interface{}{
			accounts.ParamAgentID: codec.EncodeAgentID(agentID),
		}))
		if err != nil {
			return err
		}
		result.Balances, err = accounts.DecodeBalances(bal)
		if err != nil {
			return err
		}
	}

	return c.Render(http.StatusOK, c.Path(), result)
}

type ChainAccountTemplateParams struct {
	BaseTemplateParams

	ChainID coretypes.ChainID
	AgentID coretypes.AgentID

	Balances map[ledgerstate.Color]uint64
}

const tplChainAccount = `
{{define "title"}}On-chain account details{{end}}

{{define "body"}}
	{{if .Inputs}}
		<div class="card fluid">
			<h2 class="section">On-chain account</h2>
			<dl>
				<dt>AgentID</dt><dd><tt>{{.AgentID}}</tt></dd>
			</dl>
		</div>
		<div class="card fluid">
			<h3 class="section">Inputs</h3>
			{{ template "balances" .Inputs }}
		</div>
		{{ template "ws" .ChainID }}
	{{else}}
		<div class="card fluid error">Not found.</div>
	{{end}}
{{end}}
`
