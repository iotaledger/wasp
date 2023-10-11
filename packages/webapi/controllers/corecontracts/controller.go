package corecontracts

import (
	"errors"
	"net/http"

	"github.com/pangpanglabs/echoswagger/v2"

	"github.com/iotaledger/wasp/packages/authentication"
	"github.com/iotaledger/wasp/packages/isc"
	"github.com/iotaledger/wasp/packages/webapi/apierrors"
	"github.com/iotaledger/wasp/packages/webapi/interfaces"
	"github.com/iotaledger/wasp/packages/webapi/models"
	"github.com/iotaledger/wasp/packages/webapi/params"
)

type Controller struct {
	chainService interfaces.ChainService
}

func NewCoreContractsController(chainService interfaces.ChainService) interfaces.APIController {
	return &Controller{chainService}
}

func (c *Controller) Name() string {
	return "corecontracts"
}

func (c *Controller) handleViewCallError(err error, chainID isc.ChainID) error {
	if errors.Is(err, interfaces.ErrChainNotFound) {
		return apierrors.ChainNotFoundError(chainID.String())
	}
	return apierrors.ContractExecutionError(err)
}

func (c *Controller) addAccountContractRoutes(api echoswagger.ApiGroup, mocker interfaces.Mocker) {
	api.GET("chains/:chainID/core/accounts", c.getAccounts).
		AddParamPath("", params.ParamChainID, params.DescriptionChainID).
		AddParamQuery("", params.ParamBlockIndexOrTrieRoot, params.DescriptionBlockIndexOrTrieRoot, false).
		AddResponse(http.StatusUnauthorized, "Unauthorized (Wrong permissions, missing token)", authentication.ValidationError{}, nil).
		AddResponse(http.StatusOK, "A list of all accounts", mocker.Get(models.AccountListResponse{}), nil).
		SetOperationId("accountsGetAccounts").
		SetSummary("Get a list of all accounts")

	api.GET("chains/:chainID/core/accounts/account/:agentID/balance", c.getAccountBalance).
		AddParamPath("", params.ParamChainID, params.DescriptionChainID).
		AddParamPath("", params.ParamAgentID, params.DescriptionAgentID).
		AddParamQuery("", params.ParamBlockIndexOrTrieRoot, params.DescriptionBlockIndexOrTrieRoot, false).
		AddResponse(http.StatusUnauthorized, "Unauthorized (Wrong permissions, missing token)", authentication.ValidationError{}, nil).
		AddResponse(http.StatusOK, "All assets belonging to an account", mocker.Get(models.AssetsResponse{}), nil).
		SetOperationId("accountsGetAccountBalance").
		SetSummary("Get all assets belonging to an account")

	api.GET("chains/:chainID/core/accounts/account/:agentID/nfts", c.getAccountNFTs).
		AddParamPath("", params.ParamChainID, params.DescriptionChainID).
		AddParamPath("", params.ParamAgentID, params.DescriptionAgentID).
		AddParamQuery("", params.ParamBlockIndexOrTrieRoot, params.DescriptionBlockIndexOrTrieRoot, false).
		AddResponse(http.StatusUnauthorized, "Unauthorized (Wrong permissions, missing token)", authentication.ValidationError{}, nil).
		AddResponse(http.StatusOK, "All NFT ids belonging to an account", mocker.Get(models.AccountNFTsResponse{}), nil).
		SetOperationId("accountsGetAccountNFTIDs").
		SetSummary("Get all NFT ids belonging to an account")

	api.GET("chains/:chainID/core/accounts/account/:agentID/foundries", c.getAccountFoundries).
		AddParamPath("", "chainID", "ChainID (Bech32)").
		AddParamPath("", "agentID", "AgentID (Bech32 for WasmVM | Hex for EVM)").
		AddParamQuery("", params.ParamBlockIndexOrTrieRoot, params.DescriptionBlockIndexOrTrieRoot, false).
		AddResponse(http.StatusUnauthorized, "Unauthorized (Wrong permissions, missing token)", authentication.ValidationError{}, nil).
		AddResponse(http.StatusOK, "All foundries owned by an account", mocker.Get(models.AccountFoundriesResponse{}), nil).
		SetOperationId("accountsGetAccountFoundries").
		SetSummary("Get all foundries owned by an account")

	api.GET("chains/:chainID/core/accounts/account/:agentID/nonce", c.getAccountNonce).
		AddParamPath("", params.ParamChainID, params.DescriptionChainID).
		AddParamPath("", params.ParamAgentID, params.DescriptionAgentID).
		AddParamQuery("", params.ParamBlockIndexOrTrieRoot, params.DescriptionBlockIndexOrTrieRoot, false).
		AddResponse(http.StatusUnauthorized, "Unauthorized (Wrong permissions, missing token)", authentication.ValidationError{}, nil).
		AddResponse(http.StatusOK, "The current nonce of an account", mocker.Get(models.AccountNonceResponse{}), nil).
		SetOperationId("accountsGetAccountNonce").
		SetSummary("Get the current nonce of an account")

	api.GET("chains/:chainID/core/accounts/nftdata/:nftID", c.getNFTData).
		AddParamPath("", params.ParamChainID, params.DescriptionChainID).
		AddParamPath("", params.ParamNFTID, params.DescriptionNFTID).
		AddParamQuery("", params.ParamBlockIndexOrTrieRoot, params.DescriptionBlockIndexOrTrieRoot, false).
		AddResponse(http.StatusUnauthorized, "Unauthorized (Wrong permissions, missing token)", authentication.ValidationError{}, nil).
		AddResponse(http.StatusOK, "The NFT data", mocker.Get(isc.NFTJSON{}), nil).
		SetOperationId("accountsGetNFTData").
		SetSummary("Get the NFT data by an ID")

	api.GET("chains/:chainID/core/accounts/token_registry", c.getNativeTokenIDRegistry).
		AddParamPath("", params.ParamChainID, params.DescriptionChainID).
		AddParamQuery("", params.ParamBlockIndexOrTrieRoot, params.DescriptionBlockIndexOrTrieRoot, false).
		AddResponse(http.StatusUnauthorized, "Unauthorized (Wrong permissions, missing token)", authentication.ValidationError{}, nil).
		AddResponse(http.StatusOK, "A list of all registries", mocker.Get(models.NativeTokenIDRegistryResponse{}), nil).
		SetOperationId("accountsGetNativeTokenIDRegistry").
		SetSummary("Get a list of all registries")

	//nolint:unused
	type foundryOutputParams struct {
		chainID      string `swagger:"required,desc(ChainID (Bech32))"`
		serialNumber uint32 `swagger:"required,min(1),desc(Serial Number (uint32))"`
	}

	api.GET("chains/:chainID/core/accounts/foundry_output/:serialNumber", c.getFoundryOutput).
		AddParamPathNested(foundryOutputParams{}).
		AddParamQuery("", params.ParamBlockIndexOrTrieRoot, params.DescriptionBlockIndexOrTrieRoot, false).
		AddResponse(http.StatusUnauthorized, "Unauthorized (Wrong permissions, missing token)", authentication.ValidationError{}, nil).
		AddResponse(http.StatusOK, "The foundry output", mocker.Get(models.FoundryOutputResponse{}), nil).
		SetOperationId("accountsGetFoundryOutput").
		SetSummary("Get the foundry output")

	api.GET("chains/:chainID/core/accounts/total_assets", c.getTotalAssets).
		AddParamPath("", params.ParamChainID, params.DescriptionChainID).
		AddParamQuery("", params.ParamBlockIndexOrTrieRoot, params.DescriptionBlockIndexOrTrieRoot, false).
		AddResponse(http.StatusUnauthorized, "Unauthorized (Wrong permissions, missing token)", authentication.ValidationError{}, nil).
		AddResponse(http.StatusOK, "All stored assets", mocker.Get(models.AssetsResponse{}), nil).
		SetOperationId("accountsGetTotalAssets").
		SetSummary("Get all stored assets")
}

