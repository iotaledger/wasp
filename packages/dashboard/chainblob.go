package dashboard

import (
	"fmt"
	"net/http"

	"github.com/iotaledger/wasp/packages/coretypes"
	"github.com/iotaledger/wasp/packages/hashing"
	"github.com/iotaledger/wasp/packages/kv/codec"
	"github.com/iotaledger/wasp/packages/vm/core/blob"
	"github.com/iotaledger/wasp/packages/webapi/httperrors"
	"github.com/iotaledger/wasp/plugins/chains"
	"github.com/labstack/echo/v4"
	"github.com/mr-tron/base58"
)

func initChainBlob(e *echo.Echo, r renderer) {
	route := e.GET("/chain/:chainid/blob/:hash", handleChainBlob)
	route.Name = "chainBlob"
	r[route.Path] = makeTemplate(e, tplChainBlob, tplWs)

	route = e.GET("/chain/:chainid/blob/:hash/raw/:field", handleChainBlobDownload)
	route.Name = "chainBlobDownload"
}

func handleChainBlob(c echo.Context) error {
	chainID, err := coretypes.NewChainIDFromBase58(c.Param("chainid"))
	if err != nil {
		return err
	}

	hash, err := hashing.HashValueFromBase58(c.Param("hash"))
	if err != nil {
		return err
	}

	result := &ChainBlobTemplateParams{
		BaseTemplateParams: BaseParams(c, chainBreadcrumb(c.Echo(), chainID), Tab{
			Path:  c.Path(),
			Title: fmt.Sprintf("Blob %.8s…", c.Param("hash")),
			Href:  "#",
		}),
		ChainID: chainID,
		Hash:    hash,
	}

	chain := chains.GetChain(chainID)
	if chain != nil {
		fields, err := callView(chain, blob.Interface.Hname(), blob.FuncGetBlobInfo, codec.MakeDict(map[string]interface{}{
			blob.ParamHash: hash,
		}))
		if err != nil {
			return err
		}
		result.Blob = []BlobField{}
		for field := range fields {
			field := []byte(field)
			value, err := callView(chain, blob.Interface.Hname(), blob.FuncGetBlobField, codec.MakeDict(map[string]interface{}{
				blob.ParamHash:  hash,
				blob.ParamField: field,
			}))
			if err != nil {
				return err
			}
			result.Blob = append(result.Blob, BlobField{
				Key:   field,
				Value: value[blob.ParamBytes],
			})
		}
	}

	return c.Render(http.StatusOK, c.Path(), result)
}

func handleChainBlobDownload(c echo.Context) error {
	chainID, err := coretypes.NewChainIDFromBase58(c.Param("chainid"))
	if err != nil {
		return err
	}

	hash, err := hashing.HashValueFromBase58(c.Param("hash"))
	if err != nil {
		return err
	}

	field, err := base58.Decode(c.Param("field"))
	if err != nil {
		return err
	}

	chain := chains.GetChain(chainID)
	if chain == nil {
		return httperrors.NotFound("Not found")
	}

	value, err := callView(chain, blob.Interface.Hname(), blob.FuncGetBlobField, codec.MakeDict(map[string]interface{}{
		blob.ParamHash:  hash,
		blob.ParamField: field,
	}))
	if err != nil {
		return err
	}

	return c.Blob(http.StatusOK, "application/octet-stream", value[blob.ParamBytes])
}

type ChainBlobTemplateParams struct {
	BaseTemplateParams

	ChainID coretypes.ChainID
	Hash    hashing.HashValue

	Blob []BlobField
}

type BlobField struct {
	Key   []byte
	Value []byte
}

const tplChainBlob = `
{{define "title"}}Blob details{{end}}

{{define "body"}}
	{{if .Blob}}
		{{ $chainid := .ChainID }}
		{{ $hash := .Hash }}

		<div class="card fluid">
			<h2 class="section">Blob</h2>
			<dl>
				<dt>Hash</dt><dd><tt>{{hashref $hash}}</tt></dd>
			</dl>
		</div>
		<div class="card fluid">
			<h4 class="section">Fields</h3>
			<table>
				<thead>
					<tr>
						<th>Field</th>
						<th style="flex: 2">Value (first 100 bytes)</th>
						<th class="align-right" style="flex: 0.5">Size (bytes)</th>
						<th style="flex: 0.5"></th>
					</tr>
				</thead>
				<tbody>
				{{range $i, $field := .Blob}}
					<tr>
						<td><tt>{{ trim 30 (bytesToString $field.Key) }}</tt></td>
						<td style="flex: 2"><pre style="white-space: pre-wrap">{{ trim 100 (bytesToString $field.Value) }}</pre></td>
						<td class="align-right" style="flex: 0.5">{{ len $field.Value }}</td>
						<td style="flex: 0.5"><a href="{{ uri "chainBlobDownload" $chainid (hashref $hash) (base58 $field.Key) }}">Download</a></td>
					</tr>
				{{end}}
				</tbody>
			</table>
		</div>
		{{ template "ws" .ChainID }}
	{{else}}
		<div class="card fluid error">Not found.</div>
	{{end}}
{{end}}
`
