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
