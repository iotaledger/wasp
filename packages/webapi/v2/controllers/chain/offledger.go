package chain

import (
	"encoding/base64"
	"io"
	"net/http"

	"github.com/labstack/echo/v4"

	"github.com/iotaledger/wasp/packages/isc"
	"github.com/iotaledger/wasp/packages/webapi/v2/apierrors"
	"github.com/iotaledger/wasp/packages/webapi/v2/dto"
)

func (c *Controller) handleNewRequest(e echo.Context) error {
	chainID, err := isc.ChainIDFromString(e.Param("chainID"))
	if err != nil {
		return apierrors.InvalidPropertyError("chainID", err)
	}

	offLedgerReq, err := readRequest(e)
	if err != nil {
		return apierrors.InvalidOffLedgerRequestError(err)
	}

	err = c.offLedgerService.EnqueueOffLedgerRequest(chainID, offLedgerReq)
	if err != nil {
		return apierrors.ContractExecutionError(err)
	}

	return e.NoContent(http.StatusAccepted)
}

func readBinaryRequest(c echo.Context) ([]byte, error) {
	request, err := io.ReadAll(c.Request().Body)
	if err != nil {
		return nil, err
	}

	return request, err
}

func readJSONRequest(c echo.Context) ([]byte, error) {
	request := new(dto.OffLedgerRequestBody)
	if err := c.Bind(request); err != nil {
		return nil, err
	}

	requestDecoded, err := base64.StdEncoding.DecodeString(request.Request)

	return requestDecoded, err
}

func readRequest(c echo.Context) ([]byte, error) {
	contentType := c.Request().
		Header.
		Get("Content-Type")

	if contentType == echo.MIMEApplicationJavaScript {
		return readJSONRequest(c)
	}

	return readBinaryRequest(c)
}
