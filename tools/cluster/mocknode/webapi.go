package mocknode

import (
	"context"
	"errors"
	"fmt"
	"net"
	"net/http"
	"time"

	"github.com/iotaledger/goshimmer/packages/jsonmodels"
	"github.com/iotaledger/goshimmer/packages/ledgerstate"
	"github.com/iotaledger/goshimmer/packages/tangle"
	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"
	"golang.org/x/xerrors"
)

func (m *MockNode) startWebAPI(bindAddress string) error {
	l, err := net.Listen("tcp", bindAddress)
	if err != nil {
		return err
	}

	e := echo.New()
	e.HideBanner = true
	e.HidePort = true
	e.Use(middleware.LoggerWithConfig(middleware.LoggerConfig{
		Format: `${time_rfc3339_nano} ${remote_ip} ${method} ${uri} ${status} error="${error}"` + "\n",
	}))
	e.Listener = l

	m.addEndpoints(e)

	go func() {
		if err := e.Start(""); err != nil {
			if !errors.Is(err, http.ErrServerClosed) {
				m.log.Error(err)
			}
		}
	}()

	go func() {
		<-m.shutdownSignal

		ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
		defer cancel()
		if err := e.Shutdown(ctx); err != nil {
			e.Logger.Fatal(err)
		}
	}()

	return nil
}

func (m *MockNode) addEndpoints(e *echo.Echo) {
	// These endpoints share the same schema as the endpoints in Goshimmer,
	// so they should work with the official Goshimmer client.

	e.GET("ledgerstate/addresses/:address/unspentOutputs", m.unspentOutputsHandler)
	e.GET("ledgerstate/transactions/:transactionID/inclusionState", m.getTransactionInclusionStateHandler)
	e.POST("ledgerstate/transactions", m.sendTransactionHandler)
	e.POST("faucet", m.requestFundsHandler)
}

func (m *MockNode) unspentOutputsHandler(c echo.Context) error {
	address, err := ledgerstate.AddressFromBase58EncodedString(c.Param("address"))
	if err != nil {
		return c.JSON(http.StatusBadRequest, jsonmodels.NewErrorResponse(err))
	}

	var outputs []ledgerstate.Output
	m.Ledger.GetUnspentOutputs(address, func(output ledgerstate.Output) {
		outputs = append(outputs, output.Clone())
	})

	return c.JSON(http.StatusOK, jsonmodels.NewGetAddressResponse(address, outputs))
}

func (m *MockNode) getTransactionInclusionStateHandler(c echo.Context) error {
	txID, err := ledgerstate.TransactionIDFromBase58(c.Param("transactionID"))
	if err != nil {
		return c.JSON(http.StatusBadRequest, jsonmodels.NewErrorResponse(err))
	}

	found := m.Ledger.GetConfirmedTransaction(txID, func(tx *ledgerstate.Transaction) {})
	if !found {
		return c.JSON(http.StatusNotFound, jsonmodels.NewErrorResponse(xerrors.Errorf("failed to load Transaction with %s", txID)))
	}
	return c.JSON(http.StatusOK, &jsonmodels.TransactionInclusionState{
		TransactionID: txID.Base58(),
		Pending:       false,
		Confirmed:     true,
		Rejected:      false,
		Conflicting:   false,
	})
}

func (m *MockNode) sendTransactionHandler(c echo.Context) error {
	var request jsonmodels.PostTransactionRequest
	if err := c.Bind(&request); err != nil {
		return c.JSON(http.StatusBadRequest, &jsonmodels.PostTransactionResponse{Error: err.Error()})
	}

	// parse tx
	tx, _, err := ledgerstate.TransactionFromBytes(request.TransactionBytes)
	if err != nil {
		return c.JSON(http.StatusBadRequest, &jsonmodels.PostTransactionResponse{Error: err.Error()})
	}

	err = m.Ledger.PostTransaction(tx)
	if err != nil {
		return c.JSON(http.StatusBadRequest, &jsonmodels.PostTransactionResponse{Error: err.Error()})
	}

	return c.JSON(http.StatusOK, &jsonmodels.PostTransactionResponse{TransactionID: tx.ID().Base58()})
}

func (m *MockNode) requestFundsHandler(c echo.Context) error {
	var request jsonmodels.FaucetRequest
	if err := c.Bind(&request); err != nil {
		return c.JSON(http.StatusBadRequest, jsonmodels.FaucetResponse{Error: err.Error()})
	}

	addr, err := ledgerstate.AddressFromBase58EncodedString(request.Address)
	if err != nil {
		return c.JSON(http.StatusBadRequest, jsonmodels.FaucetResponse{Error: fmt.Sprintf("invalid address (%s): %s", request.Address, err.Error())})
	}

	err = m.Ledger.RequestFunds(addr)
	if err != nil {
		return c.JSON(http.StatusBadRequest, jsonmodels.FaucetResponse{Error: fmt.Sprintf("ledger.RequestFunds: %s", err.Error())})
	}

	return c.JSON(http.StatusOK, jsonmodels.FaucetResponse{ID: tangle.EmptyMessageID.String()})
}
