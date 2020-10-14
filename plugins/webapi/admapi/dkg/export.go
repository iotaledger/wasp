package dkg

import (
	"bytes"
	"fmt"
	"net/http"

	"github.com/iotaledger/goshimmer/dapps/valuetransfers/packages/address"
	"github.com/iotaledger/wasp/client"
	"github.com/iotaledger/wasp/packages/registry"
	"github.com/iotaledger/wasp/packages/tcrypto"
	"github.com/iotaledger/wasp/plugins/webapi/httperrors"
	"github.com/labstack/echo"
)

func addExportEndpoints(adm *echo.Group) {
	adm.GET("/"+client.DKSExportRoute(":address"), handleExportDKShare)
	adm.POST("/"+client.DKSImportRoute, handleImportDKShare)
}

func handleExportDKShare(c echo.Context) error {
	addr, err := address.FromBase58(c.Param("address"))
	if err != nil {
		return httperrors.BadRequest(fmt.Sprintf("Invalid SC address: %s", c.Param("address")))
	}

	dkshare, exist, err := registry.GetDKShare(&addr)
	if err != nil {
		return err
	}
	if !exist {
		return httperrors.NotFound(fmt.Sprintf("DKShare not found for address %s", addr.String()))
	}

	blob, err := exportDKShare(dkshare)
	if err != nil {
		return err
	}
	return c.JSON(http.StatusOK, &client.DKShare{Blob: blob})
}

func exportDKShare(dkshare *tcrypto.DKShare) ([]byte, error) {
	var buf bytes.Buffer
	err := dkshare.Write(&buf)
	if err != nil {
		return nil, err
	}
	return buf.Bytes(), nil
}

func handleImportDKShare(c echo.Context) error {
	var req client.DKShare
	if err := c.Bind(&req); err != nil {
		return httperrors.BadRequest("Invalid request body")
	}

	dks, err := tcrypto.UnmarshalDKShare(req.Blob, false)
	if err != nil {
		return httperrors.BadRequest("Cannot unmarshal DKShare")
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
		if !bytes.Equal(oldBlob, req.Blob) {
			return httperrors.Conflict(fmt.Sprintf("A different DKShare exists with same address %s", dks.Address))
		}
		log.Debugf("DKShare with address %s already imported", dks.Address)
	} else {
		log.Infof("Importing DKShare with address %s...", dks.Address)
		err = registry.SaveDKShareToRegistry(dks)
		if err != nil {
			return err
		}
	}

	return c.NoContent(http.StatusOK)
}
