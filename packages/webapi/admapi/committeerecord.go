package admapi

import (
	"fmt"
	"net/http"

	"github.com/iotaledger/goshimmer/packages/ledgerstate"
	"github.com/labstack/echo/v4"
	"github.com/pangpanglabs/echoswagger/v2"

	"github.com/iotaledger/wasp/packages/coretypes"
	"github.com/iotaledger/wasp/packages/registry"
	"github.com/iotaledger/wasp/packages/webapi/httperrors"
	"github.com/iotaledger/wasp/packages/webapi/model"
	"github.com/iotaledger/wasp/packages/webapi/routes"
	"github.com/iotaledger/wasp/plugins/chains"
)

func addCommitteeRecordEndpoints(adm echoswagger.ApiGroup) {
	rnd1 := coretypes.RandomChainID()
	example := model.CommitteeRecord{
		Address: model.NewAddress(rnd1.AsAddress()),
		Nodes:   []string{"wasp1.example.org:4000", "wasp2.example.org:4000", "wasp3.example.org:4000"},
	}

	adm.POST(routes.PutCommitteeRecord(), handlePutCommitteeRecord).
		SetSummary("Create a new committee record").
		AddParamBody(example, "Record", "Committee record", true)

	adm.GET(routes.GetCommitteeRecord(":address"), handleGetCommitteeRecord).
		SetSummary("Find the committee record for the given address").
		AddParamPath("", "address", "Address (base58)").
		AddResponse(http.StatusOK, "Committee Record", example, nil)

	adm.GET(routes.GetCommitteeForChain(":chainID"), handleGetCommitteeForChain).
		SetSummary("Find the committee record that manages the given chain").
		AddParamPath("", "chainID", "ChainID (base58)").
		AddResponse(http.StatusOK, "Committee Record", example, nil)
}

func handlePutCommitteeRecord(c echo.Context) error {
	var req model.CommitteeRecord

	if err := c.Bind(&req); err != nil {
		return httperrors.BadRequest("Invalid request body")
	}

	cr := req.Record()

	bd2, err := registry.CommitteeRecordFromRegistry(cr.Address)
	if err != nil {
		return err
	}
	if bd2 != nil {
		return httperrors.Conflict(fmt.Sprintf("Record already exists: %s", cr.Address.Base58()))
	}
	if err = cr.SaveToRegistry(); err != nil {
		return err
	}

	log.Infof("Committee record saved. Address: %s", cr.Address.Base58())

	return c.NoContent(http.StatusCreated)
}

func handleGetCommitteeRecord(c echo.Context) error {
	address, err := ledgerstate.AddressFromBase58EncodedString(c.Param("address"))
	if err != nil {
		return httperrors.BadRequest(err.Error())
	}
	cr, err := registry.CommitteeRecordFromRegistry(address)
	if err != nil {
		return err
	}
	if cr == nil {
		return httperrors.NotFound(fmt.Sprintf("Record not found: %s", address))
	}
	return c.JSON(http.StatusOK, model.NewCommitteeRecord(cr))
}

func handleGetCommitteeForChain(c echo.Context) error {
	chainID, err := coretypes.ChainIDFromBase58(c.Param("chainID"))
	if err != nil {
		return httperrors.BadRequest(err.Error())
	}
	chain := chains.AllChains().Get(chainID)
	if chain == nil {
		return httperrors.NotFound(fmt.Sprintf("Active chain %s not found", chainID))
	}
	committeeInfo := chain.GetCommitteeInfo()
	if committeeInfo == nil {
		return httperrors.NotFound(fmt.Sprintf("Committee info for chain %s is not available", chainID))
	}
	cr, err := registry.CommitteeRecordFromRegistry(committeeInfo.Address)
	if err != nil {
		return err
	}
	if cr == nil {
		return httperrors.NotFound(fmt.Sprintf("Committee record not found for address: %s", committeeInfo.Address.Base58()))
	}
	return c.JSON(http.StatusOK, model.NewCommitteeRecord(cr))
}
