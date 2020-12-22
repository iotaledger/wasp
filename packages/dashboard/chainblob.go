package dashboard

import (
	"fmt"
	"net/http"

	"github.com/iotaledger/wasp/packages/coretypes"
	"github.com/iotaledger/wasp/packages/hashing"
	"github.com/iotaledger/wasp/packages/kv/codec"
	"github.com/iotaledger/wasp/packages/vm/builtinvm/blob"
	"github.com/iotaledger/wasp/plugins/chains"
	"github.com/iotaledger/wasp/plugins/webapi/httperrors"
	"github.com/labstack/echo"
	"github.com/mr-tron/base58"
)

const chainBlobRoute = "/chains/:chainid/blob/:hash"
const chainBlobTplName = "chainBlob"

func chainBlobRawRoute(chainID string, hash string, field string) string {
	return fmt.Sprintf("/chains/%s/blob/%s/raw/%s", chainID, hash, field)
}

func addChainBlobEndpoints(e *echo.Echo) {
	e.GET(chainBlobRoute, func(c echo.Context) error {
		chainID, err := coretypes.NewChainIDFromBase58(c.Param("chainid"))
		if err != nil {
			return err
		}

		hash, err := hashing.HashValueFromBase58(c.Param("hash"))
		if err != nil {
			return err
		}

		result := &ChainBlobTemplateParams{
			BaseTemplateParams: BaseParams(c, chainListRoute),
			ChainID:            chainID,
			Hash:               hash,
		}

		chain := chains.GetChain(chainID)
		if chain != nil {
			fields, err := callView(chain, blob.Interface.Hname(), blob.FuncGetBlobInfo, codec.MakeDict(map[string]interface{}{
				blob.ParamHash: hash,
			}))
			if err != nil {
				return err
			}
			result.Blob = map[string]BlobValue{}
			for field := range fields {
				value, err := callView(chain, blob.Interface.Hname(), blob.FuncGetBlobField, codec.MakeDict(map[string]interface{}{
					blob.ParamHash:  hash,
					blob.ParamField: []byte(field),
				}))
				if err != nil {
					return err
				}
				bytes := value[blob.ParamBytes]
				result.Blob[string(field)] = BlobValue{
					Len:     len(bytes),
					Encoded: string(bytes),
					RawHref: chainBlobRawRoute(c.Param("chainid"), c.Param("hash"), base58.Encode([]byte(field))),
				}
			}
		}

		return c.Render(http.StatusOK, chainBlobTplName, result)
	})

	e.GET(chainBlobRawRoute(":chainid", ":hash", ":field"), func(c echo.Context) error {
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
	})
}

type ChainBlobTemplateParams struct {
	BaseTemplateParams

	ChainID coretypes.ChainID
	Hash    hashing.HashValue

	Blob map[string]BlobValue
}

type BlobValue struct {
	Len     int
	Encoded string
	RawHref string
}

const tplChainBlob = `
{{define "title"}}Blob details{{end}}

{{define "body"}}
	{{if .Blob}}
		<div class="card fluid">
			<h3>Blob <tt>{{hashref .Hash}}</tt></h3>
			<dl>
				<dt>ChainID</dt><dd><tt>{{.ChainID}}</tt></dd>
			</dl>
			<table>
				<thead>
					<tr>
						<th>Field</th>
						<th>Value</th>
						<th class="align-right">Size (bytes)</th>
						<th></th>
					</tr>
				</thead>
				<tbody>
				{{range $field, $value := .Blob}}
					<tr>
						<td><tt>{{ quoted 30 $field }}</tt></td>
						<td><pre style="white-space: pre-wrap; max-width: 400px">{{ quoted 50 $value.Encoded }}</pre></td>
						<td class="align-right">{{ $value.Len }}</td>
						<td><a href="{{ $value.RawHref }}">Download</a></td>
					</tr>
				{{end}}
				</tbody>
			</table>
		</div>
	{{else}}
		<div class="card error">Not found.</div>
	{{end}}
{{end}}
`
