package corecontracts

import (
	"net/http"

	"github.com/pangpanglabs/echoswagger/v2"

	"github.com/iotaledger/wasp/packages/authentication"
	corecontracts2 "github.com/iotaledger/wasp/packages/webapi/v2/corecontracts"
	"github.com/iotaledger/wasp/packages/webapi/v2/interfaces"
	"github.com/iotaledger/wasp/packages/webapi/v2/models"
)

type Controller struct {
	accounts   *corecontracts2.Accounts
	blob       *corecontracts2.Blob
	blocklog   *corecontracts2.BlockLog
	errors     *corecontracts2.Errors
	governance *corecontracts2.Governance
	vmService  interfaces.VMService
}

func NewCoreContractsController(vmService interfaces.VMService) interfaces.APIController {
	return &Controller{
		accounts:   corecontracts2.NewAccounts(vmService),
		blob:       corecontracts2.NewBlob(vmService),
		blocklog:   corecontracts2.NewBlockLog(vmService),
		errors:     corecontracts2.NewErrors(vmService),
		governance: corecontracts2.NewGovernance(vmService),
		vmService:  vmService,
	}
}

func (c *Controller) Name() string {
	return "corecontracts"
}

func (c *Controller) addAccountContractRoutes(api echoswagger.ApiGroup, mocker interfaces.Mocker) {
	api.GET("chains/:chainID/core/accounts", c.getAccounts).
		AddParamPath("", "chainID", "ChainID (Bech32)").
		AddResponse(http.StatusUnauthorized, "Unauthorized (Wrong permissions, missing token)", authentication.ValidationError{}, nil).
		AddResponse(http.StatusOK, "A list of all accounts", mocker.Get(models.AccountListResponse{}), nil).
		SetOperationId("accountsGetAccounts").
		SetSummary("Get a list of all accounts")

	api.GET("chains/:chainID/core/accounts/account/:agentID/balance", c.getAccountBalance).
		AddParamPath("", "chainID", "ChainID (Bech32)").
		AddParamPath("", "agentID", "AgentID (Bech32 for WasmVM | Hex for EVM)").
		AddResponse(http.StatusUnauthorized, "Unauthorized (Wrong permissions, missing token)", authentication.ValidationError{}, nil).
		AddResponse(http.StatusOK, "All assets belonging to an account", mocker.Get(models.AssetsResponse{}), nil).
		SetOperationId("accountsGetAccountBalance").
		SetSummary("Get all assets belonging to an account")

	api.GET("chains/:chainID/core/accounts/account/:agentID/nfts", c.getAccountNFTs).
		AddParamPath("", "chainID", "ChainID (Bech32)").
		AddParamPath("", "agentID", "AgentID (Bech32 for WasmVM | Hex for EVM)").
		AddResponse(http.StatusUnauthorized, "Unauthorized (Wrong permissions, missing token)", authentication.ValidationError{}, nil).
		AddResponse(http.StatusOK, "All NFT ids belonging to an account", mocker.Get(models.AccountNFTsResponse{}), nil).
		SetOperationId("accountsGetAccountNFTIDs").
		SetSummary("Get all NFT ids belonging to an account")

	api.GET("chains/:chainID/core/accounts/account/:agentID/nonce", c.getAccountNonce).
		AddParamPath("", "chainID", "ChainID (Bech32)").
		AddParamPath("", "agentID", "AgentID (Bech32 for WasmVM | Hex for EVM | '000000@Bech32' Addresses require urlencode)").
		AddResponse(http.StatusUnauthorized, "Unauthorized (Wrong permissions, missing token)", authentication.ValidationError{}, nil).
		AddResponse(http.StatusOK, "The current nonce of an account", mocker.Get(models.AccountNonceResponse{}), nil).
		SetOperationId("accountsGetAccountNonce").
		SetSummary("Get the current nonce of an account")

	api.GET("chains/:chainID/core/accounts/nftdata", c.getNFTData).
		AddParamPath("", "chainID", "ChainID (Bech32)").
		AddParamPath("", "nftID", "NFT ID (Hex)").
		AddResponse(http.StatusUnauthorized, "Unauthorized (Wrong permissions, missing token)", authentication.ValidationError{}, nil).
		AddResponse(http.StatusOK, "The NFT data", mocker.Get(models.NFTDataResponse{}), nil).
		SetOperationId("accountsGetNFTData").
		SetSummary("Get the NFT data by an ID")

	api.GET("chains/:chainID/core/accounts/token_registry", c.getNativeTokenIDRegistry).
		AddParamPath("", "chainID", "ChainID (Bech32)").
		AddResponse(http.StatusUnauthorized, "Unauthorized (Wrong permissions, missing token)", authentication.ValidationError{}, nil).
		AddResponse(http.StatusOK, "A list of all registries", mocker.Get(models.NativeTokenIDRegistryResponse{}), nil).
		SetOperationId("accountsGetNativeTokenIDRegistry").
		SetSummary("Get a list of all registries")

	api.GET("chains/:chainID/core/accounts/foundry_output", c.getFoundryOutput).
		AddParamPath("", "chainID", "ChainID (Bech32)").
		AddParamPath("", "chainID", "Serial Number (uint32)").
		AddResponse(http.StatusUnauthorized, "Unauthorized (Wrong permissions, missing token)", authentication.ValidationError{}, nil).
		AddResponse(http.StatusOK, "The foundry output", mocker.Get(models.FoundryOutputResponse{}), nil).
		SetOperationId("accountsGetFoundryOutput").
		SetSummary("Get the foundry output")

	api.GET("chains/:chainID/core/accounts/total_assets", c.getTotalAssets).
		AddParamPath("", "chainID", "ChainID (Bech32)").
		AddResponse(http.StatusUnauthorized, "Unauthorized (Wrong permissions, missing token)", authentication.ValidationError{}, nil).
		AddResponse(http.StatusOK, "All stored assets", mocker.Get(models.AssetsResponse{}), nil).
		SetOperationId("accountsGetTotalAssets").
		SetSummary("Get all stored assets")
}

