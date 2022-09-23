package chain

import (
	"fmt"
	"io"
	"net/http"

	"github.com/labstack/echo/v4"

	"github.com/iotaledger/hive.go/core/marshalutil"
	"github.com/iotaledger/wasp/packages/isc"
	"github.com/iotaledger/wasp/packages/webapi/v1/httperrors"
	"github.com/iotaledger/wasp/packages/webapi/v1/model"
)

func (c *ChainController) handleNewRequest(e echo.Context) error {
	chainID, err := isc.ChainIDFromString(e.Param("chainID"))
	if err != nil {
		return httperrors.BadRequest(fmt.Sprintf("Invalid Chain ID %+v: %s", e.Param("chainID"), err.Error()))
	}

	offLedgerReq, err := parseOffLedgerRequest(e)
	if err != nil {
		return err
	}

	err = c.offLedgerService.EnqueueOffLedgerRequest(chainID, offLedgerReq)
	if err != nil {
		return err
	}

	return e.NoContent(http.StatusAccepted)
}

func parseBinaryRequest(c echo.Context) (reg isc.OffLedgerRequest, err error) {
	reqBytes, err := io.ReadAll(c.Request().Body)
	if err != nil {
		return nil, httperrors.BadRequest("error parsing request from payload")
	}

	rGeneric, err := isc.NewRequestFromMarshalUtil(marshalutil.New(reqBytes))
	if err != nil {
		return nil, httperrors.BadRequest("error parsing request from payload")
	}

	req, ok := rGeneric.(isc.OffLedgerRequest)
	if !ok {
		return nil, httperrors.BadRequest("error parsing request: off-ledger request expected")
	}

	return req, err
}

func parseJSONRequest(c echo.Context) (req isc.OffLedgerRequest, err error) {
	r := new(model.OffLedgerRequestBody)
	if err = c.Bind(r); err != nil {
		return nil, httperrors.BadRequest("error parsing request from payload")
	}

	rGeneric, err := isc.NewRequestFromMarshalUtil(marshalutil.New(r.Request.Bytes()))
	if err != nil {
		return nil, httperrors.BadRequest(fmt.Sprintf("cannot decode off-ledger request: %v", err))
	}

	var ok bool
	if req, ok = rGeneric.(isc.OffLedgerRequest); !ok {
		return nil, httperrors.BadRequest("error parsing request: off-ledger request is expected")
	}

	return req, err
}

func parseOffLedgerRequest(c echo.Context) (req isc.OffLedgerRequest, err error) {
	contentType := c.Request().
		Header.
		Get("Content-Type")

	if contentType == echo.MIMEApplicationJavaScript {
		return parseJSONRequest(c)
	}

	return parseBinaryRequest(c)
}
