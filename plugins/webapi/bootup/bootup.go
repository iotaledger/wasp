package bootup

import (
	"fmt"
	"net/http"

	"github.com/iotaledger/goshimmer/dapps/valuetransfers/packages/address"
	"github.com/iotaledger/hive.go/logger"
	"github.com/iotaledger/wasp/client"
	"github.com/iotaledger/wasp/client/jsonable"
	"github.com/iotaledger/wasp/packages/registry"
	"github.com/iotaledger/wasp/plugins/webapi/httperrors"
	"github.com/iotaledger/wasp/plugins/webapi/misc"
	"github.com/labstack/echo"
)

var log *logger.Logger

func initLogger() {
	log = logger.NewLogger("webapi/bootup")
}

func AddEndpoints(server *echo.Group) {
	initLogger()
	server.POST("/"+client.PutBootupDataRoute, handlePutBootupData)
	server.GET("/"+client.GetBootupDataRoute(":address"), handleGetBootupData)
	server.GET("/"+client.GetBootupDataListRoute, handleGetBootupDataList)
}

func handlePutBootupData(c echo.Context) error {
	var req jsonable.BootupData
	var err error

	if err := c.Bind(&req); err != nil {
		return httperrors.BadRequest("Invalid request body")
	}

	bd := req.BootupData()

	bd2, err := registry.GetBootupData(&bd.Address)
	if err != nil {
		return err
	}
	if bd2 != nil {
		return httperrors.Conflict(fmt.Sprintf("Bootup data already exists: %s", bd.Address.String()))
	}
	if err = registry.SaveBootupData(bd); err != nil {
		return err
	}

	log.Infof("Bootup record saved for addr: %s color: %s", bd.Address.String(), bd.Color.String())

	return c.JSON(http.StatusCreated, "OK")
}

func handleGetBootupData(c echo.Context) error {
	addr, err := address.FromBase58(c.Param("address"))
	if err != nil {
		return httperrors.BadRequest(err.Error())
	}
	bd, err := registry.GetBootupData(&addr)
	if err != nil {
		return err
	}
	if bd == nil {
		return httperrors.NotFound(fmt.Sprintf("Bootup data not found: %s", addr))
	}
	return misc.OkJson(c, jsonable.NewBootupData(bd))
}

func handleGetBootupDataList(c echo.Context) error {
	lst, err := registry.GetBootupRecords()
	if err != nil {
		return err
	}
	ret := make([]*jsonable.BootupData, len(lst))
	for i := range ret {
		ret[i] = jsonable.NewBootupData(lst[i])
	}
	return misc.OkJson(c, ret)
}
