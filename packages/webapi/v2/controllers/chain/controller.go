package chain

import (
	"net/http"

	"github.com/iotaledger/wasp/packages/webapi/v2/models"

	"github.com/iotaledger/wasp/packages/kv/dict"

	"github.com/pangpanglabs/echoswagger/v2"

	loggerpkg "github.com/iotaledger/hive.go/core/logger"
	"github.com/iotaledger/wasp/packages/webapi/v2/interfaces"
	"github.com/iotaledger/wasp/packages/webapi/v2/routes"
)

type Controller struct {
	log *loggerpkg.Logger

	chainService     interfaces.Chain
	nodeService      interfaces.Node
	offLedgerService interfaces.OffLedger
	registryService  interfaces.Registry
	vmService        interfaces.VM
}

func NewChainController(log *loggerpkg.Logger, chainService interfaces.Chain, nodeService interfaces.Node, offLedgerService interfaces.OffLedger, registryService interfaces.Registry, vmService interfaces.VM) interfaces.APIController {
	return &Controller{
		log:              log,
		chainService:     chainService,
		nodeService:      nodeService,
		offLedgerService: offLedgerService,
		registryService:  registryService,
		vmService:        vmService,
	}
}

func (c *Controller) Name() string {
	return "chains"
}

func (c *Controller) RegisterExampleData(mock interfaces.Mocker) {
	mock.AddModel(&models.ChainInfoResponse{})
	mock.AddModel(&models.ContractListResponse{})
	mock.AddModel(&models.CommitteeInfoResponse{})
	mock.AddModel(&models.ChainListResponse{})
	mock.AddModel(&models.ContractInfoResponse{})
}

func (c *Controller) RegisterPublic(publicAPI echoswagger.ApiGroup, mocker interfaces.Mocker) {
	dictExample := dict.Dict{
		"key1": []byte("value1"),
	}.JSONDict()

	publicAPI.POST(routes.CallViewByName(":chainID", ":contractName", ":functionName"), c.callViewByContractName).
		AddParamBody(dictExample, "body", "Parameters", false).
		AddParamPath("", "chainID", "ChainID (Bech32)").
		AddParamPath("", "contractName", "Contract Name").
		AddParamPath("", "functionName", "Function name").
		AddResponse(http.StatusOK, "Result", dictExample, nil).
		SetResponseContentType("application/json").
		SetSummary("Call a view function on a contract by name")

	publicAPI.POST(routes.CallViewByHname(":chainID", ":contractHName", ":functionHName"), c.callViewByHName).
		AddParamBody(dictExample, "body", "Parameters", false).
		AddParamPath("", "chainID", "ChainID (Bech32").
		AddParamPath("", "contractHName", "Contract Hname").
		AddParamPath("", "functionHName", "Function Hname").
		AddResponse(http.StatusOK, "Result", dictExample, nil).
		SetResponseContentType("application/json").
		SetSummary("Call a view function on a contract by Hname")

	publicAPI.POST(routes.NewRequest(":chainID"), c.handleNewRequest).
		AddParamBody(
			models.OffLedgerRequestBody{Request: "base64 string"},
			"body",
			"Offledger request as JSON. Request encoded in base64.",
			false).
		AddParamPath("", "chainID", "ChainID (Bech32)").
		AddResponse(http.StatusAccepted, "Request submitted", nil, nil).
		SetResponseContentType("application/json").
		SetSummary("Post an off-ledger request")
}

func (c *Controller) RegisterAdmin(adminAPI echoswagger.ApiGroup, mocker interfaces.Mocker) {
	adminAPI.POST(routes.ActivateChain(":chainID"), c.activateChain).
		AddParamPath("", "chainID", "ChainID (Bech32)").
		AddResponse(http.StatusNotModified, "Chain was not activated", nil, nil).
		AddResponse(http.StatusOK, "Chain was successfully activated", nil, nil).
		SetOperationId("activateChain").
		SetResponseContentType("application/json").
		SetSummary("Activate a chain")

	adminAPI.POST(routes.DeactivateChain(":chainID"), c.deactivateChain).
		AddParamPath("", "chainID", "ChainID (Bech32)").
		AddResponse(http.StatusNotModified, "Chain was not deactivated", nil, nil).
		AddResponse(http.StatusOK, "Chain was successfully deactivated", nil, nil).
		SetOperationId("deactivateChain").
		SetResponseContentType("application/json").
		SetSummary("Deactivate a chain")

	adminAPI.GET(routes.GetChainCommitteeInfo(":chainID"), c.getCommitteeInfo).
		AddParamPath("", "chainID", "ChainID (Bech32)").
		AddResponse(http.StatusOK, "A list of all nodes tied to the chain.", mocker.GetMockedStruct(models.CommitteeInfoResponse{}), nil).
		SetOperationId("getChainCommitteeInfo").
		SetResponseContentType("application/json").
		SetSummary("Get basic chain info.")

	adminAPI.GET(routes.GetChainList(), c.getChainList).
		AddResponse(http.StatusOK, "A list of all available chains.", mocker.GetMockedStruct(models.ChainListResponse{}), nil).
		SetOperationId("getChainList").
		SetResponseContentType("application/json").
		SetSummary("Get a list of all chains.")

	adminAPI.GET(routes.GetChainInfo(":chainID"), c.getChainInfo).
		AddParamPath("", "chainID", "ChainID (Bech32)").
		AddResponse(http.StatusOK, "Information about a specific chain.", mocker.GetMockedStruct(models.ChainInfoResponse{}), nil).
		SetOperationId("getChainInfo").
		SetResponseContentType("application/json").
		SetSummary("Get information about a specific chain.")

	adminAPI.GET(routes.GetChainContracts(":chainID"), c.getContracts).
		AddParamPath("", "chainID", "ChainID (Bech32)").
		AddResponse(http.StatusOK, "A list of all available contracts.", mocker.GetMockedStruct(models.ContractListResponse{}), nil).
		SetOperationId("getChainContracts").
		SetResponseContentType("application/json").
		SetSummary("Get all available chain contracts.")
}
