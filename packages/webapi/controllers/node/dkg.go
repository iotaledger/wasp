package node

import (
	"net/http"
	"time"

	"github.com/labstack/echo/v4"

	iotago "github.com/iotaledger/iota.go/v3"
	"github.com/iotaledger/wasp/packages/webapi/apierrors"
	"github.com/iotaledger/wasp/packages/webapi/models"
	"github.com/iotaledger/wasp/packages/webapi/params"
)

func (c *Controller) generateDKS(e echo.Context) error {
	generateDKSRequest := models.DKSharesPostRequest{}

	if err := e.Bind(&generateDKSRequest); err != nil {
		return apierrors.InvalidPropertyError("body", err)
	}

	sharesInfo, err := c.dkgService.GenerateDistributedKey(generateDKSRequest.PeerPubKeysOrNames, generateDKSRequest.Threshold, time.Duration(generateDKSRequest.TimeoutMS)*time.Millisecond)
	if err != nil {
		return apierrors.InternalServerError(err)
	}

	return e.JSON(http.StatusOK, sharesInfo)
}

func (c *Controller) getDKSInfo(e echo.Context) error {
	_, sharedAddress, err := iotago.ParseBech32(e.Param(params.ParamSharedAddress))
	if err != nil {
		return apierrors.InvalidPropertyError(params.ParamSharedAddress, err)
	}

	sharesInfo, err := c.dkgService.GetShares(sharedAddress)
	if err != nil {
		return apierrors.InternalServerError(err)
	}

	return e.JSON(http.StatusOK, sharesInfo)
}
