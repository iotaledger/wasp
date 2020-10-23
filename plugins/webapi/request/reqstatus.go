package request

import (
	"fmt"
	"github.com/iotaledger/wasp/packages/coretypes"
	"net/http"

	"github.com/iotaledger/goshimmer/dapps/valuetransfers/packages/address"
	"github.com/iotaledger/wasp/client"
	"github.com/iotaledger/wasp/packages/committee"
	"github.com/iotaledger/wasp/plugins/committees"
	"github.com/iotaledger/wasp/plugins/webapi/httperrors"
	"github.com/labstack/echo"
)

func AddEndpoints(server *echo.Echo) {
	server.GET("/"+client.RequestStatusRoute(":scAddr", ":reqId"), handleRequestStatus)
}

func handleRequestStatus(c echo.Context) error {
	addr, err := address.FromBase58(c.Param("scAddr"))
	if err != nil {
		return httperrors.BadRequest(fmt.Sprintf("Invalid SC address %+v: %s", c.Param("scAddr"), err.Error()))
	}
	cmt := committees.CommitteeByAddress(addr)
	if cmt == nil {
		return httperrors.NotFound(fmt.Sprintf("Smart contract not found: %+v", addr.String()))
	}
	reqId, err := coretypes.NewRequestIDFromBase58(c.Param("reqId"))
	if err != nil {
		return httperrors.BadRequest(fmt.Sprintf("Invalid request id %+v: %s", c.Param("reqId"), err.Error()))
	}
	var isProcessed bool
	switch cmt.GetRequestProcessingStatus(&reqId) {
	case committee.RequestProcessingStatusCompleted:
		isProcessed = true
	case committee.RequestProcessingStatusBacklog:
		isProcessed = false
	}
	return c.JSON(http.StatusOK, client.RequestStatusResponse{
		IsProcessed: isProcessed,
	})
}
