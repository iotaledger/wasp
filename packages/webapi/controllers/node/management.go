package node

import (
	"net/http"

	"github.com/iotaledger/wasp/packages/cryptolib"
	"github.com/iotaledger/wasp/packages/webapi/apierrors"
	"github.com/iotaledger/wasp/packages/webapi/models"

	"github.com/labstack/echo/v4"

	iotago "github.com/iotaledger/iota.go/v3"
)

func (c *Controller) setNodeOwner(e echo.Context) error {
	var request models.NodeOwnerCertificateRequest
	if err := e.Bind(&request); err != nil {
		return apierrors.InvalidPropertyError("body", err)
	}

	reqPublicKeyBytes, err := iotago.DecodeHex(request.PublicKey)
	if err != nil {
		return apierrors.InvalidPropertyError("PublicKey", err)
	}

	reqPublicKey, err := cryptolib.NewPublicKeyFromBytes(reqPublicKeyBytes)
	if err != nil {
		return apierrors.InvalidPropertyError("PublicKey", err)
	}

	_, reqOwnerAddress, err := iotago.ParseBech32(request.OwnerAddress)
	if err != nil {
		return apierrors.InvalidPropertyError("OwnerAddress", err)
	}

	certificateBytes, err := c.nodeService.SetNodeOwnerCertificate(reqPublicKey, reqOwnerAddress)
	if err != nil {
		return err
	}

	response := models.NodeOwnerCertificateResponse{
		Certificate: iotago.EncodeHex(certificateBytes),
	}

	return e.JSON(http.StatusOK, response)
}

func (c *Controller) shutdownNode(e echo.Context) error {
	c.nodeService.ShutdownNode()
	return e.NoContent(http.StatusOK)
}
