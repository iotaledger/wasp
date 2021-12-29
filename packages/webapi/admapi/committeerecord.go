package admapi

import (
	"fmt"
	"net/http"
	"strconv"

	"github.com/iotaledger/wasp/packages/chains"
	"github.com/iotaledger/wasp/packages/iscp"
	"github.com/iotaledger/wasp/packages/registry"
	"github.com/iotaledger/wasp/packages/webapi/httperrors"
	"github.com/iotaledger/wasp/packages/webapi/model"
	"github.com/iotaledger/wasp/packages/webapi/routes"
	"github.com/labstack/echo/v4"
	"github.com/pangpanglabs/echoswagger/v2"
)

func addCommitteeRecordEndpoints(adm echoswagger.ApiGroup, registryProvider registry.Provider, chainsProvider chains.Provider) {
	rnd1 := iscp.RandomChainID()
	example := model.CommitteeRecord{
		Address: model.NewAddress(rnd1.AsAddress()),
		Nodes:   []string{"wasp1.example.org:4000", "wasp2.example.org:4000", "wasp3.example.org:4000"},
	}

	s := &committeeRecordService{registryProvider, chainsProvider}

	adm.POST(routes.PutCommitteeRecord(), s.handlePutCommitteeRecord).
		SetSummary("Create a new committee record").
		AddParamBody(example, "Record", "Committee record", true)

	adm.GET(routes.GetCommitteeRecord(":address"), s.handleGetCommitteeRecord).
		SetSummary("Find the committee record for the given address").
		AddParamPath("", "address", "Address (base58)").
		AddResponse(http.StatusOK, "Committee Record", example, nil)

	adm.GET(routes.GetCommitteeForChain(":chainID"), s.handleGetCommitteeForChain).
		SetSummary("Find the committee record that manages the given chain").
		AddParamPath("", "chainID", "ChainID (base58)").
		AddResponse(http.StatusOK, "Committee Record", example, nil)
}

type committeeRecordService struct {
	registry registry.Provider
	chains   chains.Provider
}

func (s *committeeRecordService) handlePutCommitteeRecord(c echo.Context) error {
	var req model.CommitteeRecord

	if err := c.Bind(&req); err != nil {
		return httperrors.BadRequest("Invalid request body")
	}

	cr := req.Record()

	defaultRegistry := s.registry()
	bd2, err := defaultRegistry.GetCommitteeRecord(cr.Address)
	if err != nil {
		return err
	}
	if bd2 != nil {
		return httperrors.Conflict(fmt.Sprintf("Record already exists: %s", cr.Address.Base58()))
	}
	// TODO if I am not among committee nodes, should not save
	if err := defaultRegistry.SaveCommitteeRecord(cr); err != nil {
		return err
	}

	log.Infof("Committee record saved. Address: %s", cr.String())

	return c.NoContent(http.StatusCreated)
}

func (s *committeeRecordService) handleGetCommitteeRecord(c echo.Context) error {
	address, err := iotago.AddressFromBase58EncodedString(c.Param("address"))
	if err != nil {
		return httperrors.BadRequest(err.Error())
	}
	cr, err := s.registry().GetCommitteeRecord(address)
	if err != nil {
		return err
	}
	if cr == nil {
		return httperrors.NotFound(fmt.Sprintf("Record not found: %s", address))
	}
	return c.JSON(http.StatusOK, model.NewCommitteeRecord(cr))
}

func (s *committeeRecordService) handleGetCommitteeForChain(c echo.Context) error {
	chainID, err := iscp.ChainIDFromHex(c.Param("chainID"))
	if err != nil {
		return httperrors.BadRequest(err.Error())
	}
	includeDeactivated, _ := strconv.ParseBool(c.QueryParam("includeDeactivated"))
	chain := s.chains().Get(chainID, includeDeactivated)
	if chain == nil {
		return httperrors.NotFound(fmt.Sprintf("Active chain %s not found", chainID))
	}
	committeeInfo := chain.GetCommitteeInfo()
	if committeeInfo == nil {
		return httperrors.NotFound(fmt.Sprintf("Committee info for chain %s is not available", chainID))
	}
	cr, err := s.registry().GetCommitteeRecord(committeeInfo.Address)
	if err != nil {
		return err
	}
	if cr == nil {
		return httperrors.NotFound(fmt.Sprintf("Committee record not found for address: %s", committeeInfo.Address.Base58()))
	}
	return c.JSON(http.StatusOK, model.NewCommitteeRecord(cr))
}