func (c *Controller) addBlobContractRoutes(api echoswagger.ApiGroup, mocker interfaces.Mocker) {
	api.GET("chains/:chainID/core/blobs", c.listBlobs).
		AddParamPath("", "chainID", "ChainID (Bech32)").
		AddResponse(http.StatusUnauthorized, "Unauthorized (Wrong permissions, missing token)", authentication.ValidationError{}, nil).
		AddResponse(http.StatusOK, "All stored blobs", mocker.Get(BlobListResponse{}), nil).
		SetOperationId("blobsGetAllBlobs").
		SetSummary("Get all stored blobs")

	api.GET("chains/:chainID/core/blobs/:blobHash/data/:fieldKey", c.getBlobValue).
		AddParamPath("", "chainID", "ChainID (Bech32)").
		AddParamPath("", "blobHash", "BlobHash (Hex)").
		AddParamPath("", "fieldKey", "FieldKey (String)").
		AddResponse(http.StatusUnauthorized, "Unauthorized (Wrong permissions, missing token)", authentication.ValidationError{}, nil).
		AddResponse(http.StatusOK, "The value of the supplied field (key)", mocker.Get(BlobValueResponse{}), nil).
		SetOperationId("blobsGetBlobValue").
		SetSummary("Get the value of the supplied field (key)")

	api.GET("chains/:chainID/core/blobs/:blobHash", c.getBlobInfo).
		AddParamPath("", "chainID", "ChainID (Bech32)").
		AddParamPath("", "blobHash", "BlobHash (Hex)").
		AddResponse(http.StatusUnauthorized, "Unauthorized (Wrong permissions, missing token)", authentication.ValidationError{}, nil).
		AddResponse(http.StatusOK, "All blob fields and their values", mocker.Get(BlobInfoResponse{}), nil).
		SetOperationId("blobsGetBlobInfo").
		SetSummary("Get all fields of a blob")
}

func (c *Controller) addErrorContractRoutes(api echoswagger.ApiGroup, mocker interfaces.Mocker) {
	api.GET("chains/:chainID/core/errors/:contractHname/message/:errorID", c.getErrorMessageFormat).
		AddParamPath("", "chainID", "ChainID (Bech32)").
		AddParamPath("", "contractHname", "Contract (Hname as Hex)").
		AddParamPath("", "errorID", "Error Id (uint16)").
		AddResponse(http.StatusUnauthorized, "Unauthorized (Wrong permissions, missing token)", authentication.ValidationError{}, nil).
		AddResponse(http.StatusOK, "The error message format", mocker.Get(ErrorMessageFormatResponse{}), nil).
		SetOperationId("errorsGetErrorMessageFormat").
		SetSummary("Get the error message format of a specific error id")
}

