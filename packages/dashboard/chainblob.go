package dashboard

import (
	_ "embed"
	"fmt"
	"net/http"

	"github.com/iotaledger/wasp/packages/hashing"
	"github.com/iotaledger/wasp/packages/iscp"
	"github.com/iotaledger/wasp/packages/kv"
	"github.com/iotaledger/wasp/packages/kv/codec"
	"github.com/iotaledger/wasp/packages/kv/dict"
	"github.com/iotaledger/wasp/packages/vm/core/blob"
	"github.com/labstack/echo/v4"
	"github.com/mr-tron/base58"
)

//go:embed templates/chainblob.tmpl
var tplChainBlob string

func (d *Dashboard) initChainBlob(e *echo.Echo, r renderer) {
	route := e.GET("/chain/:chainid/blob/:hash", d.handleChainBlob)
	route.Name = "chainBlob"
	r[route.Path] = d.makeTemplate(e, tplChainBlob, tplWebSocket)

	route = e.GET("/chain/:chainid/blob/:hash/raw/:field", d.handleChainBlobDownload)
	route.Name = "chainBlobDownload"
}

func (d *Dashboard) handleChainBlob(c echo.Context) error {
	chainID, err := iscp.ChainIDFromBase58(c.Param("chainid"))
	if err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, err)
	}

	hash, err := hashing.HashValueFromBase58(c.Param("hash"))
	if err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, err)
	}

	result := &ChainBlobTemplateParams{
		BaseTemplateParams: d.BaseParams(c, chainBreadcrumb(c.Echo(), *chainID), Tab{
			Path:  c.Path(),
			Title: fmt.Sprintf("Blob %.8s…", c.Param("hash")),
			Href:  "#",
		}),
		ChainID: *chainID,
		Hash:    hash,
	}

	fields, err := d.wasp.CallView(chainID, blob.Contract.Name, blob.FuncGetBlobInfo.Name, codec.MakeDict(map[string]interface{}{
		blob.ParamHash: hash,
	}))
	if err != nil {
		return err
	}
	result.Blob = make([]BlobField, len(fields))
	i := 0
	fields.MustIterateKeysSorted("", func(key kv.Key) bool {
		field := []byte(key)
		var value dict.Dict
		value, err = d.wasp.CallView(chainID, blob.Contract.Name, blob.FuncGetBlobField.Name, codec.MakeDict(map[string]interface{}{
			blob.ParamHash:  hash,
			blob.ParamField: field,
		}))
		if err != nil {
			return false
		}
		result.Blob[i] = BlobField{
			Key:   field,
			Value: value[blob.ParamBytes],
		}
		i++
		return true
	})
	if err != nil {
		return err
	}

	return c.Render(http.StatusOK, c.Path(), result)
}

func (d *Dashboard) handleChainBlobDownload(c echo.Context) error {
	chainID, err := iscp.ChainIDFromBase58(c.Param("chainid"))
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

	value, err := d.wasp.CallView(chainID, blob.Contract.Name, blob.FuncGetBlobField.Name, codec.MakeDict(map[string]interface{}{
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

	ChainID iscp.ChainID
	Hash    hashing.HashValue

	Blob []BlobField
}

type BlobField struct {
	Key   []byte
	Value []byte
}
