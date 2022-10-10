package controllers

import (
	"net/http"

	"github.com/iotaledger/wasp/packages/wasp"

	"github.com/iotaledger/hive.go/core/configuration"

	"github.com/labstack/echo/v4"
	"github.com/pangpanglabs/echoswagger/v2"

	loggerpkg "github.com/iotaledger/hive.go/core/logger"
	"github.com/iotaledger/wasp/packages/webapi/v2/interfaces"
)

type InfoController struct {
	log *loggerpkg.Logger

	config *configuration.Configuration
}

func NewInfoController(log *loggerpkg.Logger, config *configuration.Configuration) interfaces.APIController {
	return &InfoController{
		log:    log,
		config: config,
	}
}

func (c *InfoController) Name() string {
	return "info"
}

func (c *InfoController) getConfiguration(e echo.Context) error {
	return e.JSON(http.StatusOK, c.config.Koanf().All())
}

func (c *InfoController) getInfo(e echo.Context) error {
	return e.JSON(http.StatusOK, wasp.Version)
}

func (c *InfoController) RegisterExampleData(mock interfaces.Mocker) {
}

func (c *InfoController) RegisterPublic(publicAPI echoswagger.ApiGroup, mocker interfaces.Mocker) {
	publicAPI.GET("", c.getInfo).
		AddResponse(http.StatusOK, "Dumps configuration", nil, nil).
		SetOperationId("getInfo").
		SetSummary("Returns the Wasp configuration")
}

func (c *InfoController) RegisterAdmin(adminAPI echoswagger.ApiGroup, mocker interfaces.Mocker) {
	adminAPI.GET("config", c.getConfiguration).
		AddResponse(http.StatusOK, "Dumps configuration", nil, nil).
		SetOperationId("getConfiguration").
		SetSummary("Returns the Wasp configuration")
}
