package dkgapi

import (
	"bytes"
	"net/http"

	"github.com/iotaledger/wasp/packages/registry"
	"github.com/labstack/echo"
	"github.com/mr-tron/base58"
)

type ExportDKShareRequest struct {
	Address string `json:"address"` //base58
}

type ExportDKShareResponse struct {
	DKShare string `json:"dkshare"` // base58
	Err     string `json:"err"`
}

func HandlerExportDKShare(c echo.Context) error {
	var req ExportDKShareRequest
	if err := c.Bind(&req); err != nil {
		return c.JSON(http.StatusBadRequest, &ExportDKShareResponse{Err: err.Error()})
	}
	dkshare, err := exportDKShare(req.Address)
	if err != nil {
		return c.JSON(http.StatusBadRequest, &ExportDKShareResponse{Err: err.Error()})
	}
	return c.JSON(http.StatusOK, &ExportDKShareResponse{DKShare: dkshare})
}

func exportDKShare(base58addr string) (string, error) {
	dkshare, err := registry.GetCommittedDKShare(base58addr)
	if err != nil {
		return "", err
	}
	var buf bytes.Buffer
	err = dkshare.Write(&buf)
	if err != nil {
		return "", err
	}
	return base58.Encode(buf.Bytes()), nil
}
