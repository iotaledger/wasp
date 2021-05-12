package main

import (
	"errors"
	"fmt"
	"log"
	"net/http"

	"github.com/ethereum/go-ethereum/rpc"
	"github.com/iotaledger/wasp/tools/evmproxy/service"
	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"
)

func main() {
	rpcsrv := rpc.NewServer()
	defer rpcsrv.Stop()

	soloEVMChain := service.NewSoloEVMChain()
	for _, srv := range []struct {
		namespace string
		service   interface{}
	}{
		{"eth", service.NewEthService(soloEVMChain)},
		{"net", service.NewNetService(soloEVMChain)},
	} {
		err := rpcsrv.RegisterName(srv.namespace, srv.service)
		if err != nil {
			log.Fatal(err.Error())
		}
	}

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
	e.Any("/", echo.WrapHandler(rpcsrv))

	listenAddr := ":8545"
	fmt.Printf("Starting JSON-RPC server on %s\n", listenAddr)
	if err := e.Start(listenAddr); err != nil {
		if !errors.Is(err, http.ErrServerClosed) {
			fmt.Println(err.Error())
		}
	}
}
