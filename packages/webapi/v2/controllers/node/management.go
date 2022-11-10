package node

import (
	"net/http"

	"github.com/iotaledger/wasp/packages/webapi/v2/types"

	"github.com/iotaledger/wasp/packages/webapi/v2/apierrors"
	"github.com/iotaledger/wasp/packages/webapi/v2/models"

	"github.com/labstack/echo/v4"
)

func (c *Controller) setNodeOwner(e echo.Context) error {
	var request models.NodeOwnerCertificateRequest
	if err := e.Bind(&request); err != nil {
		return apierrors.InvalidPropertyError("body", err)
	}

	reqNodePubKeyBytes := request.NodePubKey.Bytes()
	reqOwnerAddress := request.OwnerAddress.Address()

	certificateBytes, err := c.nodeService.SetNodeOwnerCertificate(reqNodePubKeyBytes, reqOwnerAddress)

	if err != nil {
		return err
	}

	response := models.NodeOwnerCertificateResponse{
		Certificate: types.NewBase64(certificateBytes),
	}

	return e.JSON(http.StatusOK, response)
}

func (c *Controller) shutdownNode(e echo.Context) error {
	c.nodeService.ShutdownNode()
	return e.NoContent(http.StatusOK)
}
