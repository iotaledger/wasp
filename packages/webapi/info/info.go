package info

import (
	"net/http"

	"github.com/iotaledger/wasp/packages/parameters"
	"github.com/iotaledger/wasp/packages/peering"
	"github.com/iotaledger/wasp/packages/wasp"
	"github.com/iotaledger/wasp/packages/webapi/model"
	"github.com/iotaledger/wasp/packages/webapi/routes"
	"github.com/labstack/echo/v4"
	"github.com/pangpanglabs/echoswagger/v2"
)

type infoService struct {
	network peering.NetworkProvider
}

func AddEndpoints(server echoswagger.ApiRouter, network peering.NetworkProvider) {
	s := &infoService{network}

	server.GET(routes.Info(), s.handleInfo).
		SetSummary("Get information about the node").
		AddResponse(http.StatusOK, "Node properties", model.InfoResponse{}, nil)
}

func (s *infoService) handleInfo(c echo.Context) error {
	return c.JSON(http.StatusOK, model.InfoResponse{
		Version:       wasp.Version,
		VersionHash:   wasp.VersionHash,
		NetworkID:     s.network.Self().NetID(),
		PublisherPort: parameters.GetInt(parameters.NanomsgPublisherPort),
	})
}
