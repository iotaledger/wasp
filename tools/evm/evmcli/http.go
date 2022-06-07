// Copyright 2020 IOTA Stiftung
// SPDX-License-Identifier: Apache-2.0

package evmcli

import (
	"crypto/ecdsa"
	"errors"
	"fmt"
	"net/http"

	"github.com/ethereum/go-ethereum/crypto"
	"github.com/ethereum/go-ethereum/rpc"
	"github.com/iotaledger/wasp/packages/evm/jsonrpc"
	"github.com/iotaledger/wasp/tools/wasp-cli/log"
	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"
	"github.com/spf13/cobra"
)

type JSONRPCServer struct {
	listenAddr       string
	corsAllowOrigins []string
	unlockedAccount  string
}

func (j *JSONRPCServer) InitFlags(cmd *cobra.Command) {
	cmd.Flags().StringVarP(&j.listenAddr, "listen", "l", ":8545", "JSON-RPC listen address")
	cmd.Flags().StringSliceVarP(&j.corsAllowOrigins, "cors", "", []string{"*"}, "CORS allow origins")
	cmd.Flags().StringVarP(&j.unlockedAccount, "account", "", "", "unlocked account (hex-encoded private key)")
}

func (j *JSONRPCServer) getUnlockedAccount() []*ecdsa.PrivateKey {
	if j.unlockedAccount == "" {
		return nil
	}
	account, err := crypto.HexToECDSA(j.unlockedAccount)
	log.Check(err)
	return []*ecdsa.PrivateKey{account}
}

func (j *JSONRPCServer) ServeJSONRPC(backend jsonrpc.ChainBackend, chainID uint16) {
	evmChain := jsonrpc.NewEVMChain(backend, chainID)

	accountManager := jsonrpc.NewAccountManager(j.getUnlockedAccount())

	rpcsrv := jsonrpc.NewServer(evmChain, accountManager)
	defer rpcsrv.Stop()

	j.serveHTTP(rpcsrv)
}

func (j *JSONRPCServer) serveHTTP(rpcsrv *rpc.Server) {
	e := echo.New()
	e.HideBanner = true
	e.HidePort = true
	e.Use(middleware.LoggerWithConfig(middleware.LoggerConfig{
		Format: "[${time_rfc3339}] ${status} ${method} ${path} (${remote_ip}) ${latency_human}\n",
		Output: e.Logger.Output(),
	}))
	if log.DebugFlag {
		e.Use(middleware.BodyDump(func(c echo.Context, reqBody, resBody []byte) {
			fmt.Printf("REQUEST:  %s\n", string(reqBody))
			fmt.Printf("RESPONSE: %s\n", string(resBody))
		}))
	}
	e.Use(middleware.CORSWithConfig(middleware.CORSConfig{
		AllowOrigins: j.corsAllowOrigins,
		AllowMethods: []string{http.MethodPost, http.MethodGet},
		AllowHeaders: []string{"*"},
	}))
	e.Any("/", echo.WrapHandler(rpcsrv))

	fmt.Printf("Starting JSON-RPC server on %s\n", j.listenAddr)
	if err := e.Start(j.listenAddr); err != nil {
		if !errors.Is(err, http.ErrServerClosed) {
			log.Check(err)
		}
	}
}
