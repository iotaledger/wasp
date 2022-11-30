package admapi

import (
	"fmt"
	"net/http"

	"github.com/labstack/echo/v4"
	"github.com/pangpanglabs/echoswagger/v2"

	"github.com/iotaledger/wasp/packages/isc"
	"github.com/iotaledger/wasp/packages/registry"
	"github.com/iotaledger/wasp/packages/webapi/httperrors"
	"github.com/iotaledger/wasp/packages/webapi/model"
	"github.com/iotaledger/wasp/packages/webapi/routes"
)

func addChainRecordEndpoints(adm echoswagger.ApiGroup, chainRecordRegistryProvider registry.ChainRecordRegistryProvider) {
	rnd1 := isc.RandomChainID()
	example := model.ChainRecord{
		ChainID: model.NewChainID(rnd1),
		Active:  false,
	}

	s := &chainRecordService{chainRecordRegistryProvider}

	adm.POST(routes.PutChainRecord(), s.handlePutChainRecord).
		SetSummary("Create a new chain record").
		AddParamBody(example, "Record", "Chain record", true)

	adm.GET(routes.GetChainRecord(":chainID"), s.handleGetChainRecord).
		SetSummary("Find the chain record for the given chain ID").
		AddParamPath("", "chainID", "ChainID (bech32)").
		AddResponse(http.StatusOK, "Chain Record", example, nil)

	adm.GET(routes.ListChainRecords(), s.handleGetChainRecordList).
		SetSummary("Get the list of chain records in the node").
		AddResponse(http.StatusOK, "Chain Record", []model.ChainRecord{example}, nil)
}

type chainRecordService struct {
	chainRecordRegistryProvider registry.ChainRecordRegistryProvider
}

func (s *chainRecordService) handlePutChainRecord(c echo.Context) error {
	var req model.ChainRecord

	if err := c.Bind(&req); err != nil {
		return httperrors.BadRequest("Invalid request body")
	}

	bd := req.Record()

	bd2, err := s.chainRecordRegistryProvider.ChainRecord(bd.ChainID())
	if err != nil {
		return err
	}
	if bd2 != nil {
		// Make this call idempotent.
		// Record has no information apart from the ChainID and activation status.
		// So just keep the existing, if it exists.
		return c.NoContent(http.StatusCreated)
	}
	if err := s.chainRecordRegistryProvider.AddChainRecord(bd); err != nil {
		return err
	}

	log.Infof("Chain record saved: ChainID: %s (active: %t)", bd.ChainID(), bd.Active)

	return c.NoContent(http.StatusCreated)
}

func (s *chainRecordService) handleGetChainRecord(c echo.Context) error {
	chainID, err := isc.ChainIDFromString(c.Param("chainID"))
	if err != nil {
		return httperrors.BadRequest(err.Error())
	}
	bd, err := s.chainRecordRegistryProvider.ChainRecord(*chainID)
	if err != nil {
		return err
	}
	if bd == nil {
		return httperrors.NotFound(fmt.Sprintf("Record not found: %s", chainID))
	}
	return c.JSON(http.StatusOK, model.NewChainRecord(bd))
}

func (s *chainRecordService) handleGetChainRecordList(c echo.Context) error {
	lst, err := s.chainRecordRegistryProvider.ChainRecords()
	if err != nil {
		return err
	}
	ret := make([]*model.ChainRecord, len(lst))
	for i := range ret {
		ret[i] = model.NewChainRecord(lst[i])
	}
	return c.JSON(http.StatusOK, ret)
}
