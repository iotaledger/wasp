package blob

import (
	"fmt"
	"net/http"

	"github.com/iotaledger/wasp/packages/hashing"
	"github.com/iotaledger/wasp/packages/webapi/httperrors"
	"github.com/iotaledger/wasp/packages/webapi/model"
	"github.com/iotaledger/wasp/packages/webapi/routes"
	"github.com/iotaledger/wasp/plugins/registry"
	"github.com/labstack/echo/v4"
	"github.com/pangpanglabs/echoswagger/v2"
)

func AddEndpoints(server echoswagger.ApiRouter) {
	example := model.NewBlobInfo(true, hashing.RandomHash(nil))

	server.GET(routes.PutBlob(), handlePutBlob).
		SetSummary("Upload a blob to the registry").
		AddResponse(http.StatusOK, "Blob properties", example, nil)

	server.GET(routes.GetBlob(":hash"), handleGetBlob).
		AddParamPath("", "hash", "Blob hash (base64)").
		SetSummary("Fetch a blob by its hash").
		AddResponse(http.StatusOK, "Blob data", model.NewBlobData([]byte("blob content")), nil).
		AddResponse(http.StatusNotFound, "Not found", httperrors.NotFound("Not found"), nil)

	server.GET(routes.HasBlob(":hash"), handleHasBlob).
		AddParamPath("", "hash", "Blob hash (base64)").
		SetSummary("Find out if a blob exists in the registry").
		AddResponse(http.StatusOK, "Blob properties", example, nil)
}

func handlePutBlob(c echo.Context) error {
	var req model.BlobData
	if err := c.Bind(&req); err != nil {
		return httperrors.BadRequest(err.Error())
	}
	hash, err := registry.DefaultRegistry().PutBlob(req.Data.Bytes())
	if err != nil {
		return err
	}
	return c.JSON(http.StatusOK, model.NewBlobInfo(true, hash))
}

func handleGetBlob(c echo.Context) error {
	hash, err := hashing.HashValueFromBase58(c.Param("hash"))
	if err != nil {
		return httperrors.BadRequest("Invalid hash")
	}
	data, ok, err := registry.DefaultRegistry().GetBlob(hash)
	if err != nil {
		return err
	}
	if !ok {
		return httperrors.NotFound(fmt.Sprintf("Blob not found: %s", hash.String()))
	}
	return c.JSON(http.StatusOK, model.NewBlobData(data))
}

func handleHasBlob(c echo.Context) error {
	hash, err := hashing.HashValueFromBase58(c.Param("hash"))
	if err != nil {
		return httperrors.BadRequest("Invalid hash")
	}
	ok, err := registry.DefaultRegistry().HasBlob(hash)
	if err != nil {
		return err
	}
	return c.JSON(http.StatusOK, model.NewBlobInfo(ok, hash))
}
