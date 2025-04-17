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

func (c *Controller) handleViewCallError(err error) error {
	if errors.Is(err, interfaces.ErrChainNotFound) {
		return apierrors.ChainNotFoundError()
	}
	return apierrors.ContractExecutionError(err)
}

func (c *Controller) addAccountContractRoutes(api echoswagger.ApiGroup, mocker interfaces.Mocker) {
	api.GET("chain/core/accounts/account/:agentID/balance", c.getAccountBalance).
		AddParamPath("", params.ParamAgentID, params.DescriptionAgentID).
		AddParamQuery("", params.ParamBlockIndexOrTrieRoot, params.DescriptionBlockIndexOrTrieRoot, false).
		AddResponse(http.StatusUnauthorized, "Unauthorized (Wrong permissions, missing token)", authentication.ValidationError{}, nil).
		AddResponse(http.StatusOK, "All assets belonging to an account", mocker.Get(models.AssetsResponse{}), nil).
		SetOperationId("accountsGetAccountBalance").
		SetSummary("Get all assets belonging to an account")

	api.GET("chain/core/accounts/account/:agentID/objects", c.getAccountObjects).
		AddParamPath("", params.ParamAgentID, params.DescriptionAgentID).
		AddParamQuery("", params.ParamBlockIndexOrTrieRoot, params.DescriptionBlockIndexOrTrieRoot, false).
		AddResponse(http.StatusUnauthorized, "Unauthorized (Wrong permissions, missing token)", authentication.ValidationError{}, nil).
		AddResponse(http.StatusOK, "All object ids belonging to an account", mocker.Get(models.AccountObjectsResponse{}), nil).
		SetOperationId("accountsGetAccountObjectIDs").
		SetSummary("Get all object ids belonging to an account")

	api.GET("chain/core/accounts/account/:agentID/nonce", c.getAccountNonce).
		AddParamPath("", params.ParamAgentID, params.DescriptionAgentID).
		AddParamQuery("", params.ParamBlockIndexOrTrieRoot, params.DescriptionBlockIndexOrTrieRoot, false).
		AddResponse(http.StatusUnauthorized, "Unauthorized (Wrong permissions, missing token)", authentication.ValidationError{}, nil).
		AddResponse(http.StatusOK, "The current nonce of an account", mocker.Get(models.AccountNonceResponse{}), nil).
		SetOperationId("accountsGetAccountNonce").
		SetSummary("Get the current nonce of an account")

	api.GET("chain/core/accounts/objectdata/:objectID", c.getObjectData).
		AddParamPath("", params.ParamObjectID, params.DescriptionObjectID).
		AddParamQuery("", params.ParamBlockIndexOrTrieRoot, params.DescriptionBlockIndexOrTrieRoot, false).
		AddResponse(http.StatusUnauthorized, "Unauthorized (Wrong permissions, missing token)", authentication.ValidationError{}, nil).
		AddResponse(http.StatusOK, "The object data", mocker.Get(isc.IotaObject{}), nil).
		SetOperationId("accountsGetObjectData").
		SetSummary("Get the object data by an ID")

	api.GET("chain/core/accounts/total_assets", c.getTotalAssets).
		AddParamQuery("", params.ParamBlockIndexOrTrieRoot, params.DescriptionBlockIndexOrTrieRoot, false).
		AddResponse(http.StatusUnauthorized, "Unauthorized (Wrong permissions, missing token)", authentication.ValidationError{}, nil).
		AddResponse(http.StatusOK, "All stored assets", mocker.Get(models.AssetsResponse{}), nil).
		SetOperationId("accountsGetTotalAssets").
		SetSummary("Get all stored assets")
}

func (c *Controller) addErrorContractRoutes(api echoswagger.ApiGroup, mocker interfaces.Mocker) {
	//nolint:unused
	type errorMessageFormat struct {
		chainID       string `swagger:"required,desc(ChainID (Hex Address))"`
		contractHname string `swagger:"required,desc(Contract (Hname as Hex))"`
		errorID       uint16 `swagger:"required,min(1),desc(Error Id (uint16))"`
	}

	api.GET("chain/core/errors/:contractHname/message/:errorID", c.getErrorMessageFormat).
		AddParamPathNested(errorMessageFormat{}).
		AddParamQuery("", params.ParamBlockIndexOrTrieRoot, params.DescriptionBlockIndexOrTrieRoot, false).
		AddResponse(http.StatusUnauthorized, "Unauthorized (Wrong permissions, missing token)", authentication.ValidationError{}, nil).
		AddResponse(http.StatusOK, "The error message format", mocker.Get(ErrorMessageFormatResponse{}), nil).
		SetOperationId("errorsGetErrorMessageFormat").
		SetSummary("Get the error message format of a specific error id")
}

