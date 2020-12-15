package info

import (
	"net/http"

	"github.com/iotaledger/wasp/client"
	"github.com/iotaledger/wasp/packages/parameters"
	"github.com/iotaledger/wasp/plugins/banner"
	"github.com/iotaledger/wasp/plugins/peering"
	"github.com/labstack/echo"
)

func AddEndpoints(server *echo.Echo) {
	server.GET("/"+client.InfoRoute, handleInfo)
}

func handleInfo(c echo.Context) error {
	return c.JSON(http.StatusOK, client.InfoResponse{
		Version:       banner.AppVersion,
		NetworkId:     peering.DefaultNetworkProvider().Self().NetID(),
		PublisherPort: parameters.GetInt(parameters.NanomsgPublisherPort),
	})
}
