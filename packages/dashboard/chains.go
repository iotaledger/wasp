package dashboard

import (
	"fmt"

	"github.com/iotaledger/wasp/packages/coretypes"
	"github.com/labstack/echo"
)

type chainsNavPage struct{}

func (n *chainsNavPage) Title() string { return "Chains" }
func (n *chainsNavPage) Href() string  { return chainListRoute }

func (n *chainsNavPage) AddTemplates(r renderer) {
	r[chainListTplName] = MakeTemplate(tplChainList)
	r[chainTplName] = MakeTemplate(tplChain, tplWs)
	r[chainAccountTplName] = MakeTemplate(tplChainAccount, tplWs)
	r[chainBlobTplName] = MakeTemplate(tplChainBlob, tplWs)
	r[chainContractTplName] = MakeTemplate(tplChainContract, tplWs)
}

func (n *chainsNavPage) AddEndpoints(e *echo.Echo) {
	addChainListEndpoints(e)
	addChainEndpoints(e)
	addChainAccountEndpoints(e)
	addChainBlobEndpoints(e)
	addChainContractEndpoints(e)
}

func chainBreadcrumb(chainID coretypes.ChainID) Breadcrumb {
	return Breadcrumb{
		Title: fmt.Sprintf("Chain %.8s…", chainID),
		Href:  fmt.Sprintf("/chain/%s", chainID),
	}
}
