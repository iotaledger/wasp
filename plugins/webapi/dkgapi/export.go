package dkgapi

import (
	"bytes"
	"fmt"
	"github.com/iotaledger/goshimmer/dapps/valuetransfers/packages/address"
	"net/http"

	"github.com/iotaledger/wasp/packages/registry"
	"github.com/iotaledger/wasp/packages/tcrypto"
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

type ImportDKShareRequest struct {
	Blob string `json:"blob"` //base58
}

type ImportDKShareResponse struct {
	Err string `json:"err"`
}

func HandlerExportDKShare(c echo.Context) error {
	log.Debugw("HandlerExportDKShare")
	var req ExportDKShareRequest
	if err := c.Bind(&req); err != nil {
		return c.JSON(http.StatusBadRequest, &ExportDKShareResponse{Err: err.Error()})
	}
	log.Debugw("HandlerExportDKShare", "req", req)

	addr, err := address.FromBase58(req.Address)
	if err != nil {
		return c.JSON(http.StatusBadRequest, &ExportDKShareResponse{Err: err.Error()})
	}
	log.Debugw("HandlerExportDKShare", "req", req, "addr")
	dkshare, exist, err := registry.GetDKShare(&addr)
	if err != nil {
		return c.JSON(http.StatusBadRequest, &ExportDKShareResponse{Err: err.Error()})
	}
	if !exist {
		return c.JSON(http.StatusBadRequest, &ExportDKShareResponse{Err: "dkshare not found"})
	}
	log.Debugw("HandlerExportDKShare", "req", "before export")
	blob, err := exportDKShare(dkshare)
	if err != nil {
		return c.JSON(http.StatusBadRequest, &ExportDKShareResponse{Err: err.Error()})
	}
	log.Debugw("HandlerExportDKShare", "req", "after export")
	return c.JSON(http.StatusOK, &ExportDKShareResponse{DKShare: blob})
}

func exportDKShare(dkshare *tcrypto.DKShare) (string, error) {
	var buf bytes.Buffer
	err := dkshare.Write(&buf)
	if err != nil {
		return "", err
	}
	return base58.Encode(buf.Bytes()), nil
}

func HandlerImportDKShare(c echo.Context) error {
	log.Debugw("HandlerImportDKShare")

	var req ImportDKShareRequest
	if err := c.Bind(&req); err != nil {
		return c.JSON(http.StatusBadRequest, &ImportDKShareResponse{Err: err.Error()})
	}
	log.Debugw("HandlerImportDKShare", "req", req)

	err := importDKShare(req.Blob)
	if err != nil {
		return c.JSON(http.StatusBadRequest, &ImportDKShareResponse{Err: err.Error()})
	}
	return c.JSON(http.StatusOK, &ImportDKShareResponse{})
}

func importDKShare(blob string) error {
	data, err := base58.Decode(blob)
	if err != nil {
		return err
	}
	dks, err := tcrypto.UnmarshalDKShare(data, false)
	if err != nil {
		return err
	}

	oldDks, exists, err := registry.GetDKShare(dks.Address)
	if err != nil {
		return err
	}
	if exists {
		oldBlob, err := exportDKShare(oldDks)
		if err != nil {
			return err
		}
		if oldBlob != blob {
			return fmt.Errorf("A different DKShare exists with same address %s", dks.Address)
		}
		log.Debugf("DKShare with address %s already imported", dks.Address)
		return nil
	}

	log.Infof("Importing DKShare with address %s...", dks.Address)
	return registry.SaveDKShareToRegistry(dks)
}
