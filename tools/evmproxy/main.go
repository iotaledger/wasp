package main

import (
	"errors"
	"fmt"
	"log"
	"math/big"
	"net/http"

	"github.com/ethereum/go-ethereum/core"
	"github.com/ethereum/go-ethereum/crypto"
	"github.com/ethereum/go-ethereum/rpc"
	"github.com/iotaledger/wasp/tools/evmproxy/service"
	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"
)

var (
	faucetKey, _  = crypto.HexToECDSA("b71c71a67e1177ad4e901695e1b4b9ee17ae16c6668d313eac2f96dbcda3f291")
	faucetAddress = crypto.PubkeyToAddress(faucetKey.PublicKey)
	faucetSupply  = new(big.Int).Sub(new(big.Int).Lsh(big.NewInt(1), 256), big.NewInt(9))
)

func NewRPCServer(chain service.EVMChain) *rpc.Server {
	rpcsrv := rpc.NewServer()
	for _, srv := range []struct {
		namespace string
		service   interface{}
	}{
		{"eth", service.NewEthService(chain)},
		{"net", service.NewNetService(chain)},
	} {
		err := rpcsrv.RegisterName(srv.namespace, srv.service)
		if err != nil {
			log.Fatal(err.Error())
		}
	}
	return rpcsrv
}

func main() {
	soloEVMChain := service.NewSoloEVMChain(core.GenesisAlloc{
		faucetAddress: {Balance: faucetSupply},
	})

	rpcsrv := NewRPCServer(soloEVMChain)
	defer rpcsrv.Stop()

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
