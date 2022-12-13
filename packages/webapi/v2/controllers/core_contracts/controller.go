package corecontracts

import (
	"net/http"

	"github.com/iotaledger/wasp/packages/authentication/shared/permissions"
	"github.com/iotaledger/wasp/packages/webapi/v2/models"

	"github.com/iotaledger/wasp/packages/webapi/v2/controllers/core_contracts/internal"

	"github.com/iotaledger/wasp/packages/webapi/v2/apierrors"

	"github.com/iotaledger/wasp/packages/kv/dict"

	"github.com/iotaledger/wasp/packages/isc"
	"github.com/labstack/echo/v4"

	"github.com/iotaledger/wasp/packages/authentication"
	"github.com/pangpanglabs/echoswagger/v2"

	loggerpkg "github.com/iotaledger/hive.go/core/logger"
	"github.com/iotaledger/wasp/packages/webapi/v2/interfaces"
)

type Controller struct {
	log *loggerpkg.Logger

	accounts   *internal.Accounts
	blob       *internal.Blob
	blocklog   *internal.BlockLog
	errors     *internal.Errors
	governance *internal.Governance

	vmService interfaces.VMService
}

func NewCoreContractsController(log *loggerpkg.Logger, vmService interfaces.VMService) interfaces.APIController {
	return &Controller{
		log: log,

		accounts:   internal.NewAccounts(vmService),
		blob:       internal.NewBlob(vmService),
		blocklog:   internal.NewBlockLog(vmService),
		errors:     internal.NewErrors(vmService),
		governance: internal.NewGovernance(vmService),
		vmService:  vmService,
	}
}

func (c *Controller) Name() string {
	return "corecontracts"
}

func (c *Controller) ExecuteCallView(e echo.Context, contract, entrypoint isc.Hname, params dict.Dict) (dict.Dict, error) {
	chainID, err := isc.ChainIDFromString(e.Param("chainID"))

	if err != nil {
		return nil, apierrors.InvalidPropertyError("chainID", err)
	}

	ret, err := c.vmService.CallViewByChainID(chainID, contract, entrypoint, params)

	return ret, apierrors.ContractExecutionError(err)
}

func (c *Controller) RegisterPublic(publicAPI echoswagger.ApiGroup, mocker interfaces.Mocker) {
	publicAPI.GET("chains/:chainID/accounts", c.getAccounts).
		AddResponse(http.StatusUnauthorized, "Unauthorized (Wrong permissions, missing token)", authentication.ValidationError{}, nil).
		AddResponse(http.StatusOK, "A list of all accounts", mocker.Get(AccountListResponse{}), nil).
		SetOperationId("coreGetAccounts").
		SetSummary("Get a list of all accounts")

	publicAPI.GET("chains/:chainID/accounts/:accountID", c.getAccounts).
		AddResponse(http.StatusUnauthorized, "Unauthorized (Wrong permissions, missing token)", authentication.ValidationError{}, nil).
		AddResponse(http.StatusOK, "A list of all accounts", mocker.Get(AccountListResponse{}), nil).
		SetOperationId("coreGetAccounts").
		SetSummary("Get a list of all accounts")

	publicAPI.GET("chains/:chainID/governance/info", c.getChainInfo, authentication.ValidatePermissions([]string{permissions.ChainRead})).
		AddParamPath("", "chainID", "ChainID (Bech32)").
		AddResponse(http.StatusUnauthorized, "Unauthorized (Wrong permissions, missing token)", authentication.ValidationError{}, nil).
		AddResponse(http.StatusOK, "Information about a specific chain", mocker.Get(models.ChainInfoResponse{}), nil).
		SetOperationId("getChainInfo").
		SetSummary("Get information about a specific chain")
}

func (c *Controller) RegisterAdmin(adminAPI echoswagger.ApiGroup, mocker interfaces.Mocker) {

}
