package chain

import (
	"net/http"

	"github.com/pangpanglabs/echoswagger/v2"

	loggerpkg "github.com/iotaledger/hive.go/core/logger"
	"github.com/iotaledger/wasp/packages/authentication"
	"github.com/iotaledger/wasp/packages/authentication/shared/permissions"
	"github.com/iotaledger/wasp/packages/publisher/publisherws"
	"github.com/iotaledger/wasp/packages/webapi/interfaces"
	"github.com/iotaledger/wasp/packages/webapi/models"
)

type Controller struct {
	log *loggerpkg.Logger

	chainService     interfaces.ChainService
	evmService       interfaces.EVMService
	nodeService      interfaces.NodeService
	committeeService interfaces.CommitteeService
	offLedgerService interfaces.OffLedgerService
	registryService  interfaces.RegistryService
	vmService        interfaces.VMService

	webSocketHandler *publisherws.PublisherWebSocket
}

func NewChainController(log *loggerpkg.Logger, chainService interfaces.ChainService, committeeService interfaces.CommitteeService, evmService interfaces.EVMService, nodeService interfaces.NodeService, offLedgerService interfaces.OffLedgerService, registryService interfaces.RegistryService, vmService interfaces.VMService) interfaces.APIController {
	return &Controller{
		log:              log,
		chainService:     chainService,
		evmService:       evmService,
		committeeService: committeeService,
		nodeService:      nodeService,
		offLedgerService: offLedgerService,
		registryService:  registryService,
		vmService:        vmService,
		webSocketHandler: publisherws.New(log, []string{"vmmsg"}),
	}
}

func (c *Controller) Name() string {
	return "chains"
}

func (c *Controller) RegisterPublic(publicAPI echoswagger.ApiGroup, mocker interfaces.Mocker) {
	publicAPI.EchoGroup().Any("chains/:chainID/evm", c.handleJSONRPC)
	publicAPI.GET("chains/:chainID/evm/tx/:txHash", c.getRequestID).
		SetSummary("Get the ISC request ID for the given Ethereum transaction hash").
		SetOperationId("getRequestIDFromEVMTransactionID").
		AddParamPath("", "chainID", "ChainID (Bech32)").
		AddParamPath("", "txHash", "Transaction hash (Hex-encoded)").
		AddResponse(http.StatusOK, "Request ID", "", nil).
		AddResponse(http.StatusNotFound, "Request ID not found", "", nil)

	publicAPI.GET("chains/:chainID/state/:stateKey", c.getState).
		SetSummary("Fetch the raw value associated with the given key in the chain state").
		SetOperationId("getStateValue").
		AddParamPath("", "chainID", "ChainID (Bech32)").
		AddParamPath("", "stateKey", "Key (Hex-encoded)").
		AddResponse(http.StatusOK, "Result", mocker.Get(models.StateResponse{}), nil)

	publicAPI.GET("chains/:chainID/ws", c.handleWebSocket).
		SetOperationId("attachToWebsocket").
		AddParamPath("", "chainID", "ChainID (Bech32-encoded)")
}

