package dashboard

import (
	"github.com/labstack/echo"
)

type chainsNavPage struct{}

func (n *chainsNavPage) Title() string { return "Chains" }
func (n *chainsNavPage) Href() string  { return chainListRoute }

func (n *chainsNavPage) AddTemplates(r renderer) {
	r[chainListTplName] = MakeTemplate(tplChainList)
	r[chainTplName] = MakeTemplate(tplChain)
	r[chainAccountTplName] = MakeTemplate(tplChainAccount)
	r[chainBlobTplName] = MakeTemplate(tplChainBlob)
}

func (n *chainsNavPage) AddEndpoints(e *echo.Echo) {
	addChainListEndpoints(e)
	addChainEndpoints(e)
	addChainAccountEndpoints(e)
	addChainBlobEndpoints(e)
}
