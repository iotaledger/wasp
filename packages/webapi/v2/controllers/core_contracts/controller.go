package corecontracts

import (
	"net/http"

	"github.com/iotaledger/wasp/packages/webapi/v2/apierrors"

	"github.com/iotaledger/wasp/packages/kv/dict"

	"github.com/iotaledger/wasp/packages/isc"
	"github.com/labstack/echo/v4"

	"github.com/iotaledger/wasp/packages/authentication"
	"github.com/pangpanglabs/echoswagger/v2"

	"github.com/iotaledger/wasp/packages/webapi/v2/models"

	loggerpkg "github.com/iotaledger/hive.go/core/logger"
	"github.com/iotaledger/wasp/packages/webapi/v2/interfaces"
)

type Controller struct {
	log *loggerpkg.Logger

	vmService interfaces.VMService
}

func NewCoreContractsController(log *loggerpkg.Logger, vmService interfaces.VMService) interfaces.APIController {
	return &Controller{
		log:       log,
		vmService: vmService,
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
		AddResponse(http.StatusOK, "A list of all users", mocker.Get([]models.User{}), nil).
		SetOperationId("getUsers").
		SetSummary("Get a list of all users")
}

func (c *Controller) RegisterAdmin(adminAPI echoswagger.ApiGroup, mocker interfaces.Mocker) {
}
