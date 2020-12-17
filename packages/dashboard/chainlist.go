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
	crs, err := registry.GetChainRecords()
	if err != nil {
		return nil, err
	}
	r := make([]*ChainOverview, len(crs))
	for i, cr := range crs {
		info, err := fetchRootInfo(chains.GetChain(cr.ChainID))
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

const tplChainList = `
{{define "title"}}Chains{{end}}

{{define "body"}}
	<h2>Chains</h2>
	{{range $_, $c := .Chains}}
		{{ $desc := printf "%.50s" $c.RootInfo.Description }}
		{{ $id := $c.ChainRecord.ChainID }}
		<details open="">
			<summary>{{ if $desc }}{{ $desc }}{{ else }}<tt>{{ $c.ChainRecord.ChainID }}</tt>{{ end }}</summary>
			<p>ChainID: <code>{{ $id }}</code></p>
			{{ if $c.Error }}
				<p><b>Error: {{ $c.Error }}</b></p>
			{{ else if not $c.ChainRecord.Active }}
				<p>This chain is <b>inactive</b>.</p>
			{{ else }}
				<p>Committee: <code>{{ $c.ChainRecord.CommitteeNodes }}</code></p>
				<p>#Contracts: <code>{{ len $c.RootInfo.Contracts }}</code></p>
				<p><a href="/chains/{{$c.ChainRecord.ChainID}}">Chain dashboard</a></p>
			{{ end }}
		</details>
	{{end}}
	</table>
{{end}}
`
