package dashboard

import (
	"fmt"
	"net/http"

	"github.com/iotaledger/wasp/packages/coretypes"
	"github.com/iotaledger/wasp/packages/kv/codec"
	"github.com/iotaledger/wasp/packages/kv/datatypes"
	"github.com/iotaledger/wasp/packages/vm/core/eventlog"
	"github.com/iotaledger/wasp/packages/vm/core/root"
	"github.com/iotaledger/wasp/plugins/chains"
	"github.com/labstack/echo"
)

func initChainContract(e *echo.Echo, r renderer) {
	route := e.GET("/chain/:chainid/contract/:hname", handleChainContract)
	route.Name = "chainContract"
	r[route.Path] = makeTemplate(e, tplChainContract, tplWs)
}

func handleChainContract(c echo.Context) error {
	chainID, err := coretypes.NewChainIDFromBase58(c.Param("chainid"))
	if err != nil {
		return err
	}

	hname, err := coretypes.HnameFromString(c.Param("hname"))
	if err != nil {
		return err
	}

	result := &ChainContractTemplateParams{
		BaseTemplateParams: BaseParams(c, chainBreadcrumb(c.Echo(), chainID), Tab{
			Path:  c.Path(),
			Title: fmt.Sprintf("Contract %d", hname),
			Href:  "#",
		}),
		ChainID: chainID,
		Hname:   hname,
	}

	chain := chains.GetChain(chainID)
	if chain != nil {
		r, err := callView(chain, root.Interface.Hname(), root.FuncFindContract, codec.MakeDict(map[string]interface{}{
			root.ParamHname: codec.EncodeHname(hname),
		}))
		if err != nil {
			return err
		}
		result.ContractRecord, err = root.DecodeContractRecord(r[root.ParamData])
		if err != nil {
			return err
		}

		r, err = callView(chain, eventlog.Interface.Hname(), eventlog.FuncGetLogRecords, codec.MakeDict(map[string]interface{}{
			eventlog.ParamContractHname: codec.EncodeHname(hname),
		}))
		if err != nil {
			return err
		}
		records := datatypes.NewMustArray(r, eventlog.ParamRecords)
		result.Log = make([]*datatypes.TimestampedLogRecord, records.Len())
		for i := uint16(0); i < records.Len(); i++ {
			b := records.GetAt(i)
			result.Log[i], err = datatypes.ParseRawLogRecord(b)
			if err != nil {
				return err
			}
		}

		result.RootInfo, err = fetchRootInfo(chain)
		if err != nil {
			return err
		}
	}

	return c.Render(http.StatusOK, c.Path(), result)
}

type ChainContractTemplateParams struct {
	BaseTemplateParams

	ChainID coretypes.ChainID
	Hname   coretypes.Hname

	ContractRecord *root.ContractRecord
	Log            []*datatypes.TimestampedLogRecord
	RootInfo       RootInfo
}

const tplChainContract = `
{{define "title"}}Contract details{{end}}

{{define "body"}}
	{{ $c := .ContractRecord }}
	{{ $chainid := .ChainID }}
	{{ $rootinfo := .RootInfo }}
	{{ if $c }}
		<div class="card fluid">
			<h2 class="section">Contract</h2>
			<dl>
				<dt>Name</dt><dd><tt>{{trim 50 $c.Name}}</tt></dd>
				<dt>Hname</dt><dd><tt>{{.Hname}}</tt></dd>
				<dt>Description</dt><dd><tt>{{trim 50 $c.Description}}</tt></dd>
				<dt>Program hash</dt><dd><tt>{{$c.ProgramHash.String}}</tt></dd>
				{{if $c.HasCreator}}<dt>Creator</dt><dd>{{ template "agentid" (args $chainid $c.Creator) }}</dd>{{end}}
				<dt>Owner fee</dt><dd>
					{{- if $c.OwnerFee -}}
						<tt>{{- $c.OwnerFee }} {{ $rootinfo.FeeColor -}}</tt>
					{{- else -}}
						<tt>{{- $rootinfo.DefaultOwnerFee }} {{ $rootinfo.FeeColor }}</tt> (chain default)
					{{- end -}}
				</dd>
				<dt>Validator fee</dt><dd>
					{{- if $c.ValidatorFee -}}
						<tt>{{- $c.ValidatorFee }} {{ $rootinfo.FeeColor -}}
					{{- else -}}
						<tt>{{- $rootinfo.DefaultValidatorFee }} {{ $rootinfo.FeeColor }}</tt> (chain default)
					{{- end -}}
				</dd>
			</dl>
		</div>

		<div class="card fluid">
			<h3 class="section">Log</h3>
			<dl style="align-items: center">
				{{ range $_, $rec := .Log }}
					<dt><tt>{{ formatTimestamp $rec.Timestamp }}</tt></dt>
					<dd><pre>{{- trim 1000 (bytesToString $rec.Data) -}}</pre></dd>
				{{ end }}
			</dl>
		</div>
		{{ template "ws" .ChainID }}
	{{else}}
		<div class="card fluid error">Not found.</div>
	{{end}}
</div>
</div>
</div>
{{end}}
`
