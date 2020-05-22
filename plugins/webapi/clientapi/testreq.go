package clientapi

import (
	"github.com/iotaledger/wasp/packages/testapilib"
	"github.com/iotaledger/wasp/plugins/nodeconn"
	"github.com/iotaledger/wasp/plugins/webapi/misc"
	"github.com/labstack/echo"
)

// HandlerTestRequestTx handler creates testing request transaction and forwards it to node
func HandlerTestRequestTx(c echo.Context) error {
	var req testapilib.RequestTransactionJson

	if err := c.Bind(&req); err != nil {
		return misc.OkJsonErr(c, err)
	}

	tx, err := testapilib.TransactionFromJsonTesting(&req)
	if err != nil {
		return misc.OkJsonErr(c, err)
	}

	if err := nodeconn.PostTransactionToNode(tx.Transaction); err != nil {
		log.Debug("failed to post request transaction to node")
	} else {
		log.Debug("successfully forwarded request transaction to node")
	}
	return misc.OkJsonErr(c, nil)
}
