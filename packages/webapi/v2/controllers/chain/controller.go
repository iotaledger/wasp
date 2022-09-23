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

type ChainController struct {
	log *loggerpkg.Logger

	chainService     interfaces.Chain
	nodeService      interfaces.Node
	offLedgerService interfaces.OffLedger
	registryService  interfaces.Registry
	vmService        interfaces.VM
}

func NewChainController(log *loggerpkg.Logger, chainService interfaces.Chain, nodeService interfaces.Node, offLedgerService interfaces.OffLedger, registryService interfaces.Registry, vmService interfaces.VM) interfaces.APIController {
	return &ChainController{
		log:              log,
		chainService:     chainService,
		nodeService:      nodeService,
		offLedgerService: offLedgerService,
		registryService:  registryService,
		vmService:        vmService,
	}
}

func (c *ChainController) Name() string {
	return "chains"
}

func (c *ChainController) RegisterExampleData(mock interfaces.Mocker) {
	mock.AddModel(&models.ChainInfoResponse{})
	mock.AddModel(&models.ContractListResponse{})
	mock.AddModel(&models.CommitteeInfoResponse{})
	mock.AddModel(&models.ChainListResponse{})
	mock.AddModel(&models.ContractInfoResponse{})
}

func (c *ChainController) RegisterPublic(publicAPI echoswagger.ApiGroup, mocker interfaces.Mocker) {
	dictExample := dict.Dict{
		"key1": []byte("value1"),
	}.JSONDict()

	publicAPI.POST(routes.CallViewByName(":chainID", ":contractName", ":functionName"), c.callViewByContractName).
		SetSummary("Call a view function on a contract by name").
		AddParamPath("", "chainID", "ChainID (Bech32)").
		AddParamPath("", "contractName", "Contract Name").
		AddParamPath("", "functionName", "Function name").
		AddParamBody(dictExample, "params", "Parameters", false).
		AddResponse(http.StatusOK, "Result", dictExample, nil)

	publicAPI.POST(routes.CallViewByHname(":chainID", ":contractHName", ":functionHName"), c.callViewByHName).
		SetSummary("Call a view function on a contract by Hname").
		AddParamPath("", "chainID", "ChainID (Bech32").
		AddParamPath("", "contractHName", "Contract Hname").
		AddParamPath("getInfo", "functionHName", "Function Hname").
		AddParamBody(dictExample, "params", "Parameters", false).
		AddResponse(http.StatusOK, "Result", dictExample, nil)

	publicAPI.POST(routes.NewRequest(":chainID"), c.handleNewRequest).
		SetSummary("Post an off-ledger request").
		AddParamPath("", "chainID", "chainID represented in base58").
		AddParamBody(
			models.OffLedgerRequestBody{Request: "base64 string"},
			"Request",
			"Offledger Request encoded in base64. Optionally, the body can be the binary representation of the offledger request, but mime-type must be specified to \"application/octet-stream\"",
			false).
		AddResponse(http.StatusAccepted, "Request submitted", nil, nil)
}

func (c *ChainController) RegisterAdmin(adminAPI echoswagger.ApiGroup, mocker interfaces.Mocker) {
	adminAPI.POST(routes.ActivateChain(":chainID"), c.activateChain).
		AddParamPath("", "chainID", "ChainID (Bech32)").
		AddResponse(http.StatusOK, "Chain was successfully activated", nil, nil).
		AddResponse(http.StatusNotModified, "Chain was not activated", nil, nil).
		SetOperationId("activateChain").
		SetSummary("Activate a chain")

	adminAPI.POST(routes.DeactivateChain(":chainID"), c.deactivateChain).
		AddParamPath("", "chainID", "ChainID (Bech32)").
		AddResponse(http.StatusOK, "Chain was successfully deactivated", nil, nil).
		AddResponse(http.StatusNotModified, "Chain was not deactivated", nil, nil).
		SetOperationId("deactivateChain").
		SetSummary("Deactivate a chain")

	adminAPI.GET(routes.GetChainCommitteeInfo(":chainID"), c.getCommitteeInfo).
		AddParamPath("", "chainID", "ChainID (Bech32)").
		AddResponse(http.StatusOK, "A list of all nodes tied to the chain.", mocker.GetMockedStruct(models.CommitteeInfoResponse{}), nil).
		SetOperationId("getChainCommitteeInfo").
		SetSummary("Get basic chain info.")

	adminAPI.GET(routes.GetChainList(), c.getChainList).
		AddResponse(http.StatusOK, "A list of all available chains.", mocker.GetMockedStruct(models.ChainListResponse{}), nil).
		SetOperationId("getChainList").
		SetSummary("Get a list of all chains.")

	adminAPI.GET(routes.GetChainInfo(":chainID"), c.getChainInfo).
		AddParamPath("", "chainID", "ChainID (Bech32)").
		AddResponse(http.StatusOK, "Information about a specific chain.", mocker.GetMockedStruct(models.ChainInfoResponse{}), nil).
		SetOperationId("getChainInfo").
		SetSummary("Get information about a specific chain.")

	adminAPI.GET(routes.GetChainContracts(":chainID"), c.getContracts).
		AddParamPath("", "chainID", "ChainID (Bech32)").
		AddResponse(http.StatusOK, "A list of all available contracts.", mocker.GetMockedStruct(models.ContractListResponse{}), nil).
		SetOperationId("getChainContracts").
		SetSummary("Get all available chain contracts.")
}