func (c *Controller) RegisterAdmin(adminAPI echoswagger.ApiGroup, mocker interfaces.Mocker) {
	adminAPI.GET("chains", c.getChainList, authentication.ValidatePermissions([]string{permissions.Read})).
		AddResponse(http.StatusUnauthorized, "Unauthorized (Wrong permissions, missing token)", authentication.ValidationError{}, nil).
		AddResponse(http.StatusOK, "A list of all available chains", mocker.Get([]models.ChainInfoResponse{}), nil).
		SetOperationId("getChains").
		SetSummary("Get a list of all chains")

	adminAPI.POST("chains/:chainID/activate", c.activateChain, authentication.ValidatePermissions([]string{permissions.Write})).
		AddParamPath("", "chainID", "ChainID (Bech32)").
		AddResponse(http.StatusUnauthorized, "Unauthorized (Wrong permissions, missing token)", authentication.ValidationError{}, nil).
		AddResponse(http.StatusNotModified, "Chain was not activated", nil, nil).
		AddResponse(http.StatusOK, "Chain was successfully activated", nil, nil).
		SetOperationId("activateChain").
		SetSummary("Activate a chain")

	adminAPI.POST("chains/:chainID/deactivate", c.deactivateChain, authentication.ValidatePermissions([]string{permissions.Write})).
		AddParamPath("", "chainID", "ChainID (Bech32)").
		AddResponse(http.StatusUnauthorized, "Unauthorized (Wrong permissions, missing token)", authentication.ValidationError{}, nil).
		AddResponse(http.StatusNotModified, "Chain was not deactivated", nil, nil).
		AddResponse(http.StatusOK, "Chain was successfully deactivated", nil, nil).
		SetOperationId("deactivateChain").
		SetSummary("Deactivate a chain")

	adminAPI.GET("chains/:chainID", c.getChainInfo, authentication.ValidatePermissions([]string{permissions.Read})).
		AddParamPath("", "chainID", "ChainID (Bech32)").
		AddResponse(http.StatusUnauthorized, "Unauthorized (Wrong permissions, missing token)", authentication.ValidationError{}, nil).
		AddResponse(http.StatusOK, "Information about a specific chain", mocker.Get(models.ChainInfoResponse{}), nil).
		SetOperationId("getChainInfo").
		SetSummary("Get information about a specific chain")

	adminAPI.GET("chains/:chainID/committee", c.getCommitteeInfo, authentication.ValidatePermissions([]string{permissions.Read})).
		AddParamPath("", "chainID", "ChainID (Bech32)").
		AddResponse(http.StatusUnauthorized, "Unauthorized (Wrong permissions, missing token)", authentication.ValidationError{}, nil).
		AddResponse(http.StatusOK, "A list of all nodes tied to the chain", mocker.Get(models.CommitteeInfoResponse{}), nil).
		SetOperationId("getCommitteeInfo").
		SetSummary("Get information about the deployed committee")

	adminAPI.GET("chains/:chainID/contracts", c.getContracts, authentication.ValidatePermissions([]string{permissions.Read})).
		AddParamPath("", "chainID", "ChainID (Bech32)").
		AddResponse(http.StatusUnauthorized, "Unauthorized (Wrong permissions, missing token)", authentication.ValidationError{}, nil).
		AddResponse(http.StatusOK, "A list of all available contracts", mocker.Get([]models.ContractInfoResponse{}), nil).
		SetOperationId("getContracts").
		SetSummary("Get all available chain contracts")

	adminAPI.PUT("chains/:chainID/access-node/:publicKey", c.addAccessNode, authentication.ValidatePermissions([]string{permissions.Write})).
		AddParamPath("", "chainID", "ChainID (Bech32)").
		AddParamPath("", "publicKey", "Nodes public key (Hex)").
		AddResponse(http.StatusUnauthorized, "Unauthorized (Wrong permissions, missing token)", authentication.ValidationError{}, nil).
		AddResponse(http.StatusCreated, "Access node was successfully added", nil, nil).
		SetSummary("Configure a trusted node to be an access node.").
		SetOperationId("addAccessNode")

	adminAPI.DELETE("chains/:chainID/access-node/:publicKey", c.removeAccessNode, authentication.ValidatePermissions([]string{permissions.Write})).
		AddParamPath("", "chainID", "ChainID (Bech32)").
		AddParamPath("", "publicKey", "Nodes public key (Hex)").
		AddResponse(http.StatusUnauthorized, "Unauthorized (Wrong permissions, missing token)", authentication.ValidationError{}, nil).
		AddResponse(http.StatusOK, "Access node was successfully removed", nil, nil).
		SetSummary("Remove an access node.").
		SetOperationId("removeAccessNode")
}
