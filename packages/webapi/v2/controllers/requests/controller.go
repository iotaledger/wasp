package requests

import (
	"net/http"

	"github.com/pangpanglabs/echoswagger/v2"

	loggerpkg "github.com/iotaledger/hive.go/core/logger"
	"github.com/iotaledger/wasp/packages/kv/dict"
	"github.com/iotaledger/wasp/packages/webapi/v2/interfaces"
	"github.com/iotaledger/wasp/packages/webapi/v2/models"
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

	publicAPI.GET("chains/:chainID/receipts/:requestID", c.getReceipt).
		AddParamPath("", "chainID", "ChainID (Bech32)").
		AddParamPath("", "requestID", "RequestID (Hex)").
		AddResponse(http.StatusOK, "ReceiptResponse", mocker.Get(models.ReceiptResponse{}), nil).
		SetSummary("Get a receipt from a request ID").
		SetOperationId("getReceipt")

	publicAPI.POST("requests/callview", c.executeCallView).
		AddParamBody(mocker.Get(models.ContractCallViewRequest{}), "", "Parameters", false).
		AddResponse(http.StatusOK, "Result", dictExample, nil).
		SetSummary("Call a view function on a contract by Hname").
		SetDescription("Execute a view call. Either use HName or Name properties. If both are supplied, HName are used.").
		SetOperationId("callView")

	publicAPI.POST("requests/offledger", c.handleOffLedgerRequest).
		AddParamBody(
			models.OffLedgerRequest{Request: "Hex string"},
			"",
			"Offledger request as JSON. Request encoded in Hex",
			false).
		AddResponse(http.StatusAccepted, "Request submitted", nil, nil).
		SetSummary("Post an off-ledger request").
		SetOperationId("offLedger")

	publicAPI.GET("chains/:chainID/requests/:requestID/wait", c.waitForRequestToFinish).
		SetSummary("Wait until the given request has been processed by the node").
		SetOperationId("waitForTransaction").
		AddParamPath("", "chainID", "ChainID (Bech32)").
		AddParamPath("", "requestID", "RequestID (Hex)").
		AddResponse(http.StatusNotFound, "The chain or request id is invalid", nil, nil).
		AddResponse(http.StatusRequestTimeout, "The waiting time has reached the defined limit", nil, nil).
		AddResponse(http.StatusOK, "The request receipt", mocker.Get(models.ReceiptResponse{}), nil)
}

func (c *Controller) RegisterAdmin(adminAPI echoswagger.ApiGroup, mocker interfaces.Mocker) {
}
