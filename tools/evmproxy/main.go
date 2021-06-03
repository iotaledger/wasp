package main

import (
	"errors"
	"fmt"
	"net/http"

	"github.com/ethereum/go-ethereum/core"
	"github.com/ethereum/go-ethereum/rpc"
	"github.com/iotaledger/wasp/packages/evm/evmtest"
	"github.com/iotaledger/wasp/packages/evm/jsonrpc"
	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"
)

func main() {
	// TODO: use wasp backend
	// TODO: make genesis configurable
	solo := jsonrpc.NewSoloBackend(core.GenesisAlloc{
		evmtest.FaucetAddress: {Balance: evmtest.FaucetSupply},
	})
	soloEVMChain := jsonrpc.NewEVMChain(solo)

	// TODO: make signer key configurable
	signer, _ := solo.Env.NewKeyPairWithFunds()

	// TODO: make accounts configurable
	accountManager := jsonrpc.NewAccountManager(evmtest.Accounts)

	rpcsrv := jsonrpc.NewServer(soloEVMChain, signer, accountManager)
	defer rpcsrv.Stop()

	serveHTTP(rpcsrv)
}

func serveHTTP(rpcsrv *rpc.Server) {
	e := echo.New()
	e.HideBanner = true
	e.HidePort = true
	e.Use(middleware.LoggerWithConfig(middleware.LoggerConfig{
		Format: "[${time_rfc3339}] ${status} ${method} ${path} (${remote_ip}) ${latency_human}\n",
		Output: e.Logger.Output(),
	}))
	e.Use(middleware.BodyDump(func(c echo.Context, reqBody, resBody []byte) {
		fmt.Printf("REQUEST:  %s\n", string(reqBody))
		fmt.Printf("RESPONSE: %s\n", string(resBody))
	}))
	e.Use(middleware.CORSWithConfig(middleware.CORSConfig{
		AllowOrigins: []string{"*"}, // TODO make CORS configurable
		AllowMethods: []string{http.MethodPost, http.MethodGet},
		AllowHeaders: []string{"*"},
	}))
	e.Any("/", echo.WrapHandler(rpcsrv))

	listenAddr := ":8545"
	fmt.Printf("Starting JSON-RPC server on %s\n", listenAddr)
	if err := e.Start(listenAddr); err != nil {
		if !errors.Is(err, http.ErrServerClosed) {
			fmt.Println(err.Error())
		}
	}
}
