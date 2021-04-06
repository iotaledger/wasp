package dashboard

import (
	_ "embed"
	"net/http"

	"github.com/iotaledger/wasp/packages/registry"
	"github.com/iotaledger/wasp/plugins/chains"
	"github.com/labstack/echo/v4"
)

//go:embed templates/chainlist.tmpl
var tplChainList string

func initChainList(e *echo.Echo, r renderer) Tab {
	route := e.GET("/chains", handleChainList)
	route.Name = "chainList"

	r[route.Path] = makeTemplate(e, tplChainList)

	return Tab{
		Path:  route.Path,
		Title: "Chains",
		Href:  route.Path,
	}
}

func handleChainList(c echo.Context) error {
	chains, err := fetchChains()
	if err != nil {
		return err
	}
	return c.Render(http.StatusOK, c.Path(), &ChainListTemplateParams{
		BaseTemplateParams: BaseParams(c),
		Chains:             chains,
	})
}

func fetchChains() ([]*ChainOverview, error) {
	crs, err := registry.GetChainRecords()
	if err != nil {
		return nil, err
	}
	r := make([]*ChainOverview, len(crs))
	for i, cr := range crs {
		info, err := fetchRootInfo(chains.AllChains().Get(&cr.ChainID))
		r[i] = &ChainOverview{
			ChainRecord: cr,
			RootInfo:    info,
			Error:       err,
		}
	}
	return r, nil
}

type ChainListTemplateParams struct {
	BaseTemplateParams
	Chains []*ChainOverview
}

type ChainOverview struct {
	ChainRecord *registry.ChainRecord
	RootInfo    RootInfo
	Error       error
}