func (c *Controller) addBlobContractRoutes(api echoswagger.ApiGroup, mocker interfaces.Mocker) {
	api.GET("chains/:chainID/core/blobs", c.listBlobs).
		AddParamPath("", params.ParamChainID, params.DescriptionChainID).
		AddParamQuery("", params.ParamBlockIndexOrTrieRoot, params.DescriptionBlockIndexOrTrieRoot, false).
		AddResponse(http.StatusUnauthorized, "Unauthorized (Wrong permissions, missing token)", authentication.ValidationError{}, nil).
		AddResponse(http.StatusOK, "All stored blobs", mocker.Get(BlobListResponse{}), nil).
		SetOperationId("blobsGetAllBlobs").
		SetSummary("Get all stored blobs")

	api.GET("chains/:chainID/core/blobs/:blobHash/data/:fieldKey", c.getBlobValue).
		AddParamPath("", params.ParamChainID, params.DescriptionChainID).
		AddParamPath("", params.ParamBlobHash, params.DescriptionBlobHash).
		AddParamPath("", params.ParamFieldKey, params.DescriptionFieldKey).
		AddParamQuery("", params.ParamBlockIndexOrTrieRoot, params.DescriptionBlockIndexOrTrieRoot, false).
		AddResponse(http.StatusUnauthorized, "Unauthorized (Wrong permissions, missing token)", authentication.ValidationError{}, nil).
		AddResponse(http.StatusOK, "The value of the supplied field (key)", mocker.Get(BlobValueResponse{}), nil).
		SetOperationId("blobsGetBlobValue").
		SetSummary("Get the value of the supplied field (key)")

	api.GET("chains/:chainID/core/blobs/:blobHash", c.getBlobInfo).
		AddParamPath("", params.ParamChainID, params.DescriptionChainID).
		AddParamPath("", params.ParamBlobHash, params.DescriptionBlobHash).
		AddParamQuery("", params.ParamBlockIndexOrTrieRoot, params.DescriptionBlockIndexOrTrieRoot, false).
		AddResponse(http.StatusUnauthorized, "Unauthorized (Wrong permissions, missing token)", authentication.ValidationError{}, nil).
		AddResponse(http.StatusOK, "All blob fields and their values", mocker.Get(BlobInfoResponse{}), nil).
		SetOperationId("blobsGetBlobInfo").
		SetSummary("Get all fields of a blob")
}

