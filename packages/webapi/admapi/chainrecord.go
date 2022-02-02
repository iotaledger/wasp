package admapi

import (
	"fmt"
	"net/http"

	"github.com/iotaledger/wasp/packages/iscp"
	"github.com/iotaledger/wasp/packages/registry"
	"github.com/iotaledger/wasp/packages/webapi/httperrors"
	"github.com/iotaledger/wasp/packages/webapi/model"
	"github.com/iotaledger/wasp/packages/webapi/routes"
	"github.com/labstack/echo/v4"
	"github.com/pangpanglabs/echoswagger/v2"
)

func addChainRecordEndpoints(adm echoswagger.ApiGroup, registryProvider registry.Provider) {
	rnd1 := iscp.RandomChainID()
	example := model.ChainRecord{
		ChainID: model.NewChainID(rnd1),
		Active:  false,
	}

	s := &chainRecordService{registryProvider}

	adm.POST(routes.PutChainRecord(), s.handlePutChainRecord).
		SetSummary("Create a new chain record").
		AddParamBody(example, "Record", "Chain record", true)

	adm.GET(routes.GetChainRecord(":chainID"), s.handleGetChainRecord).
		SetSummary("Find the chain record for the given chain ID").
		AddParamPath("", "chainID", "ChainID (base58)").
		AddResponse(http.StatusOK, "Chain Record", example, nil)

	adm.GET(routes.ListChainRecords(), s.handleGetChainRecordList).
		SetSummary("Get the list of chain records in the node").
		AddResponse(http.StatusOK, "Chain Record", []model.ChainRecord{example}, nil)
}

type chainRecordService struct {
	registry registry.Provider
}

func (s *chainRecordService) handlePutChainRecord(c echo.Context) error {
	var req model.ChainRecord

	if err := c.Bind(&req); err != nil {
		return httperrors.BadRequest("Invalid request body")
	}

	reg := s.registry()
	bd := req.Record()

	bd2, err := reg.GetChainRecordByChainID(bd.ChainID)
	if err != nil {
		return err
	}
	if bd2 != nil {
		// Make this call idempotent.
		// Record has no information apart from the ChainID and activation status.
		// So just keep the existing, if it exists.
		return c.NoContent(http.StatusCreated)
	}
	if err := reg.SaveChainRecord(bd); err != nil {
		return err
	}

	log.Infof("Chain record saved: %s", bd.String())

	return c.NoContent(http.StatusCreated)
}

func (s *chainRecordService) handleGetChainRecord(c echo.Context) error {
	chainID, err := iscp.ChainIDFromBase58(c.Param("chainID"))
	if err != nil {
		return httperrors.BadRequest(err.Error())
	}
	bd, err := s.registry().GetChainRecordByChainID(chainID)
	if err != nil {
		return err
	}
	if bd == nil {
		return httperrors.NotFound(fmt.Sprintf("Record not found: %s", chainID))
	}
	return c.JSON(http.StatusOK, model.NewChainRecord(bd))
}

func (s *chainRecordService) handleGetChainRecordList(c echo.Context) error {
	lst, err := s.registry().GetChainRecords()
	if err != nil {
		return err
	}
	ret := make([]*model.ChainRecord, len(lst))
	for i := range ret {
		ret[i] = model.NewChainRecord(lst[i])
	}
	return c.JSON(http.StatusOK, ret)
}
