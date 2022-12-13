package node

import (
	"net/http"

	"github.com/ethereum/go-ethereum/common/hexutil"

	iotago "github.com/iotaledger/iota.go/v3"

	"github.com/iotaledger/wasp/packages/webapi/v2/apierrors"
	"github.com/iotaledger/wasp/packages/webapi/v2/models"

	"github.com/labstack/echo/v4"
)

func (c *Controller) setNodeOwner(e echo.Context) error {
	var request models.NodeOwnerCertificateRequest
	if err := e.Bind(&request); err != nil {
		return apierrors.InvalidPropertyError("body", err)
	}

	reqNodePubKeyBytes, err := hexutil.Decode(request.NodePubKey)
	if err != nil {
		return apierrors.InvalidPropertyError("NodePubKey", err)
	}

	_, reqOwnerAddress, err := iotago.ParseBech32(request.OwnerAddress)
	if err != nil {
		return apierrors.InvalidPropertyError("OwnerAddress", err)
	}

	certificateBytes, err := c.nodeService.SetNodeOwnerCertificate(reqNodePubKeyBytes, reqOwnerAddress)
	if err != nil {
		return err
	}

	response := models.NodeOwnerCertificateResponse{
		Certificate: hexutil.Encode(certificateBytes),
	}

	return e.JSON(http.StatusOK, response)
}

func (c *Controller) shutdownNode(e echo.Context) error {
	c.nodeService.ShutdownNode()
	return e.NoContent(http.StatusOK)
}
