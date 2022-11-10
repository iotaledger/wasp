package requests

import (
	"net/http"

	"github.com/iotaledger/wasp/packages/kv/dict"
	"github.com/iotaledger/wasp/packages/webapi/v2/models"
	"github.com/pangpanglabs/echoswagger/v2"

	loggerpkg "github.com/iotaledger/hive.go/core/logger"
	"github.com/iotaledger/wasp/packages/webapi/v2/interfaces"
)

type Controller struct {
	log *loggerpkg.Logger

	chainService     interfaces.ChainService
	offLedgerService interfaces.OffLedgerService
	peeringService   interfaces.PeeringService
	vmService        interfaces.VMService
}

func NewRequestsController(log *loggerpkg.Logger, chainService interfaces.ChainService, offLedgerService interfaces.OffLedgerService, peeringService interfaces.PeeringService, vmService interfaces.VMService) interfaces.APIController {
	return &Controller{
		log:              log,
		chainService:     chainService,
		offLedgerService: offLedgerService,
		peeringService:   peeringService,
		vmService:        vmService,
	}
}

func (c *Controller) Name() string {
	return "requests"
}

func (c *Controller) RegisterPublic(publicAPI echoswagger.ApiGroup, mocker interfaces.Mocker) {
	dictExample := dict.Dict{
		"key1": []byte("value1"),
	}.JSONDict()

	publicAPI.GET("requests/:chainID/receipt/:requestID", c.getReceipt).
		AddResponse(http.StatusOK, "ReceiptResponse", mocker.Get(models.ReceiptResponse{}), nil).
		SetSummary("Get a receipt from a request ID").
		SetOperationId("getReceipt")

	publicAPI.POST("request/callview", c.executeCallView).
		AddParamBody(dictExample, "body", "Parameters", false).
		AddResponse(http.StatusOK, "Result", dictExample, nil).
		SetSummary("Call a view function on a contract by Hname").
		SetOperationId("callView")

	publicAPI.POST("request/offledger", c.handleOffLedgerRequest).
		AddParamBody(
			models.OffLedgerRequest{Request: "base64 string"},
			"body",
			"Offledger request as JSON. Request encoded in base64.",
			false).
		AddResponse(http.StatusAccepted, "Request submitted", nil, nil).
		SetSummary("Post an off-ledger request").
		SetOperationId("offLedger")

	publicAPI.GET("request/:chainID/request/:requestID/wait", c.waitForRequestToFinish).
		SetSummary("Wait until the given request has been processed by the node").
		AddParamPath("", "chainID", "ChainID (bech32)").
		AddParamPath("", "requestID", "Request ID").
		AddResponse(http.StatusNotFound, "", nil, nil).
		AddResponse(http.StatusRequestTimeout, "", nil, nil).
		AddResponse(http.StatusOK, "Request Receipt", mocker.Get(models.ReceiptResponse{}), nil)

}

func (c *Controller) RegisterAdmin(adminAPI echoswagger.ApiGroup, mocker interfaces.Mocker) {

}
