package blob

import (
	"fmt"
	"net/http"

	"github.com/iotaledger/wasp/packages/hashing"
	"github.com/iotaledger/wasp/packages/registry"
	"github.com/iotaledger/wasp/packages/webapi/httperrors"
	"github.com/iotaledger/wasp/packages/webapi/model"
	"github.com/iotaledger/wasp/packages/webapi/routes"
	"github.com/labstack/echo/v4"
	"github.com/pangpanglabs/echoswagger/v2"
)

type BlobCacheProvider func() registry.BlobCache

func AddEndpoints(server echoswagger.ApiRouter, provider BlobCacheProvider) {
	example := model.NewBlobInfo(true, hashing.RandomHash(nil))

	b := &blobWebAPI{provider}

	server.POST(routes.PutBlob(), b.handlePutBlob).
		SetSummary("Upload a blob to the blob cache").
		AddParamBody(model.BlobData{Data: "base64 string"}, "Blob data", "Blob data", true).
		AddResponse(http.StatusOK, "Blob properties", example, nil)

	server.GET(routes.GetBlob(":hash"), b.handleGetBlob).
		AddParamPath("", "hash", "Blob hash (base64)").
		SetSummary("Fetch a blob from the blob cache").
		AddResponse(http.StatusOK, "Blob data", model.NewBlobData([]byte("blob content")), nil).
		AddResponse(http.StatusNotFound, "Not found", httperrors.NotFound("Not found"), nil)

	server.GET(routes.HasBlob(":hash"), b.handleHasBlob).
		AddParamPath("", "hash", "Blob hash (base64)").
		SetSummary("Find out if a blob exists in the blob cache").
		AddResponse(http.StatusOK, "Blob properties", example, nil)
}

type blobWebAPI struct {
	getCache BlobCacheProvider
}

func (b *blobWebAPI) handlePutBlob(c echo.Context) error {
	var req model.BlobData
	if err := c.Bind(&req); err != nil {
		return httperrors.BadRequest(err.Error())
	}
	hash, err := b.getCache().PutBlob(req.Data.Bytes())
	if err != nil {
		return err
	}
	return c.JSON(http.StatusOK, model.NewBlobInfo(true, hash))
}

func (b *blobWebAPI) handleGetBlob(c echo.Context) error {
	hash, err := hashing.HashValueFromBase58(c.Param("hash"))
	if err != nil {
		return httperrors.BadRequest(fmt.Sprintf("Invalid hash: %q", c.Param("hash")))
	}
	data, ok, err := b.getCache().GetBlob(hash)
	if err != nil {
		return err
	}
	if !ok {
		return httperrors.NotFound(fmt.Sprintf("Blob not found: %s", hash.String()))
	}
	return c.JSON(http.StatusOK, model.NewBlobData(data))
}

func (b *blobWebAPI) handleHasBlob(c echo.Context) error {
	hash, err := hashing.HashValueFromBase58(c.Param("hash"))
	if err != nil {
		return httperrors.BadRequest(fmt.Sprintf("Invalid hash: %q", c.Param("hash")))
	}
	ok, err := b.getCache().HasBlob(hash)
	if err != nil {
		return err
	}
	return c.JSON(http.StatusOK, model.NewBlobInfo(ok, hash))
}
