package admapi

import (
	"fmt"
	"net/http"

	"github.com/iotaledger/wasp/packages/webapi/v1/httperrors"
	"github.com/iotaledger/wasp/packages/webapi/v1/model"
	"github.com/iotaledger/wasp/packages/webapi/v1/routes"

	"github.com/labstack/echo/v4"
	"github.com/pangpanglabs/echoswagger/v2"

	"github.com/iotaledger/wasp/packages/chains"
	"github.com/iotaledger/wasp/packages/isc"
	"github.com/iotaledger/wasp/packages/registry"
)

func addChainRecordEndpoints(adm echoswagger.ApiGroup, chainRecordRegistryProvider registry.ChainRecordRegistryProvider, chainsProvider chains.Provider) {
	rnd1 := isc.RandomChainID()
	example := model.ChainRecord{
		ChainID: model.NewChainIDBech32(rnd1),
		Active:  false,
	}

	s := &chainRecordService{chainRecordRegistryProvider, chainsProvider}

	adm.POST(routes.PutChainRecord(), s.handlePutChainRecord).
		SetDeprecated().
		SetSummary("Create a new chain record").
		AddParamBody(example, "Record", "Chain record", true)

	adm.GET(routes.GetChainRecord(":chainID"), s.handleGetChainRecord).
		SetDeprecated().
		SetSummary("Find the chain record for the given chain ID").
		AddParamPath("", "chainID", "ChainID (bech32)").
		AddResponse(http.StatusOK, "Chain Record", example, nil)

	adm.GET(routes.ListChainRecords(), s.handleGetChainRecordList).
		SetDeprecated().
		SetSummary("Get the list of chain records in the node").
		AddResponse(http.StatusOK, "Chain Record", []model.ChainRecord{example}, nil)
}

type chainRecordService struct {
	chainRecordRegistryProvider registry.ChainRecordRegistryProvider
	chainsProvider              chains.Provider
}

func (s *chainRecordService) handlePutChainRecord(c echo.Context) error {
	var req model.ChainRecord

	if err := c.Bind(&req); err != nil {
		return httperrors.BadRequest("Invalid request body")
	}

	requestChainRec, err := req.Record()
	if err != nil {
		return httperrors.BadRequest("Error parsing chain record")
	}

	storedChainRec, err := s.chainRecordRegistryProvider.ChainRecord(requestChainRec.ChainID())
	if err != nil {
		return err
	}
	if storedChainRec != nil {
		_, err = s.chainRecordRegistryProvider.UpdateChainRecord(
			requestChainRec.ChainID(),
			func(rec *registry.ChainRecord) bool {
				rec.AccessNodes = requestChainRec.AccessNodes
				rec.Active = requestChainRec.Active
				return true
			},
		)
		if err != nil {
			return err
		}
		return c.NoContent(http.StatusAccepted)
	}
	if err := s.chainRecordRegistryProvider.AddChainRecord(requestChainRec); err != nil {
		return err
	}

	log.Infof("Chain record saved: ChainID: %s (active: %t)", requestChainRec.ChainID(), requestChainRec.Active)

	// Activate/deactivate the chain accordingly.
	if requestChainRec.Active {
		log.Debugw("calling Chains.Activate", "chainID", requestChainRec.ChainID().String())
		if err := s.chainsProvider().Activate(requestChainRec.ChainID()); err != nil {
			return err
		}
	} else {
		log.Debugw("calling Chains.Deactivate", "chainID", requestChainRec.ChainID().String())
		if err := s.chainsProvider().Deactivate(requestChainRec.ChainID()); err != nil {
			return err
		}
	}

	return c.NoContent(http.StatusCreated)
}

func (s *chainRecordService) handleGetChainRecord(c echo.Context) error {
	chainID, err := isc.ChainIDFromString(c.Param("chainID"))
	if err != nil {
		return httperrors.BadRequest(err.Error())
	}
	bd, err := s.chainRecordRegistryProvider.ChainRecord(chainID)
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
