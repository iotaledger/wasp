package chain

import (
	"net/http"

	"github.com/pangpanglabs/echoswagger/v2"

	loggerpkg "github.com/iotaledger/hive.go/logger"
	"github.com/iotaledger/wasp/packages/authentication"
	"github.com/iotaledger/wasp/packages/authentication/shared/permissions"
	"github.com/iotaledger/wasp/packages/kv/dict"
	"github.com/iotaledger/wasp/packages/webapi/interfaces"
	"github.com/iotaledger/wasp/packages/webapi/models"
	"github.com/iotaledger/wasp/packages/webapi/params"
	"github.com/iotaledger/wasp/packages/webapi/routes"
)

type Controller struct {
	log *loggerpkg.Logger

	chainService     interfaces.ChainService
	evmService       interfaces.EVMService
	nodeService      interfaces.NodeService
	committeeService interfaces.CommitteeService
	offLedgerService interfaces.OffLedgerService
	registryService  interfaces.RegistryService
}

func NewChainController(log *loggerpkg.Logger,
	chainService interfaces.ChainService,
	committeeService interfaces.CommitteeService,
	evmService interfaces.EVMService,
	nodeService interfaces.NodeService,
	offLedgerService interfaces.OffLedgerService,
	registryService interfaces.RegistryService,
) interfaces.APIController {
	return &Controller{
		log:              log,
		chainService:     chainService,
		evmService:       evmService,
		committeeService: committeeService,
		nodeService:      nodeService,
		offLedgerService: offLedgerService,
		registryService:  registryService,
	}
}

func (c *Controller) Name() string {
	return "chains"
}

func (c *Controller) RegisterPublic(publicAPI echoswagger.ApiGroup, mocker interfaces.Mocker) {
	publicAPI.GET("chains/:chainID", c.getChainInfo).
		AddParamPath("", params.ParamChainID, params.DescriptionChainID).
		AddParamQuery("", params.ParamBlockIndexOrTrieRoot, params.DescriptionBlockIndexOrTrieRoot, false).
		AddResponse(http.StatusOK, "Information about a specific chain", mocker.Get(models.ChainInfoResponse{}), nil).
		SetOperationId("getChainInfo").
		SetSummary("Get information about a specific chain")

	// Echoswagger does not support ANY, so create a fake route, and overwrite it with Echo ANY afterwords.
	publicAPI.
		POST("chains/:chainID/"+routes.EVMJsonRPCPathSuffix, c.handleJSONRPC).
		AddParamPath("", params.ParamChainID, params.DescriptionChainID).
		SetSummary("Ethereum JSON-RPC")

	publicAPI.
		EchoGroup().Any("chains/:chainID/"+routes.EVMJsonRPCPathSuffix, c.handleJSONRPC)

	publicAPI.
		GET("chains/:chainID/"+routes.EVMJsonWebSocketPathSuffix, c.handleWebsocket).
		AddParamPath("", params.ParamChainID, params.DescriptionChainID).
		SetSummary("Ethereum JSON-RPC (Websocket transport)")

	publicAPI.GET("chains/:chainID/state/:stateKey", c.getState).
		AddParamPath("", params.ParamChainID, params.DescriptionChainID).
		AddParamPath("", params.ParamStateKey, params.DescriptionStateKey).
		AddResponse(http.StatusOK, "Result", mocker.Get(models.StateResponse{}), nil).
		SetSummary("Fetch the raw value associated with the given key in the chain state").
		SetOperationId("getStateValue")

	publicAPI.GET("chains/:chainID/receipts/:requestID", c.getReceipt).
		AddParamPath("", params.ParamChainID, params.DescriptionChainID).
		AddParamPath("", params.ParamRequestID, params.DescriptionRequestID).
		AddResponse(http.StatusNotFound, "Chain or request id not found", nil, nil).
		AddResponse(http.StatusOK, "ReceiptResponse", mocker.Get(models.ReceiptResponse{}), nil).
		SetSummary("Get a receipt from a request ID").
		SetOperationId("getReceipt")

	dictExample := dict.Dict{
		"key1": []byte("value1"),
	}.JSONDict()

	publicAPI.POST("chains/:chainID/callview", c.executeCallView).
		AddParamPath("", params.ParamChainID, params.DescriptionChainID).
		AddParamBody(mocker.Get(models.ContractCallViewRequest{}), "", "Parameters", true).
		AddResponse(http.StatusOK, "Result", dictExample, nil).
		SetSummary("Call a view function on a contract by Hname").
		SetDescription("Execute a view call. Either use HName or Name properties. If both are supplied, HName are used.").
		SetOperationId("callView")

	publicAPI.POST("chains/:chainID/estimategas-onledger", c.estimateGasOnLedger).
		AddParamPath("", params.ParamChainID, params.DescriptionChainID).
		AddParamBody(mocker.Get(models.EstimateGasRequestOnledger{}), "Request", "Request", true).
		AddResponse(http.StatusOK, "ReceiptResponse", mocker.Get(models.ReceiptResponse{}), nil).
		SetSummary("Estimates gas for a given on-ledger ISC request").
		SetOperationId("estimateGasOnledger")

	publicAPI.POST("chains/:chainID/estimategas-offledger", c.estimateGasOffLedger).
		AddParamPath("", params.ParamChainID, params.DescriptionChainID).
		AddParamBody(mocker.Get(models.EstimateGasRequestOffledger{}), "Request", "Request", true).
		AddResponse(http.StatusOK, "ReceiptResponse", mocker.Get(models.ReceiptResponse{}), nil).
		SetSummary("Estimates gas for a given off-ledger ISC request").
		SetOperationId("estimateGasOffledger")

	publicAPI.GET("chains/:chainID/requests/:requestID/wait", c.waitForRequestToFinish).
		SetSummary("Wait until the given request has been processed by the node").
		SetOperationId("waitForRequest").
		AddParamPath("", params.ParamChainID, params.DescriptionChainID).
		AddParamPath("", params.ParamRequestID, params.DescriptionRequestID).
		AddParamQuery(0, "timeoutSeconds", "The timeout in seconds, maximum 60s", false).
		AddParamQuery(false, "waitForL1Confirmation", "Wait for the block to be confirmed on L1", false).
		AddResponse(http.StatusNotFound, "The chain or request id not found", nil, nil).
		AddResponse(http.StatusRequestTimeout, "The waiting time has reached the defined limit", nil, nil).
		AddResponse(http.StatusOK, "The request receipt", mocker.Get(models.ReceiptResponse{}), nil)
}

