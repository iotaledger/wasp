package mocknode

import (
	"errors"
	"net/http"
	"time"

	"github.com/iotaledger/goshimmer/packages/ledgerstate"
	"github.com/iotaledger/goshimmer/packages/tangle"
	webapi_faucet "github.com/iotaledger/goshimmer/plugins/webapi/faucet"
	"github.com/iotaledger/goshimmer/plugins/webapi/value"
	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"
)

func (m *MockNode) startWebAPI(bindAddress string) {
	e := echo.New()
	e.HideBanner = true
	e.HidePort = true
	e.Use(middleware.LoggerWithConfig(middleware.LoggerConfig{
		Format: `${time_rfc3339_nano} ${remote_ip} ${method} ${uri} ${status} error="${error}"` + "\n",
	}))

	m.addEndpoints(e)

	stopped := make(chan struct{})
	go func() {
		defer close(stopped)
		if err := e.Start(bindAddress); err != nil {
			if !errors.Is(err, http.ErrServerClosed) {
				m.log.Error(err)
			}
		}
	}()

	select {
	case <-m.shutdownSignal:
	case <-stopped:
	}
}

func (m *MockNode) addEndpoints(e *echo.Echo) {
	// These endpoints share the same schema as the endpoints in Goshimmer,
	// so they should work with the official Goshimmer client.

	e.POST("value/unspentOutputs", m.unspentOutputsHandler)
	e.GET("value/transactionByID", m.getTransactionByIDHandler)
	e.POST("value/sendTransaction", m.sendTransactionHandler)
	e.POST("faucet", m.requestFundsHandler)
}

func (m *MockNode) unspentOutputsHandler(c echo.Context) error {
	var request value.UnspentOutputsRequest
	if err := c.Bind(&request); err != nil {
		m.log.Error(err)
		return c.JSON(http.StatusBadRequest, value.UnspentOutputsResponse{Error: err.Error()})
	}

	var unspents []value.UnspentOutput
	for _, strAddress := range request.Addresses {
		address, err := ledgerstate.AddressFromBase58EncodedString(strAddress)
		if err != nil {
			return c.JSON(http.StatusBadRequest, value.UnspentOutputsResponse{Error: err.Error()})
		}

		outputids := make([]value.OutputID, 0)
		// get outputids by address
		m.Ledger.GetUnspentOutputs(address, func(output ledgerstate.Output) {
			// iterate balances
			var b []value.Balance
			output.Balances().ForEach(func(color ledgerstate.Color, balance uint64) bool {
				b = append(b, value.Balance{
					Value: int64(balance),
					Color: color.String(),
				})
				return true
			})

			var timestamp time.Time
			m.Ledger.GetConfirmedTransaction(output.ID().TransactionID(), func(tx *ledgerstate.Transaction) {
				timestamp = tx.Essence().Timestamp()
			})

			outputids = append(outputids, value.OutputID{
				ID:       output.ID().Base58(),
				Balances: b,
				InclusionState: value.InclusionState{
					Finalized: true,
					Confirmed: true,
				},
				Metadata: value.Metadata{Timestamp: timestamp},
			})
		})

		unspents = append(unspents, value.UnspentOutput{
			Address:   strAddress,
			OutputIDs: outputids,
		})
	}

	return c.JSON(http.StatusOK, value.UnspentOutputsResponse{UnspentOutputs: unspents})
}

func (m *MockNode) getTransactionByIDHandler(c echo.Context) error {
	txID, err := ledgerstate.TransactionIDFromBase58(c.QueryParam("txnID"))
	if err != nil {
		return c.JSON(http.StatusBadRequest, value.GetTransactionByIDResponse{Error: err.Error()})
	}

	var txn value.Transaction
	found := m.Ledger.GetConfirmedTransaction(txID, func(tx *ledgerstate.Transaction) {
		txn = value.ParseTransaction(tx)
	})
	if !found {
		return c.JSON(http.StatusNotFound, value.GetTransactionByIDResponse{Error: "Transaction not found"})
	}

	return c.JSON(http.StatusOK, value.GetTransactionByIDResponse{
		TransactionMetadata: value.TransactionMetadata{
			BranchID:   ledgerstate.MasterBranchID.String(),
			Solid:      true,
			Finalized:  true,
			LazyBooked: false,
		},
		Transaction: txn,
		InclusionState: value.InclusionState{
			Confirmed:   true,
			Conflicting: false,
			Liked:       true,
			Solid:       true,
			Rejected:    false,
			Finalized:   true,
			Preferred:   false,
		},
	})
}

func (m *MockNode) sendTransactionHandler(c echo.Context) error {
	var request value.SendTransactionRequest
	if err := c.Bind(&request); err != nil {
		return c.JSON(http.StatusBadRequest, value.SendTransactionResponse{Error: err.Error()})
	}

	// parse tx
	tx, _, err := ledgerstate.TransactionFromBytes(request.TransactionBytes)
	if err != nil {
		return c.JSON(http.StatusBadRequest, value.SendTransactionResponse{Error: err.Error()})
	}

	err = m.Ledger.PostTransaction(tx)
	if err != nil {
		return c.JSON(http.StatusBadRequest, value.SendTransactionResponse{Error: err.Error()})
	}

	return c.JSON(http.StatusOK, value.SendTransactionResponse{TransactionID: tx.ID().Base58()})
}

func (m *MockNode) requestFundsHandler(c echo.Context) error {
	var request webapi_faucet.Request
	var addr ledgerstate.Address
	if err := c.Bind(&request); err != nil {
		return c.JSON(http.StatusBadRequest, webapi_faucet.Response{Error: err.Error()})
	}

	addr, err := ledgerstate.AddressFromBase58EncodedString(request.Address)
	if err != nil {
		return c.JSON(http.StatusBadRequest, webapi_faucet.Response{Error: "Invalid address"})
	}

	err = m.Ledger.RequestFunds(addr)

	return c.JSON(http.StatusOK, webapi_faucet.Response{ID: tangle.EmptyMessageID.String()})
}
