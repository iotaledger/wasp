package admapi

import (
	"fmt"
	"net/http"

	"github.com/iotaledger/wasp/packages/coretypes"
	"github.com/iotaledger/wasp/packages/registry"
	"github.com/iotaledger/wasp/packages/webapi/httperrors"
	"github.com/iotaledger/wasp/packages/webapi/model"
	"github.com/iotaledger/wasp/packages/webapi/routes"
	"github.com/labstack/echo/v4"
	"github.com/pangpanglabs/echoswagger/v2"
)

func addChainRecordEndpoints(adm echoswagger.ApiGroup) {
	example := model.ChainRecord{
		ChainID:        model.NewChainID(&coretypes.ChainID{1, 2, 3, 4}),
		Color:          model.NewColor(&ledgerstate.Color{5, 6, 7, 8}),
		CommitteeNodes: []string{"wasp1:4000", "wasp2:4000"},
		Active:         false,
	}

	adm.POST(routes.PutChainRecord(), handlePutChainRecord).
		SetSummary("Create a new chain record").
		AddParamBody(example, "ChainRecord", "Chain record", true)

	adm.GET(routes.GetChainRecord(":chainID"), handleGetChainRecord).
		SetSummary("Find the chain record for the given chain ID").
		AddParamPath("", "chainID", "ChainID (base58)").
		AddResponse(http.StatusOK, "Chain Record", example, nil)

	adm.GET(routes.ListChainRecords(), handleGetChainRecordList).
		SetSummary("Get the list of chain records in the node").
		AddResponse(http.StatusOK, "Chain Record", []model.ChainRecord{example}, nil)
}

func handlePutChainRecord(c echo.Context) error {
	var req model.ChainRecord

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
	chainID, err := coretypes.NewChainIDFromBase58(c.Param("chainID"))
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
	return c.JSON(http.StatusOK, model.NewChainRecord(bd))
}

func handleGetChainRecordList(c echo.Context) error {
	lst, err := registry.GetChainRecords()
	if err != nil {
		return err
	}
	ret := make([]*model.ChainRecord, len(lst))
	for i := range ret {
		ret[i] = model.NewChainRecord(lst[i])
	}
	return c.JSON(http.StatusOK, ret)
}
