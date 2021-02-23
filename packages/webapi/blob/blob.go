package blob

import (
	"fmt"
	"net/http"
	"strings"

	"github.com/iotaledger/wasp/packages/downloader"
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

	server.GET(routes.DownloadBlob(), handleDownloadBlob).
		AddParamPath("", "uri", "Uri of the file to download").
		SetSummary("Download the blob from the uri and put it in the registry").
		AddResponse(http.StatusOK, "Blob data", example, nil).
		AddResponse(http.StatusNotFound, "Not found", httperrors.NotFound("Not found"), nil)
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

func handleDownloadBlob(c echo.Context) error {
	var uri string = c.QueryParam("uri")
	var split []string = strings.SplitN(uri, "://", 2)
	if len(split) != 2 {
		return httperrors.BadRequest("Illegal uri " + uri)
	}

	var protocol string = split[0]
	var path string = split[1]
	switch protocol {
	case "ipfs":
		return donwloadBlobFromHttp(c, "https://ipfs.io/ipfs/"+path)
	case "http":
		return donwloadBlobFromHttp(c, uri)
	default:
		return httperrors.BadRequest("Unsupported uri protocol " + protocol)
	}
}

func donwloadBlobFromHttp(c echo.Context, url string) error {
	var hash hashing.HashValue
	var err error
	hash, err = downloader.DonwloadBlobFromHttp(url)
	if err != nil {
		return err
	}
	return c.JSON(http.StatusOK, model.NewBlobInfo(true, hash))
}