func (c *Controller) RegisterAdmin(adminAPI echoswagger.ApiGroup, mocker interfaces.Mocker) {
	adminAPI.GET("chains", c.getChainList, authentication.ValidatePermissions([]string{permissions.Read})).
		AddResponse(http.StatusOK, "A list of all available chains", mocker.Get([]models.ChainInfoResponse{}), nil).
		SetOperationId("getChains").
		SetSummary("Get a list of all chains")

	adminAPI.POST("chains/:chainID/activate", c.activateChain, authentication.ValidatePermissions([]string{permissions.Write})).
		AddParamPath("", params.ParamChainID, params.DescriptionChainID).
		AddResponse(http.StatusNotModified, "Chain was not activated", nil, nil).
		AddResponse(http.StatusOK, "Chain was successfully activated", nil, nil).
		SetOperationId("activateChain").
		SetSummary("Activate a chain")

	adminAPI.POST("chains/:chainID/deactivate", c.deactivateChain, authentication.ValidatePermissions([]string{permissions.Write})).
		AddParamPath("", params.ParamChainID, params.DescriptionChainID).
		AddResponse(http.StatusNotModified, "Chain was not deactivated", nil, nil).
		AddResponse(http.StatusOK, "Chain was successfully deactivated", nil, nil).
		SetOperationId("deactivateChain").
		SetSummary("Deactivate a chain")

	adminAPI.GET("chains/:chainID/committee", c.getCommitteeInfo, authentication.ValidatePermissions([]string{permissions.Read})).
		AddParamPath("", params.ParamChainID, params.DescriptionChainID).
		AddParamQuery("", params.ParamBlockIndexOrTrieRoot, params.DescriptionBlockIndexOrTrieRoot, false).
		AddResponse(http.StatusOK, "A list of all nodes tied to the chain", mocker.Get(models.CommitteeInfoResponse{}), nil).
		SetOperationId("getCommitteeInfo").
		SetSummary("Get information about the deployed committee")

	adminAPI.GET("chains/:chainID/contracts", c.getContracts, authentication.ValidatePermissions([]string{permissions.Read})).
		AddParamPath("", params.ParamChainID, params.DescriptionChainID).
		AddParamQuery("", params.ParamBlockIndexOrTrieRoot, params.DescriptionBlockIndexOrTrieRoot, false).
		AddResponse(http.StatusOK, "A list of all available contracts", mocker.Get([]models.ContractInfoResponse{}), nil).
		SetOperationId("getContracts").
		SetSummary("Get all available chain contracts")

	adminAPI.POST("chains/:chainID/chainrecord", c.setChainRecord, authentication.ValidatePermissions([]string{permissions.Write})).
		AddParamPath("", params.ParamChainID, params.DescriptionChainID).
		AddParamBody(mocker.Get(models.ChainRecord{}), "ChainRecord", "Chain Record", true).
		AddResponse(http.StatusCreated, "Chain record was saved", nil, nil).
		SetSummary("Sets the chain record.").
		SetOperationId("setChainRecord")

	adminAPI.PUT("chains/:chainID/access-node/:peer", c.addAccessNode, authentication.ValidatePermissions([]string{permissions.Write})).
		AddParamPath("", params.ParamChainID, params.DescriptionChainID).
		AddParamPath("", params.ParamPeer, params.DescriptionPeer).
		AddResponse(http.StatusCreated, "Access node was successfully added", nil, nil).
		SetSummary("Configure a trusted node to be an access node.").
		SetOperationId("addAccessNode")

	adminAPI.DELETE("chains/:chainID/access-node/:peer", c.removeAccessNode, authentication.ValidatePermissions([]string{permissions.Write})).
		AddParamPath("", params.ParamChainID, params.DescriptionChainID).
		AddParamPath("", params.ParamPeer, params.DescriptionPeer).
		AddResponse(http.StatusOK, "Access node was successfully removed", nil, nil).
		SetSummary("Remove an access node.").
		SetOperationId("removeAccessNode")

	adminAPI.GET("chains/:chainID/mempool", c.getMempoolContents, authentication.ValidatePermissions([]string{permissions.Read})).
		AddParamPath("", params.ParamChainID, params.DescriptionChainID).
		SetResponseContentType("application/octet-stream").
		AddResponse(http.StatusOK, "stream of JSON representation of the requests in the mempool", []byte{}, nil).
		SetSummary("Get the contents of the mempool.").
		SetOperationId("getMempoolContents")
}
