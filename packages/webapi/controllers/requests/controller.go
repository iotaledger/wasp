package requests

import (
	"net/http"

	"github.com/pangpanglabs/echoswagger/v2"

	"github.com/iotaledger/wasp/v2/packages/webapi/interfaces"
	"github.com/iotaledger/wasp/v2/packages/webapi/models"
)

type Controller struct {
	chainService     interfaces.ChainService
	offLedgerService interfaces.OffLedgerService
	peeringService   interfaces.PeeringService
}

func NewRequestsController(chainService interfaces.ChainService, offLedgerService interfaces.OffLedgerService, peeringService interfaces.PeeringService) interfaces.APIController {
	return &Controller{
		chainService:     chainService,
		offLedgerService: offLedgerService,
		peeringService:   peeringService,
	}
}

func (c *Controller) Name() string {
	return "requests"
}

func (c *Controller) RegisterPublic(publicAPI echoswagger.ApiGroup, mocker interfaces.Mocker) {
	publicAPI.POST("requests/offledger", c.handleOffLedgerRequest).
		AddParamBody(
			models.OffLedgerRequest{Request: "Hex string"},
			"",
			"Offledger request as JSON. Request encoded in Hex",
			true).
		AddResponse(http.StatusAccepted, "Request submitted", nil, nil).
		SetSummary("Post an off-ledger request").
		SetOperationId("offLedger")
}

func (c *Controller) RegisterAdmin(adminAPI echoswagger.ApiGroup, mocker interfaces.Mocker) {
}