func (c *Controller) addGovernanceContractRoutes(api echoswagger.ApiGroup, mocker interfaces.Mocker) {
	api.GET("chain/core/governance/chaininfo", c.getChainInfo).
		AddParamQuery("", params.ParamBlockIndexOrTrieRoot, params.DescriptionBlockIndexOrTrieRoot, false).
		AddResponse(http.StatusUnauthorized, "Unauthorized (Wrong permissions, missing token)", authentication.ValidationError{}, nil).
		AddResponse(http.StatusOK, "The chain info", mocker.Get(models.GovChainInfoResponse{}), nil).
		SetOperationId("governanceGetChainInfo").
		SetDescription("If you are using the common API functions, you most likely rather want to use '/v1/chains/:chainID' to get information about a chain.").
		SetSummary("Get the chain info")

	api.GET("chain/core/governance/chainadmin", c.getChainAdmin).
		AddParamQuery("", params.ParamBlockIndexOrTrieRoot, params.DescriptionBlockIndexOrTrieRoot, false).
		AddResponse(http.StatusUnauthorized, "Unauthorized (Wrong permissions, missing token)", authentication.ValidationError{}, nil).
		AddResponse(http.StatusOK, "The chain admin", mocker.Get(models.GovChainAdminResponse{}), nil).
		SetOperationId("governanceGetChainAdmin").
		SetDescription("Returns the chain admin").
		SetSummary("Get the chain admin")
}

