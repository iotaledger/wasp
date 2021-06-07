package dashboard

import (
	_ "embed"
	"net/http"

	"github.com/iotaledger/wasp/packages/registry_pkg/chainrecord"

	"github.com/labstack/echo/v4"
)

//go:embed templates/chainlist.tmpl
var tplChainList string

func (d *Dashboard) initChainList(e *echo.Echo, r renderer) Tab {
	route := e.GET("/chains", d.handleChainList)
	route.Name = "chainList"

	r[route.Path] = d.makeTemplate(e, tplChainList)

	return Tab{
		Path:  route.Path,
		Title: "Chains",
		Href:  route.Path,
	}
}

func (d *Dashboard) handleChainList(c echo.Context) error {
	chains, err := d.fetchChains()
	if err != nil {
		return err
	}
	return c.Render(http.StatusOK, c.Path(), &ChainListTemplateParams{
		BaseTemplateParams: d.BaseParams(c),
		Chains:             chains,
	})
}

func (d *Dashboard) fetchChains() ([]*ChainOverview, error) {
	crs, err := d.wasp.GetChainRecords()
	if err != nil {
		return nil, err
	}
	r := make([]*ChainOverview, len(crs))
	for i, cr := range crs {
		info, err := d.fetchRootInfo(d.wasp.GetChain(cr.ChainID))
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
	ChainRecord *chainrecord.ChainRecord
	RootInfo    RootInfo
	Error       error
}