func (c *Controller) addErrorContractRoutes(api echoswagger.ApiGroup, mocker interfaces.Mocker) {
	//nolint:unused
	type errorMessageFormat struct {
		chainID       string `swagger:"required,desc(ChainID (Bech32))"`
		contractHname string `swagger:"required,desc(Contract (Hname as Hex))"`
		errorID       uint16 `swagger:"required,min(1),desc(Error Id (uint16))"`
	}

	api.GET("chains/:chainID/core/errors/:contractHname/message/:errorID", c.getErrorMessageFormat).
		AddParamPathNested(errorMessageFormat{}).
		AddParamQuery("", params.ParamBlockIndexOrTrieRoot, params.DescriptionBlockIndexOrTrieRoot, false).
		AddResponse(http.StatusUnauthorized, "Unauthorized (Wrong permissions, missing token)", authentication.ValidationError{}, nil).
		AddResponse(http.StatusOK, "The error message format", mocker.Get(ErrorMessageFormatResponse{}), nil).
		SetOperationId("errorsGetErrorMessageFormat").
		SetSummary("Get the error message format of a specific error id")
}

func (c *Controller) addGovernanceContractRoutes(api echoswagger.ApiGroup, mocker interfaces.Mocker) {
	api.GET("chains/:chainID/core/governance/chaininfo", c.getChainInfo).
		AddParamPath("", params.ParamChainID, params.DescriptionChainID).
		AddParamQuery("", params.ParamBlockIndexOrTrieRoot, params.DescriptionBlockIndexOrTrieRoot, false).
		AddResponse(http.StatusUnauthorized, "Unauthorized (Wrong permissions, missing token)", authentication.ValidationError{}, nil).
		AddResponse(http.StatusOK, "The chain info", mocker.Get(models.GovChainInfoResponse{}), nil).
		SetOperationId("governanceGetChainInfo").
		SetDescription("If you are using the common API functions, you most likely rather want to use '/v1/chains/:chainID' to get information about a chain.").
		SetSummary("Get the chain info")

	api.GET("chains/:chainID/core/governance/chainowner", c.getChainOwner).
		AddParamPath("", params.ParamChainID, params.DescriptionChainID).
		AddParamQuery("", params.ParamBlockIndexOrTrieRoot, params.DescriptionBlockIndexOrTrieRoot, false).
		AddResponse(http.StatusUnauthorized, "Unauthorized (Wrong permissions, missing token)", authentication.ValidationError{}, nil).
		AddResponse(http.StatusOK, "The chain owner", mocker.Get(models.GovChainOwnerResponse{}), nil).
		SetOperationId("governanceGetChainOwner").
		SetDescription("Returns the chain owner").
		SetSummary("Get the chain owner")

	api.GET("chains/:chainID/core/governance/allowedstatecontrollers", c.getAllowedStateControllerAddresses).
		AddParamPath("", params.ParamChainID, params.DescriptionChainID).
		AddParamQuery("", params.ParamBlockIndexOrTrieRoot, params.DescriptionBlockIndexOrTrieRoot, false).
		AddResponse(http.StatusUnauthorized, "Unauthorized (Wrong permissions, missing token)", authentication.ValidationError{}, nil).
		AddResponse(http.StatusOK, "The state controller addresses", mocker.Get(models.GovAllowedStateControllerAddressesResponse{}), nil).
		SetOperationId("governanceGetAllowedStateControllerAddresses").
		SetDescription("Returns the allowed state controller addresses").
		SetSummary("Get the allowed state controller addresses")
}

