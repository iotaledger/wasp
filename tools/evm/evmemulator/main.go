// Copyright 2020 IOTA Stiftung
// SPDX-License-Identifier: Apache-2.0

package main

import (
	"time"

	"github.com/spf13/cobra"

	"github.com/iotaledger/wasp/v2/packages/evm/jsonrpc"
	"github.com/iotaledger/wasp/v2/packages/metrics"
	"github.com/iotaledger/wasp/v2/packages/testutil/l1starter"
	"github.com/iotaledger/wasp/v2/tools/evm/evmemulator/pkg/chains"
	"github.com/iotaledger/wasp/v2/tools/evm/evmemulator/pkg/cli"
	"github.com/iotaledger/wasp/v2/tools/evm/evmemulator/pkg/genesis"
	"github.com/iotaledger/wasp/v2/tools/evm/evmemulator/pkg/server"
	"github.com/iotaledger/wasp/v2/tools/wasp-cli/log"
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
	cmd.PersistentFlags().StringVarP(&cli.ListenAddress, "listen", "l", ":8545", "listen address")
	cmd.PersistentFlags().StringVar(&cli.EngineListenAddress, "engine-listen", ":8551", "engine (consensus) JSON-RPC listen address")
	cmd.PersistentFlags().StringVar(
		&cli.NodeLaunchMode,
		"node-launch-mode",
		string(cli.EnumNodeLaunchModeStandalone),
		"How to launch the L1 node: 'standalone' (start container) or 'docker-compose' (wait for external service)",
	)
	cmd.PersistentFlags().StringVar(&cli.RemoteHost, "remote-host", "http://localhost", "remote host")
	cmd.PersistentFlags().StringVar(
		&cli.GenesisJsonPath,
		"genesis",
		"",
		"path to the genesis JSON file",
	)
	cmd.PersistentFlags().BoolVar(&cli.IsHive, "hive", false, "whether running for hive tests")
	cmd.PersistentFlags().BoolVar(&cli.LogBodies, "log-bodies", false, "log JSON-RPC request/response bodies (verbose)")

	err := cmd.Execute()
	log.Check(err)
}

func start(cmd *cobra.Command, args []string) {
	var cancel func()
	if cli.NodeLaunchMode == string(cli.EnumNodeLaunchModeStandalone) {
		cancel = l1starter.TestLocal()
	} else if cli.NodeLaunchMode == string(cli.EnumNodeLaunchModeDockerCompose) {
		cancel = l1starter.TestExternal(cli.RemoteHost)
	}
	defer cancel()

	g, err := genesis.InitGenesis(cli.GenesisJsonPath)
	if err != nil {
		log.Fatalf("failed to initialize genesis: %v", err)
	}
	g = genesis.RegulateGenesisAccountBalance(g)

	log.Printf("Initialize Solo Env\n")
	initSoloTime := time.Now()
	ctx, chain := chains.InitSolo(g)
	defer ctx.CleanupAll()
	log.Printf("Finish Initializing Solo Env: %s\n", time.Since(initSoloTime))

	accounts := chains.CreateAccounts(chain)

	jsonRPCServer, err := jsonrpc.NewServer(
		chain.EVM(),
		jsonrpc.NewAccountManager(accounts),
		metrics.NewChainWebAPIMetricsProvider().CreateForChain(chain.ChainID),
		jsonrpc.ParametersDefault(),
	)
	log.Check(err)

	go func() {
		log.Printf("starting JSONRPC server on %s...\n", cli.ListenAddress)
		e := server.StartServer(jsonRPCServer, cli.ListenAddress, cli.LogBodies)
		err = e.ListenAndServe()
		log.Check(err)
	}()

	engineAPI, err := jsonrpc.NewEngineAPI(chain.EVM(), jsonrpc.NewAccountManager(accounts),
		metrics.NewChainWebAPIMetricsProvider().CreateForChain(chain.ChainID),
		jsonrpc.ParametersDefault())

	log.Printf("starting Engine JSONRPC server on %s...\n", cli.EngineListenAddress)
	e := server.StartServer(engineAPI, cli.EngineListenAddress, cli.LogBodies)
	err = e.ListenAndServe()
	log.Check(err)
}
