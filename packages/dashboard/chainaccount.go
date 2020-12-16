package dashboard

import (
	"net/http"

	"github.com/iotaledger/goshimmer/dapps/valuetransfers/packages/balance"
	"github.com/iotaledger/wasp/packages/coretypes"
	"github.com/iotaledger/wasp/packages/kv/codec"
	"github.com/iotaledger/wasp/packages/vm/builtinvm/accountsc"
	"github.com/iotaledger/wasp/plugins/chains"
	"github.com/labstack/echo"
)

const chainAccountRoute = "/chains/:chainid/account/:agentid"
const chainAccountTplName = "chainAccount"

func addChainAccountEndpoints(e *echo.Echo) {
	e.GET(chainAccountRoute, func(c echo.Context) error {
		chainID, err := coretypes.NewChainIDFromBase58(c.Param("chainid"))
		if err != nil {
			return err
		}

		agentID, err := coretypes.NewAgentIDFromString(c.Param("agentid"))
		if err != nil {
			return err
		}

		result := &ChainAccountTemplateParams{
			BaseTemplateParams: BaseParams(c, chainListRoute),
			ChainID:            chainID,
			AgentID:            agentID,
		}

		chain := chains.GetChain(chainID)
		if chain != nil {
			bal, err := callView(chain, accountsc.Interface.Hname(), accountsc.FuncBalance, codec.MakeDict(map[string]interface{}{
				accountsc.ParamAgentID: codec.EncodeAgentID(agentID),
			}))
			if err != nil {
				return err
			}
			result.Balances, err = accountsc.DecodeBalances(bal)
			if err != nil {
				return err
			}
		}

		return c.Render(http.StatusOK, chainAccountTplName, result)
	})
}

type ChainAccountTemplateParams struct {
	BaseTemplateParams

	ChainID coretypes.ChainID
	AgentID coretypes.AgentID

	Balances map[balance.Color]int64
}

const tplChainAccount = `
{{define "title"}}On-chain account details{{end}}

{{define "body"}}
	<div>
		<p>ChainID: <code>{{.ChainID}}</code></p>
	</div>
	<h3>On-chain account <tt>{{.AgentID}}</tt></h3>
	{{if .Balances}}
		<div>
			<h4>Balances</h4>
			{{ template "balances" .Balances }}
		</div>
	{{else}}
		<p>Not found.</p>
	{{end}}
{{end}}
`
