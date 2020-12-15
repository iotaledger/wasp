package dashboard

import (
	"fmt"

	"github.com/iotaledger/wasp/packages/chain"
	"github.com/iotaledger/wasp/packages/coretypes"
	"github.com/iotaledger/wasp/packages/kv/dict"
	"github.com/iotaledger/wasp/packages/vm/viewcontext"
	"github.com/labstack/echo"
)

type chainsNavPage struct{}

func (n *chainsNavPage) Title() string { return "Chains" }
func (n *chainsNavPage) Href() string  { return chainListRoute }

func (n *chainsNavPage) AddTemplates(r renderer) {
	r[chainListTplName] = MakeTemplate(tplChainList)
	r[chainTplName] = MakeTemplate(tplChain)
}

func (n *chainsNavPage) AddEndpoints(e *echo.Echo) {
	addChainListEndpoints(e)
	addChainEndpoints(e)
}

func callView(chain chain.Chain, hname coretypes.Hname, fname string, params dict.Dict) (dict.Dict, error) {
	vctx, err := viewcontext.NewFromDB(*chain.ID(), chain.Processors())
	if err != nil {
		return nil, fmt.Errorf(fmt.Sprintf("Failed to create context: %v", err))
	}

	ret, err := vctx.CallView(hname, coretypes.Hn(fname), nil)
	if err != nil {
		return nil, fmt.Errorf("root view call failed: %v", err)
	}
	return ret, nil
}
