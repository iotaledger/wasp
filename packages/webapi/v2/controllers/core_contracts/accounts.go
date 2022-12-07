package corecontracts

import (
	"net/http"

	"github.com/iotaledger/wasp/packages/webapi/v2/models"

	"github.com/iotaledger/wasp/packages/vm/core/accounts"

	"github.com/iotaledger/wasp/packages/isc"

	"github.com/labstack/echo/v4"
)

func (c *Controller) getAccounts(e echo.Context) error {
	ret, err := c.ExecuteCallView(e, accounts.Contract.Hname(), accounts.ViewAccounts.Hname(), nil)

	if err != nil {
		return err
	}

	accountIds := make([]string, 0)

	for k, _ := range ret {
		agentID, _ := isc.AgentIDFromBytes([]byte(k))
		accountIds = append(accountIds, agentID.String())
	}

	return e.JSON(http.StatusCreated, models.AccountsResponse{AccountIDs: accountIds})
}
