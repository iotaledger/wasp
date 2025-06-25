// Copyright 2020 IOTA Stiftung
// SPDX-License-Identifier: Apache-2.0

package main

import (
	"crypto/ecdsa"
	"encoding/hex"
	"net/http"
	"os"

	"github.com/ethereum/go-ethereum/crypto"
	"github.com/spf13/cobra"

	"github.com/iotaledger/wasp/v2/packages/evm/jsonrpc"
	"github.com/iotaledger/wasp/v2/packages/isc"
	"github.com/iotaledger/wasp/v2/packages/metrics"
	"github.com/iotaledger/wasp/v2/packages/solo"
	"github.com/iotaledger/wasp/v2/packages/testutil/l1starter"
	"github.com/iotaledger/wasp/v2/packages/vm/core/evm/emulator"
	"github.com/iotaledger/wasp/v2/tools/wasp-cli/log"
)

type soloContext struct {
	cleanup []func()
}

func (s *soloContext) cleanupAll() {
	for i := len(s.cleanup) - 1; i >= 0; i-- {
		s.cleanup[i]()
	}
}

func (s *soloContext) Cleanup(f func()) {
	s.cleanup = append(s.cleanup, f)
}

func (*soloContext) Errorf(format string, args ...interface{}) {
	log.Printf("error: "+format, args)
}

func (*soloContext) FailNow() {
	os.Exit(1)
}

func (s *soloContext) Fatalf(format string, args ...any) {
	log.Printf("fatal: "+format, args)
	s.FailNow()
}

func (*soloContext) Helper() {
}

func (*soloContext) Logf(format string, args ...any) {
	log.Printf(format, args...)
}

func (*soloContext) Name() string {
	return "evmemulator"
}

var listenAddress string = ":8545"
var nodeLaunchMode string

type NodeLaunchMode string

const (
	EnumNodeLaunchModeStandalone    NodeLaunchMode = "standalone"
	EnumNodeLaunchModeDockerCompose NodeLaunchMode = "docker-compose"
)

func main() {
	cmd := &cobra.Command{
		Args:  cobra.NoArgs,
		Run:   start,
		Use:   "evmemulator",
		Short: "evmemulator runs a JSONRPC server with Solo as backend",
		Long: `evmemulator runs a JSONRPC server with Solo as backend.

evmemulator does the following:

- Starts an ISC chain in a Solo environment
- Initializes 10 ethereum accounts with funds (private keys and addresses printed after init)
- Starts a JSONRPC server at http://localhost:8545 (websocket: ws://localhost:8545/ws)

You can connect any Ethereum tool (eg Metamask) to this JSON-RPC server and use it for testing Ethereum contracts.

Note: chain data is stored in-memory and will be lost upon termination.
`,
	}

	log.Init(cmd)
	cmd.PersistentFlags().StringVarP(&listenAddress, "listen", "l", ":8545", "listen address")
	cmd.PersistentFlags().StringVar(
		&nodeLaunchMode,
		"node-launch-mode",
		string(EnumNodeLaunchModeStandalone),
		"How to launch the L1 node: 'standalone' (start container) or 'docker-compose' (wait for external service)",
	)

	err := cmd.Execute()
	log.Check(err)
}

func initSolo() (*soloContext, *solo.Chain) {
	ctx := &soloContext{}

	env := solo.New(ctx, &solo.InitOptions{Debug: log.DebugFlag, PrintStackTrace: log.DebugFlag})

	chainAdmin, _ := env.NewKeyPairWithFunds()
	chain, _ := env.NewChainExt(chainAdmin, 1*isc.Million, "evmemulator", 1074, emulator.BlockKeepAll)
	return ctx, chain
}

func createAccounts(chain *solo.Chain) (accounts []*ecdsa.PrivateKey) {
	log.Printf("creating accounts with funds...\n")
	header := []string{"private key", "address"}
	var rows [][]string
	for i := 0; i < len(solo.EthereumAccounts); i++ {
		pk, addr := chain.EthereumAccountByIndexWithL2Funds(i)
		accounts = append(accounts, pk)
		rows = append(rows, []string{hex.EncodeToString(crypto.FromECDSA(pk)), addr.String()})
	}
	log.PrintTable(header, rows)
	return accounts
}

func start(cmd *cobra.Command, args []string) {
	var cancel func()
	if nodeLaunchMode == string(EnumNodeLaunchModeStandalone) {
		cancel = l1starter.TestLocal()
	} else if nodeLaunchMode == string(EnumNodeLaunchModeDockerCompose) {
		cancel = l1starter.TestLocalExternal()
	}
	defer cancel()

	ctx, chain := initSolo()
	defer ctx.cleanupAll()

	accounts := createAccounts(chain)

	jsonRPCServer, err := jsonrpc.NewServer(
		chain.EVM(),
		jsonrpc.NewAccountManager(accounts),
		metrics.NewChainWebAPIMetricsProvider().CreateForChain(chain.ChainID),
		jsonrpc.ParametersDefault(),
	)
	log.Check(err)

	mux := http.NewServeMux()
	mux.Handle("/ws", http.HandlerFunc(func(w http.ResponseWriter, req *http.Request) {
		jsonRPCServer.WebsocketHandler([]string{"*"}).ServeHTTP(w, req)
	}))
	mux.Handle("/", jsonRPCServer)

	s := &http.Server{
		Addr:    listenAddress,
		Handler: mux,
	}
	log.Printf("starting JSONRPC server on %s...\n", listenAddress)
	err = s.ListenAndServe()
	log.Check(err)
}