//nolint:funlen
func (c *Controller) addBlockLogContractRoutes(api echoswagger.ApiGroup, mocker interfaces.Mocker) {
	api.GET("chain/core/blocklog/controladdresses", c.getControlAddresses).
		AddParamQuery("", params.ParamBlockIndexOrTrieRoot, params.DescriptionBlockIndexOrTrieRoot, false).
		AddResponse(http.StatusUnauthorized, "Unauthorized (Wrong permissions, missing token)", authentication.ValidationError{}, nil).
		AddResponse(http.StatusOK, "The chain info", mocker.Get(models.ControlAddressesResponse{}), nil).
		SetOperationId("blocklogGetControlAddresses").
		SetSummary("Get the control addresses")

	//nolint:unused
	type blocks struct {
		blockIndex uint32 `swagger:"required,min(1),desc(BlockIndex (uint32))"`
	}

	api.GET("chain/core/blocklog/blocks/:blockIndex", c.getBlockInfo).
		AddParamPathNested(blocks{}).
		AddParamQuery("", params.ParamBlockIndexOrTrieRoot, params.DescriptionBlockIndexOrTrieRoot, false).
		AddResponse(http.StatusUnauthorized, "Unauthorized (Wrong permissions, missing token)", authentication.ValidationError{}, nil).
		AddResponse(http.StatusOK, "The block info", mocker.Get(models.BlockInfoResponse{}), nil).
		SetOperationId("blocklogGetBlockInfo").
		SetSummary("Get the block info of a certain block index")

	api.GET("chain/core/blocklog/blocks/latest", c.getBlockInfo).
		AddParamQuery("", params.ParamBlockIndexOrTrieRoot, params.DescriptionBlockIndexOrTrieRoot, false).
		AddResponse(http.StatusUnauthorized, "Unauthorized (Wrong permissions, missing token)", authentication.ValidationError{}, nil).
		AddResponse(http.StatusOK, "The block info", mocker.Get(models.BlockInfoResponse{}), nil).
		SetOperationId("blocklogGetLatestBlockInfo").
		SetSummary("Get the block info of the latest block")

	api.GET("chain/core/blocklog/blocks/:blockIndex/requestids", c.getRequestIDsForBlock).
		AddParamPathNested(blocks{}).
		AddParamQuery("", params.ParamBlockIndexOrTrieRoot, params.DescriptionBlockIndexOrTrieRoot, false).
		AddResponse(http.StatusUnauthorized, "Unauthorized (Wrong permissions, missing token)", authentication.ValidationError{}, nil).
		AddResponse(http.StatusOK, "A list of request ids (ISCRequestID[])", mocker.Get(models.RequestIDsResponse{}), nil).
		SetOperationId("blocklogGetRequestIDsForBlock").
		SetSummary("Get the request ids for a certain block index")

	api.GET("chain/core/blocklog/blocks/latest/requestids", c.getRequestIDsForBlock).
		AddParamQuery("", params.ParamBlockIndexOrTrieRoot, params.DescriptionBlockIndexOrTrieRoot, false).
		AddResponse(http.StatusUnauthorized, "Unauthorized (Wrong permissions, missing token)", authentication.ValidationError{}, nil).
		AddResponse(http.StatusOK, "A list of request ids (ISCRequestID[])", mocker.Get(models.RequestIDsResponse{}), nil).
		SetOperationId("blocklogGetRequestIDsForLatestBlock").
		SetSummary("Get the request ids for the latest block")

	api.GET("chain/core/blocklog/blocks/:blockIndex/receipts", c.getRequestReceiptsForBlock).
		AddParamPathNested(blocks{}).
		AddParamQuery("", params.ParamBlockIndexOrTrieRoot, params.DescriptionBlockIndexOrTrieRoot, false).
		AddResponse(http.StatusUnauthorized, "Unauthorized (Wrong permissions, missing token)", authentication.ValidationError{}, nil).
		AddResponse(http.StatusOK, "The receipts", mocker.Get([]models.ReceiptResponse{}), nil).
		SetOperationId("blocklogGetRequestReceiptsOfBlock").
		SetSummary("Get all receipts of a certain block")

	api.GET("chain/core/blocklog/blocks/latest/receipts", c.getRequestReceiptsForBlock).
		AddParamQuery("", params.ParamBlockIndexOrTrieRoot, params.DescriptionBlockIndexOrTrieRoot, false).
		AddResponse(http.StatusUnauthorized, "Unauthorized (Wrong permissions, missing token)", authentication.ValidationError{}, nil).
		AddResponse(http.StatusOK, "The receipts", mocker.Get([]models.ReceiptResponse{}), nil).
		SetOperationId("blocklogGetRequestReceiptsOfLatestBlock").
		SetSummary("Get all receipts of the latest block")

	api.GET("chain/core/blocklog/requests/:requestID", c.getRequestReceipt).
		AddParamPath("", params.ParamRequestID, params.DescriptionRequestID).
		AddParamQuery("", params.ParamBlockIndexOrTrieRoot, params.DescriptionBlockIndexOrTrieRoot, false).
		AddResponse(http.StatusUnauthorized, "Unauthorized (Wrong permissions, missing token)", authentication.ValidationError{}, nil).
		AddResponse(http.StatusOK, "The receipt", mocker.Get(models.ReceiptResponse{}), nil).
		SetOperationId("blocklogGetRequestReceipt").
		SetSummary("Get the receipt of a certain request id")

	api.GET("chain/core/blocklog/requests/:requestID/is_processed", c.getIsRequestProcessed).
		AddParamPath("", params.ParamRequestID, params.DescriptionRequestID).
		AddParamQuery("", params.ParamBlockIndexOrTrieRoot, params.DescriptionBlockIndexOrTrieRoot, false).
		AddResponse(http.StatusUnauthorized, "Unauthorized (Wrong permissions, missing token)", authentication.ValidationError{}, nil).
		AddResponse(http.StatusOK, "The processing result", mocker.Get(models.RequestProcessedResponse{}), nil).
		SetOperationId("blocklogGetRequestIsProcessed").
		SetSummary("Get the request processing status")

	api.GET("chain/core/blocklog/events/block/:blockIndex", c.getBlockEvents).
		AddParamPathNested(blocks{}).
		AddParamQuery("", params.ParamBlockIndexOrTrieRoot, params.DescriptionBlockIndexOrTrieRoot, false).
		AddResponse(http.StatusUnauthorized, "Unauthorized (Wrong permissions, missing token)", authentication.ValidationError{}, nil).
		AddResponse(http.StatusOK, "The events", mocker.Get(models.EventsResponse{}), nil).
		SetOperationId("blocklogGetEventsOfBlock").
		SetSummary("Get events of a block")

	api.GET("chain/core/blocklog/events/block/latest", c.getBlockEvents).
		AddParamQuery("", params.ParamBlockIndexOrTrieRoot, params.DescriptionBlockIndexOrTrieRoot, false).
		AddResponse(http.StatusUnauthorized, "Unauthorized (Wrong permissions, missing token)", authentication.ValidationError{}, nil).
		AddResponse(http.StatusOK, "The receipts", mocker.Get(models.EventsResponse{}), nil).
		SetOperationId("blocklogGetEventsOfLatestBlock").
		SetSummary("Get events of the latest block")

	api.GET("chain/core/blocklog/events/request/:requestID", c.getRequestEvents).
		AddParamPath("", params.ParamRequestID, params.DescriptionRequestID).
		AddParamQuery("", params.ParamBlockIndexOrTrieRoot, params.DescriptionBlockIndexOrTrieRoot, false).
		AddResponse(http.StatusUnauthorized, "Unauthorized (Wrong permissions, missing token)", authentication.ValidationError{}, nil).
		AddResponse(http.StatusOK, "The events", mocker.Get(models.EventsResponse{}), nil).
		SetOperationId("blocklogGetEventsOfRequest").
		SetSummary("Get events of a request")
}

func (c *Controller) RegisterPublic(publicAPI echoswagger.ApiGroup, mocker interfaces.Mocker) {
	c.addAccountContractRoutes(publicAPI, mocker)
	c.addBlockLogContractRoutes(publicAPI, mocker)
	c.addErrorContractRoutes(publicAPI, mocker)
	c.addGovernanceContractRoutes(publicAPI, mocker)
}

func (c *Controller) RegisterAdmin(adminAPI echoswagger.ApiGroup, mocker interfaces.Mocker) {
}
