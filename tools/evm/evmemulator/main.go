// Copyright 2020 IOTA Stiftung
// SPDX-License-Identifier: Apache-2.0

package main

import (
	"crypto/ecdsa"
	"encoding/hex"
	"fmt"
	"net/http"
	"os"

	"github.com/ethereum/go-ethereum/crypto"
	"github.com/spf13/cobra"

	"github.com/iotaledger/wasp/components/app"
	"github.com/iotaledger/wasp/packages/evm/jsonrpc"
	"github.com/iotaledger/wasp/packages/isc"
	"github.com/iotaledger/wasp/packages/kv/codec"
	"github.com/iotaledger/wasp/packages/kv/dict"
	"github.com/iotaledger/wasp/packages/metrics"
	"github.com/iotaledger/wasp/packages/origin"
	"github.com/iotaledger/wasp/packages/parameters"
	"github.com/iotaledger/wasp/packages/solo"
	"github.com/iotaledger/wasp/packages/vm/core/evm/emulator"
	"github.com/iotaledger/wasp/tools/wasp-cli/log"
)

type soloContext struct {
	cleanup []func()
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

func init() {
	parameters.InitL1(parameters.L1ForTesting)
}

func main() {
	cmd := &cobra.Command{
		Args:  cobra.NoArgs,
		Run:   start,
		Use:   "evmemulator",
		Short: "evmemulator runs a JSONRPC server with Solo as backend",
		Long: fmt.Sprintf(`evmemulator runs a JSONRPC server with Solo as backend.

evmemulator does the following:

- Starts an ISC chain in a Solo environment
- Initializes 10 ethereum accounts with funds (private keys and addresses printed after init)
- Starts a JSONRPC server

You can connect any Ethereum tool (eg Metamask) to this JSON-RPC server and use it for testing Ethereum contracts.

Note: chain data is stored in-memory and will be lost upon termination.
`,
		),
	}

	log.Init(cmd)

	err := cmd.Execute()
	log.Check(err)
}

func start(cmd *cobra.Command, args []string) {
	ctx := &soloContext{}
	defer func() {
		for i := len(ctx.cleanup) - 1; i >= 0; i-- {
			ctx.cleanup[i]()
		}
	}()

	env := solo.New(ctx, &solo.InitOptions{Debug: log.DebugFlag, PrintStackTrace: log.DebugFlag})

	chainOwner, chainOwnerAddr := env.NewKeyPairWithFunds()
	chain, _ := env.NewChainExt(chainOwner, 1*isc.Million, "evmemulator", dict.Dict{
		origin.ParamChainOwner:      isc.NewAgentID(chainOwnerAddr).Bytes(),
		origin.ParamEVMChainID:      codec.EncodeUint16(1074),
		origin.ParamBlockKeepAmount: codec.EncodeInt32(emulator.BlockKeepAll),
		origin.ParamWaspVersion:     codec.EncodeString(app.Version),
	})

	var accounts []*ecdsa.PrivateKey
	log.Printf("creating accounts with funds...\n")
	header := []string{"private key", "address"}
	var rows [][]string
	for i := 0; i < len(solo.EthereumAccounts); i++ {
		pk, addr := chain.EthereumAccountByIndexWithL2Funds(i)
		accounts = append(accounts, pk)
		rows = append(rows, []string{hex.EncodeToString(crypto.FromECDSA(pk)), addr.String()})
	}
	log.PrintTable(header, rows)

	srv, err := jsonrpc.NewServer(
		chain.EVM(),
		jsonrpc.NewAccountManager(accounts),
		metrics.NewChainWebAPIMetricsProvider().CreateForChain(chain.ChainID),
		jsonrpc.ParametersDefault(),
	)
	log.Check(err)

	const addr = ":8545"
	s := &http.Server{
		Addr:    addr,
		Handler: srv,
	}
	log.Printf("starting JSONRPC server on %s...\n", addr)
	err = s.ListenAndServe()
	log.Check(err)
}
