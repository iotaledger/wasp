package node

import (
	"net/http"

	"github.com/ethereum/go-ethereum/common/hexutil"

	"github.com/iotaledger/wasp/v2/packages/webapi/models"

	"github.com/labstack/echo/v4"
)

func (c *Controller) nodeOwnerCertificate(e echo.Context) error {
	certificateBytes := c.nodeService.NodeOwnerCertificate()

	response := models.NodeOwnerCertificateResponse{
		Certificate: hexutil.Encode(certificateBytes),
	}

	return e.JSON(http.StatusOK, response)
}

func (c *Controller) shutdownNode(e echo.Context) error {
	c.nodeService.ShutdownNode()
	return e.NoContent(http.StatusOK)
}