func (c *Controller) addGovernanceContractRoutes(api echoswagger.ApiGroup, mocker interfaces.Mocker) {
	api.GET("chains/:chainID/core/governance/chaininfo", c.getChainInfo).
		AddParamPath("", "chainID", "ChainID (Bech32)").
		AddResponse(http.StatusUnauthorized, "Unauthorized (Wrong permissions, missing token)", authentication.ValidationError{}, nil).
		AddResponse(http.StatusOK, "The chain info", mocker.Get(GovChainInfoResponse{}), nil).
		SetOperationId("governanceGetChainInfo").
		SetDescription("If you are using the common API functions, you most likely rather want to use '/chains/:chainID' to get information about a chain.").
		SetSummary("Get the chain info")
}

func (c *Controller) addBlockLogContractRoutes(api echoswagger.ApiGroup, mocker interfaces.Mocker) {
	api.GET("chains/:chainID/core/blocklog/controladdresses", c.getControlAddresses).
		AddParamPath("", "chainID", "ChainID (Bech32)").
		AddResponse(http.StatusUnauthorized, "Unauthorized (Wrong permissions, missing token)", authentication.ValidationError{}, nil).
		AddResponse(http.StatusOK, "The chain info", mocker.Get(models.ControlAddressesResponse{}), nil).
		SetOperationId("blocklogGetControlAddresses").
		SetSummary("Get the control addresses")

	api.GET("chains/:chainID/core/blocklog/blocks/:blockIndex", c.getBlockInfo).
		AddParamPath("", "chainID", "ChainID (Bech32)").
		AddParamPath(0, "blockIndex", "Block Index (uint32)").
		AddResponse(http.StatusUnauthorized, "Unauthorized (Wrong permissions, missing token)", authentication.ValidationError{}, nil).
		AddResponse(http.StatusOK, "The block info", mocker.Get(models.BlockInfoResponse{}), nil).
		SetOperationId("blocklogGetBlockInfo").
		SetSummary("Get the block info of a certain block index")

	api.GET("chains/:chainID/core/blocklog/blocks/latest", c.getBlockInfo).
		AddParamPath("", "chainID", "ChainID (Bech32)").
		AddResponse(http.StatusUnauthorized, "Unauthorized (Wrong permissions, missing token)", authentication.ValidationError{}, nil).
		AddResponse(http.StatusOK, "The block info", mocker.Get(models.BlockInfoResponse{}), nil).
		SetOperationId("blocklogGetLatestBlockInfo").
		SetSummary("Get the block info of the latest block")

	api.GET("chains/:chainID/core/blocklog/blocks/:blockIndex/requestids", c.getRequestIDsForBlock).
		AddParamPath("", "chainID", "ChainID (Bech32)").
		AddParamPath(0, "blockIndex", "Block Index (uint32)").
		AddResponse(http.StatusUnauthorized, "Unauthorized (Wrong permissions, missing token)", authentication.ValidationError{}, nil).
		AddResponse(http.StatusOK, "A list of request ids (ISCRequestID[])", mocker.Get(models.RequestIDsResponse{}), nil).
		SetOperationId("blocklogGetRequestIDsForBlock").
		SetSummary("Get the request ids for a certain block index")

	api.GET("chains/:chainID/core/blocklog/blocks/latest/requestids", c.getRequestIDsForBlock).
		AddParamPath("", "chainID", "ChainID (Bech32)").
		AddResponse(http.StatusUnauthorized, "Unauthorized (Wrong permissions, missing token)", authentication.ValidationError{}, nil).
		AddResponse(http.StatusOK, "A list of request ids (ISCRequestID[])", mocker.Get(models.RequestIDsResponse{}), nil).
		SetOperationId("blocklogGetRequestIDsForLatestBlock").
		SetSummary("Get the request ids for the latest block")

	api.GET("chains/:chainID/core/blocklog/blocks/:blockIndex/receipts", c.getRequestReceiptsForBlock).
		AddParamPath("", "chainID", "ChainID (Bech32)").
		AddParamPath(0, "blockIndex", "Block Index (uint32)").
		AddResponse(http.StatusUnauthorized, "Unauthorized (Wrong permissions, missing token)", authentication.ValidationError{}, nil).
		AddResponse(http.StatusOK, "The receipts", mocker.Get(models.BlockReceiptsResponse{}), nil).
		SetOperationId("blocklogGetRequestReceiptsOfBlock").
		SetSummary("Get all receipts of a certain block")

	api.GET("chains/:chainID/core/blocklog/blocks/latest/receipts", c.getRequestReceiptsForBlock).
		AddParamPath("", "chainID", "ChainID (Bech32)").
		AddResponse(http.StatusUnauthorized, "Unauthorized (Wrong permissions, missing token)", authentication.ValidationError{}, nil).
		AddResponse(http.StatusOK, "The receipts", mocker.Get(models.BlockReceiptsResponse{}), nil).
		SetOperationId("blocklogGetRequestReceiptsOfLatestBlock").
		SetSummary("Get all receipts of the latest block")

	api.GET("chains/:chainID/core/blocklog/requests/:requestID", c.getRequestReceipt).
		AddParamPath("", "chainID", "ChainID (Bech32)").
		AddParamPath("", "requestID", "Request ID (ISCRequestID)").
		AddResponse(http.StatusUnauthorized, "Unauthorized (Wrong permissions, missing token)", authentication.ValidationError{}, nil).
		AddResponse(http.StatusOK, "The receipt", mocker.Get(models.RequestReceiptResponse{}), nil).
		SetOperationId("blocklogGetRequestReceipt").
		SetSummary("Get the receipt of a certain request id")

	api.GET("chains/:chainID/core/blocklog/requests/:requestID/is_processed", c.getIsRequestProcessed).
		AddParamPath("", "chainID", "ChainID (Bech32)").
		AddParamPath("", "requestID", "Request ID (ISCRequestID)").
		AddResponse(http.StatusUnauthorized, "Unauthorized (Wrong permissions, missing token)", authentication.ValidationError{}, nil).
		AddResponse(http.StatusOK, "The processing result", mocker.Get(models.RequestProcessedResponse{}), nil).
		SetOperationId("blocklogGetRequestIsProcessed").
		SetSummary("Get the request processing status")

	api.GET("chains/:chainID/core/blocklog/events/block/:blockIndex", c.getBlockEvents).
		AddParamPath("", "chainID", "ChainID (Bech32)").
		AddParamPath(0, "blockIndex", "Block Index (uint32)").
		AddResponse(http.StatusUnauthorized, "Unauthorized (Wrong permissions, missing token)", authentication.ValidationError{}, nil).
		AddResponse(http.StatusOK, "The events", mocker.Get(models.EventsResponse{}), nil).
		SetOperationId("blocklogGetEventsOfBlock").
		SetSummary("Get events of a block")

	api.GET("chains/:chainID/core/blocklog/events/block/latest", c.getBlockEvents).
		AddParamPath("", "chainID", "ChainID (Bech32)").
		AddResponse(http.StatusUnauthorized, "Unauthorized (Wrong permissions, missing token)", authentication.ValidationError{}, nil).
		AddResponse(http.StatusOK, "The receipts", mocker.Get(models.EventsResponse{}), nil).
		SetOperationId("blocklogGetEventsOfLatestBlock").
		SetSummary("Get events of the latest block")

	api.GET("chains/:chainID/core/blocklog/events/request/:requestID", c.getRequestEvents).
		AddParamPath("", "chainID", "ChainID (Bech32)").
		AddParamPath("", "requestID", "Request ID (ISCRequestID)").
		AddResponse(http.StatusUnauthorized, "Unauthorized (Wrong permissions, missing token)", authentication.ValidationError{}, nil).
		AddResponse(http.StatusOK, "The events", mocker.Get(models.EventsResponse{}), nil).
		SetOperationId("blocklogGetEventsOfRequest").
		SetSummary("Get events of a request")

	api.GET("chains/:chainID/core/blocklog/events/contract/:contractHname", c.getContractEvents).
		AddParamPath("", "chainID", "ChainID (Bech32)").
		AddParamPath("", "contractHname", "Contract (Hname)").
		AddResponse(http.StatusUnauthorized, "Unauthorized (Wrong permissions, missing token)", authentication.ValidationError{}, nil).
		AddResponse(http.StatusOK, "The events", mocker.Get(models.EventsResponse{}), nil).
		SetOperationId("blocklogGetEventsOfContract").
		SetSummary("Get events of a contract")
}

func (c *Controller) RegisterPublic(publicAPI echoswagger.ApiGroup, mocker interfaces.Mocker) {
	c.addAccountContractRoutes(publicAPI, mocker)
	c.addBlobContractRoutes(publicAPI, mocker)
	c.addBlockLogContractRoutes(publicAPI, mocker)
	c.addErrorContractRoutes(publicAPI, mocker)
	c.addGovernanceContractRoutes(publicAPI, mocker)
}

func (c *Controller) RegisterAdmin(adminAPI echoswagger.ApiGroup, mocker interfaces.Mocker) {
}
