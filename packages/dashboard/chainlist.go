package dashboard

import (
	"net/http"

	"github.com/iotaledger/wasp/packages/registry"
	"github.com/iotaledger/wasp/plugins/chains"
	"github.com/labstack/echo"
)

const chainListRoute = "/chains"
const chainListTplName = "chainList"

func addChainListEndpoints(e *echo.Echo) {
	e.GET(chainListRoute, func(c echo.Context) error {
		chains, err := fetchChains()
		if err != nil {
			return err
		}
		return c.Render(http.StatusOK, chainListTplName, &ChainListTemplateParams{
			BaseTemplateParams: BaseParams(c, chainListRoute),
			Chains:             chains,
		})
	})
}

func fetchChains() ([]*ChainOverview, error) {
	r := make([]*ChainOverview, 0)
	crs, err := registry.GetChainRecords()
	if err != nil {
		return nil, err
	}
	for _, cr := range crs {
		info, err := fetchRootInfo(chains.GetChain(cr.ChainID))
		if err != nil {
			return nil, err
		}
		r = append(r, &ChainOverview{
			ChainRecord: cr,
			RootInfo:    info,
		})
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
}

const tplChainList = `
{{define "title"}}Chains{{end}}

{{define "body"}}
	<h2>Chains</h2>
	<table>
		<thead>
			<tr>
				<th>ID</th>
				<th>Description</th>
				<th>#Nodes</th>
				<th>#Contracts</th>
				<th>Status</th>
				<th></th>
			</tr>
		</thead>
		<tbody>
		{{range $_, $c := .Chains}}
			<tr>
				<td><abbr title="{{$c.ChainRecord.ChainID.String}}"><code>{{printf "%.8s" $c.ChainRecord.ChainID}}…</code></abbr></td>
				<td>{{printf "%.50s" $c.RootInfo.Description}}</td>
				<td>{{len $c.ChainRecord.CommitteeNodes}}</td>
				<td>{{len $c.RootInfo.Contracts}}</td>
				<td>{{if $c.ChainRecord.Active}}active{{else}}inactive{{end}}</td>
				<td><a href="/chains/{{$c.ChainRecord.ChainID}}">Details</a></td>
			</tr>
		{{end}}
		</tbody>
	</table>
{{end}}
`
