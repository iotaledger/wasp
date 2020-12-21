package dashboard

import (
	"net/http"

	"github.com/iotaledger/goshimmer/dapps/valuetransfers/packages/balance"
	"github.com/iotaledger/wasp/packages/coretypes"
	"github.com/iotaledger/wasp/packages/kv/codec"
	"github.com/iotaledger/wasp/packages/vm/builtinvm/accounts"
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
<div class="container">
<div class="row">
<div class="col-sm">
	<h3>On-chain account</h3>
	<dl>
		<dt>ChainID</dt><dd><tt>{{.ChainID}}</tt></dd>
		<dt>AgentID</dt><dd><tt>{{.AgentID}}</tt></dd>
	</dl>
	{{if .Balances}}
		<div>
			<h4>Balances</h4>
			{{ template "balances" .Balances }}
		</div>
	{{else}}
		<div class="card error">Not found.</div>
	{{end}}
</div>
</div>
</div>
{{end}}
`
