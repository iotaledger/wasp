package info

import (
	"net/http"

	"github.com/iotaledger/wasp/packages/parameters"
	"github.com/iotaledger/wasp/packages/webapi/model"
	"github.com/iotaledger/wasp/packages/webapi/routes"
	"github.com/iotaledger/wasp/plugins/banner"
	"github.com/iotaledger/wasp/plugins/peering"
	"github.com/labstack/echo/v4"
	"github.com/pangpanglabs/echoswagger/v2"
)

func AddEndpoints(server echoswagger.ApiRouter) {
	server.GET(routes.Info(), handleInfo).
		SetSummary("Get information about the node").
		AddResponse(http.StatusOK, "Node properties", model.InfoResponse{}, nil)
}

func handleInfo(c echo.Context) error {
	return c.JSON(http.StatusOK, model.InfoResponse{
		Version:       banner.Version,
		NetworkID:     peering.DefaultNetworkProvider().Self().NetID(),
		PublisherPort: parameters.GetInt(parameters.NanomsgPublisherPort),
	})
}
