package admapi

import (
	"fmt"
	"github.com/iotaledger/wasp/packages/coretypes"
	"net/http"

	"github.com/iotaledger/wasp/client"
	"github.com/iotaledger/wasp/client/jsonable"
	"github.com/iotaledger/wasp/packages/registry"
	"github.com/iotaledger/wasp/plugins/webapi/httperrors"
	"github.com/labstack/echo"
)

func addChainRecordEndpoints(adm *echo.Group) {
	adm.POST("/"+client.PutChainRecordRoute, handlePutChainRecord)
	adm.GET("/"+client.GetChainRecordRoute(":chainid"), handleGetChainRecord)
	adm.GET("/"+client.GetChainRecordListRoute, handleGetChainRecordList)
}

func handlePutChainRecord(c echo.Context) error {
	var req jsonable.ChainRecord
	var err error

	if err := c.Bind(&req); err != nil {
		return httperrors.BadRequest("Invalid request body")
	}

	bd := req.ChainRecord()

	bd2, err := registry.GetChainRecord(&bd.ChainID)
	if err != nil {
		return err
	}
	if bd2 != nil {
		return httperrors.Conflict(fmt.Sprintf("ChainRecord already exists: %s", bd.ChainID.String()))
	}
	if err = registry.SaveChainRecord(bd); err != nil {
		return err
	}

	log.Infof("ChainRecord saved for addr: %s color: %s", bd.ChainID.String(), bd.Color.String())

	return c.NoContent(http.StatusCreated)
}

func handleGetChainRecord(c echo.Context) error {
	chainID, err := coretypes.NewChainIDFromBase58(c.Param("chainid"))
	if err != nil {
		return httperrors.BadRequest(err.Error())
	}
	bd, err := registry.GetChainRecord(&chainID)
	if err != nil {
		return err
	}
	if bd == nil {
		return httperrors.NotFound(fmt.Sprintf("ChainRecord not found: %s", chainID))
	}
	return c.JSON(http.StatusOK, jsonable.NewChainRecord(bd))
}

func handleGetChainRecordList(c echo.Context) error {
	lst, err := registry.GetChainRecords()
	if err != nil {
		return err
	}
	ret := make([]*jsonable.ChainRecord, len(lst))
	for i := range ret {
		ret[i] = jsonable.NewChainRecord(lst[i])
	}
	return c.JSON(http.StatusOK, ret)
}
