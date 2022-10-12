package chain

import (
	"net/http"

	"github.com/iotaledger/wasp/packages/webapi/v2/models"

	"github.com/pangpanglabs/echoswagger/v2"

	loggerpkg "github.com/iotaledger/hive.go/core/logger"
	"github.com/iotaledger/wasp/packages/webapi/v2/interfaces"
)

type Controller struct {
	log *loggerpkg.Logger

	chainService     interfaces.ChainService
	evmService       interfaces.EVMService
	committeeService interfaces.CommitteeService
	offLedgerService interfaces.OffLedgerService
	registryService  interfaces.RegistryService
	vmService        interfaces.VMService
}

func NewChainController(log *loggerpkg.Logger, chainService interfaces.ChainService, committeeService interfaces.CommitteeService, evmService interfaces.EVMService, offLedgerService interfaces.OffLedgerService, registryService interfaces.RegistryService, vmService interfaces.VMService) interfaces.APIController {
	return &Controller{
		log:              log,
		chainService:     chainService,
		evmService:       evmService,
		committeeService: committeeService,
		offLedgerService: offLedgerService,
		registryService:  registryService,
		vmService:        vmService,
	}
}

func (c *Controller) Name() string {
	return "chains"
}

func (c *Controller) RegisterPublic(publicAPI echoswagger.ApiGroup, mocker interfaces.Mocker) {
	publicAPI.EchoGroup().Any("chains/:chainID/evm", c.handleJSONRPC)
	publicAPI.GET("chains/:chainID/evm/tx/:txHash", c.getRequestID).
		SetSummary("Get the ISC request ID for the given Ethereum transaction hash").
		AddParamPath("", "chainID", "ChainID (bech32-encoded)").
		AddParamPath("", "txHash", "Transaction hash (hex-encoded)").
		AddResponse(http.StatusOK, "Request ID", "", nil).
		AddResponse(http.StatusNotFound, "Request ID not found", "", nil)

	publicAPI.GET("chains/:chainID/state/:stateKey", c.getState).
		SetSummary("Fetch the raw value associated with the given key in the chain state").
		AddParamPath("", "chainID", "ChainID (bech32-encoded)").
		AddParamPath("", "stateKey", "Key (hex-encoded)").
		AddResponse(http.StatusOK, "Result", []byte("value"), nil)
}

func (c *Controller) RegisterAdmin(adminAPI echoswagger.ApiGroup, mocker interfaces.Mocker) {
	adminAPI.GET("chains", c.getChainList).
		AddResponse(http.StatusOK, "A list of all available chains.", mocker.Get([]models.ChainInfo{}), nil).
		SetOperationId("getChainList").
		SetResponseContentType("application/json").
		SetSummary("Get a list of all chains.")

	adminAPI.PUT("chains", c.saveChain).
		AddParamBody(&models.SaveChainRecordRequest{}, "body", "The save chain request", true).
		AddResponse(http.StatusNotModified, "ChainService was not saved", nil, nil).
		AddResponse(http.StatusOK, "ChainService was saved", nil, nil).
		SetOperationId("saveChain").
		SetSummary("Saves a chain")

	adminAPI.POST("chains/:chainID/activate", c.activateChain).
		AddParamPath("", "chainID", "ChainID (Bech32)").
		AddResponse(http.StatusNotModified, "ChainService was not activated", nil, nil).
		AddResponse(http.StatusOK, "ChainService was successfully activated", nil, nil).
		SetOperationId("activateChain").
		SetSummary("Activate a chain")

	adminAPI.POST("chains/:chainID/deactivate", c.deactivateChain).
		AddParamPath("", "chainID", "ChainID (Bech32)").
		AddResponse(http.StatusNotModified, "ChainService was not deactivated", nil, nil).
		AddResponse(http.StatusOK, "ChainService was successfully deactivated", nil, nil).
		SetOperationId("deactivateChain").
		SetSummary("Deactivate a chain")

	adminAPI.GET("chains/:chainID", c.getChainInfo).
		AddParamPath("", "chainID", "ChainID (Bech32)").
		AddResponse(http.StatusOK, "Information about a specific chain.", mocker.Get(models.ChainInfo{}), nil).
		SetOperationId("getChainInfo").
		SetResponseContentType("application/json").
		SetSummary("Get information about a specific chain.")

	adminAPI.GET("chains/:chainID/committee", c.getCommitteeInfo).
		AddParamPath("", "chainID", "ChainID (Bech32)").
		AddResponse(http.StatusOK, "A list of all nodes tied to the chain.", mocker.Get(models.CommitteeInfo{}), nil).
		SetOperationId("getChainCommitteeInfo").
		SetResponseContentType("application/json").
		SetSummary("Get basic chain info.")

	adminAPI.GET("chains/:chainID/contracts", c.getContracts).
		AddParamPath("", "chainID", "ChainID (Bech32)").
		AddResponse(http.StatusOK, "A list of all available contracts.", mocker.Get([]models.ContractInfo{}), nil).
		SetOperationId("getChainContracts").
		SetResponseContentType("application/json").
		SetSummary("Get all available chain contracts.")
}