//nolint:funlen
func (c *Controller) addBlockLogContractRoutes(api echoswagger.ApiGroup, mocker interfaces.Mocker) {
	api.GET("chains/:chainID/core/blocklog/controladdresses", c.getControlAddresses).
		AddParamPath("", params.ParamChainID, params.DescriptionChainID).
		AddParamQuery("", params.ParamBlockIndexOrTrieRoot, params.DescriptionBlockIndexOrTrieRoot, false).
		AddResponse(http.StatusUnauthorized, "Unauthorized (Wrong permissions, missing token)", authentication.ValidationError{}, nil).
		AddResponse(http.StatusOK, "The chain info", mocker.Get(models.ControlAddressesResponse{}), nil).
		SetOperationId("blocklogGetControlAddresses").
		SetSummary("Get the control addresses")

	//nolint:unused
	type blocks struct {
		chainID    string `swagger:"required,desc(ChainID (Bech32))"`
		blockIndex uint32 `swagger:"required,min(1),desc(BlockIndex (uint32))"`
	}

	api.GET("chains/:chainID/core/blocklog/blocks/:blockIndex", c.getBlockInfo).
		AddParamPathNested(blocks{}).
		AddParamQuery("", params.ParamBlockIndexOrTrieRoot, params.DescriptionBlockIndexOrTrieRoot, false).
		AddResponse(http.StatusUnauthorized, "Unauthorized (Wrong permissions, missing token)", authentication.ValidationError{}, nil).
		AddResponse(http.StatusOK, "The block info", mocker.Get(models.BlockInfoResponse{}), nil).
		SetOperationId("blocklogGetBlockInfo").
		SetSummary("Get the block info of a certain block index")

	api.GET("chains/:chainID/core/blocklog/blocks/latest", c.getBlockInfo).
		AddParamPath("", params.ParamChainID, params.DescriptionChainID).
		AddParamQuery("", params.ParamBlockIndexOrTrieRoot, params.DescriptionBlockIndexOrTrieRoot, false).
		AddResponse(http.StatusUnauthorized, "Unauthorized (Wrong permissions, missing token)", authentication.ValidationError{}, nil).
		AddResponse(http.StatusOK, "The block info", mocker.Get(models.BlockInfoResponse{}), nil).
		SetOperationId("blocklogGetLatestBlockInfo").
		SetSummary("Get the block info of the latest block")

	api.GET("chains/:chainID/core/blocklog/blocks/:blockIndex/requestids", c.getRequestIDsForBlock).
		AddParamPathNested(blocks{}).
		AddParamQuery("", params.ParamBlockIndexOrTrieRoot, params.DescriptionBlockIndexOrTrieRoot, false).
		AddResponse(http.StatusUnauthorized, "Unauthorized (Wrong permissions, missing token)", authentication.ValidationError{}, nil).
		AddResponse(http.StatusOK, "A list of request ids (ISCRequestID[])", mocker.Get(models.RequestIDsResponse{}), nil).
		SetOperationId("blocklogGetRequestIDsForBlock").
		SetSummary("Get the request ids for a certain block index")

	api.GET("chains/:chainID/core/blocklog/blocks/latest/requestids", c.getRequestIDsForBlock).
		AddParamPath("", params.ParamChainID, params.DescriptionChainID).
		AddParamQuery("", params.ParamBlockIndexOrTrieRoot, params.DescriptionBlockIndexOrTrieRoot, false).
		AddResponse(http.StatusUnauthorized, "Unauthorized (Wrong permissions, missing token)", authentication.ValidationError{}, nil).
		AddResponse(http.StatusOK, "A list of request ids (ISCRequestID[])", mocker.Get(models.RequestIDsResponse{}), nil).
		SetOperationId("blocklogGetRequestIDsForLatestBlock").
		SetSummary("Get the request ids for the latest block")

	api.GET("chains/:chainID/core/blocklog/blocks/:blockIndex/receipts", c.getRequestReceiptsForBlock).
		AddParamPathNested(blocks{}).
		AddParamQuery("", params.ParamBlockIndexOrTrieRoot, params.DescriptionBlockIndexOrTrieRoot, false).
		AddResponse(http.StatusUnauthorized, "Unauthorized (Wrong permissions, missing token)", authentication.ValidationError{}, nil).
		AddResponse(http.StatusOK, "The receipts", mocker.Get([]models.ReceiptResponse{}), nil).
		SetOperationId("blocklogGetRequestReceiptsOfBlock").
		SetSummary("Get all receipts of a certain block")

	api.GET("chains/:chainID/core/blocklog/blocks/latest/receipts", c.getRequestReceiptsForBlock).
		AddParamPath("", params.ParamChainID, params.DescriptionChainID).
		AddParamQuery("", params.ParamBlockIndexOrTrieRoot, params.DescriptionBlockIndexOrTrieRoot, false).
		AddResponse(http.StatusUnauthorized, "Unauthorized (Wrong permissions, missing token)", authentication.ValidationError{}, nil).
		AddResponse(http.StatusOK, "The receipts", mocker.Get([]models.ReceiptResponse{}), nil).
		SetOperationId("blocklogGetRequestReceiptsOfLatestBlock").
		SetSummary("Get all receipts of the latest block")

	api.GET("chains/:chainID/core/blocklog/requests/:requestID", c.getRequestReceipt).
		AddParamPath("", params.ParamChainID, params.DescriptionChainID).
		AddParamPath("", params.ParamRequestID, params.DescriptionRequestID).
		AddParamQuery("", params.ParamBlockIndexOrTrieRoot, params.DescriptionBlockIndexOrTrieRoot, false).
		AddResponse(http.StatusUnauthorized, "Unauthorized (Wrong permissions, missing token)", authentication.ValidationError{}, nil).
		AddResponse(http.StatusOK, "The receipt", mocker.Get(models.ReceiptResponse{}), nil).
		SetOperationId("blocklogGetRequestReceipt").
		SetSummary("Get the receipt of a certain request id")

	api.GET("chains/:chainID/core/blocklog/requests/:requestID/is_processed", c.getIsRequestProcessed).
		AddParamPath("", params.ParamChainID, params.DescriptionChainID).
		AddParamPath("", params.ParamRequestID, params.DescriptionRequestID).
		AddParamQuery("", params.ParamBlockIndexOrTrieRoot, params.DescriptionBlockIndexOrTrieRoot, false).
		AddResponse(http.StatusUnauthorized, "Unauthorized (Wrong permissions, missing token)", authentication.ValidationError{}, nil).
		AddResponse(http.StatusOK, "The processing result", mocker.Get(models.RequestProcessedResponse{}), nil).
		SetOperationId("blocklogGetRequestIsProcessed").
		SetSummary("Get the request processing status")

	api.GET("chains/:chainID/core/blocklog/events/block/:blockIndex", c.getBlockEvents).
		AddParamPathNested(blocks{}).
		AddParamQuery("", params.ParamBlockIndexOrTrieRoot, params.DescriptionBlockIndexOrTrieRoot, false).
		AddResponse(http.StatusUnauthorized, "Unauthorized (Wrong permissions, missing token)", authentication.ValidationError{}, nil).
		AddResponse(http.StatusOK, "The events", mocker.Get(models.EventsResponse{}), nil).
		SetOperationId("blocklogGetEventsOfBlock").
		SetSummary("Get events of a block")

	api.GET("chains/:chainID/core/blocklog/events/block/latest", c.getBlockEvents).
		AddParamPath("", params.ParamChainID, params.DescriptionChainID).
		AddParamQuery("", params.ParamBlockIndexOrTrieRoot, params.DescriptionBlockIndexOrTrieRoot, false).
		AddResponse(http.StatusUnauthorized, "Unauthorized (Wrong permissions, missing token)", authentication.ValidationError{}, nil).
		AddResponse(http.StatusOK, "The receipts", mocker.Get(models.EventsResponse{}), nil).
		SetOperationId("blocklogGetEventsOfLatestBlock").
		SetSummary("Get events of the latest block")

	api.GET("chains/:chainID/core/blocklog/events/request/:requestID", c.getRequestEvents).
		AddParamPath("", params.ParamChainID, params.DescriptionChainID).
		AddParamPath("", params.ParamRequestID, params.DescriptionRequestID).
		AddParamQuery("", params.ParamBlockIndexOrTrieRoot, params.DescriptionBlockIndexOrTrieRoot, false).
		AddResponse(http.StatusUnauthorized, "Unauthorized (Wrong permissions, missing token)", authentication.ValidationError{}, nil).
		AddResponse(http.StatusOK, "The events", mocker.Get(models.EventsResponse{}), nil).
		SetOperationId("blocklogGetEventsOfRequest").
		SetSummary("Get events of a request")

	api.GET("chains/:chainID/core/blocklog/events/contract/:contractHname", c.getContractEvents).
		AddParamPath("", params.ParamChainID, params.DescriptionChainID).
		AddParamPath("", params.ParamContractHName, params.DescriptionContractHName).
		AddParamQuery("", params.ParamBlockIndexOrTrieRoot, params.DescriptionBlockIndexOrTrieRoot, false).
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
