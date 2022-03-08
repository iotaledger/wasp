// Copyright 2020 IOTA Stiftung
// SPDX-License-Identifier: Apache-2.0

package main

import (
	"encoding/hex"
	"fmt"
	"time"

	"github.com/ethereum/go-ethereum/core"
	"github.com/ethereum/go-ethereum/crypto"
	"github.com/iotaledger/wasp/packages/evm/evmtest"
	"github.com/iotaledger/wasp/packages/evm/evmtypes"
	"github.com/iotaledger/wasp/packages/evm/jsonrpc"
	"github.com/iotaledger/wasp/packages/kv/codec"
	"github.com/iotaledger/wasp/packages/kv/dict"
	"github.com/iotaledger/wasp/packages/solo"
	"github.com/iotaledger/wasp/packages/vm/core/evm"
	"github.com/iotaledger/wasp/packages/vm/core/root"
	"github.com/iotaledger/wasp/tools/evm/evmcli"
	"github.com/iotaledger/wasp/tools/wasp-cli/log"
	"github.com/spf13/cobra"
)

var (
	evmParams     evmcli.DeployParams
	jsonRPCServer evmcli.JSONRPCServer
)

func main() {
	cmd := &cobra.Command{
		Args:  cobra.NoArgs,
		Run:   start,
		Use:   "evmemulator",
		Short: "evmemulator runs an instance of the evm contract with Solo as backend",
		Long: fmt.Sprintf(`evmemulator runs an instance of the evm contract with Solo as backend.

evmemulator does the following:

- Starts a Solo environment (a framework for running local ISC chains in-memory)
- Deploys an ISC chain
- Deploys the evm ISC contract (which runs an Ethereum chain on top of the ISC chain)
- Starts a JSON-RPC server with the deployed ISC contract as backend

You can connect any Ethereum tool (eg Metamask) to this JSON-RPC server and use it for testing Ethereum contracts running on ISCP.

The default genesis allocation is: %s:%d
                                   private key: %s

By default the server has no unlocked accounts. To send transactions, either:

- use eth_sendRawTransaction
- configure an unlocked account with --account, and use eth_sendTransaction
`,
			evmtest.FaucetAddress,
			evmtest.FaucetSupply,
			hex.EncodeToString(crypto.FromECDSA(evmtest.FaucetKey)),
		),
	}

	log.Init(cmd)

	evmParams.InitFlags(cmd)
	jsonRPCServer.InitFlags(cmd)

	err := cmd.Execute()
	log.Check(err)
}

func start(cmd *cobra.Command, args []string) {
	env := solo.New(solo.NewTestContext("evmemulator"), &solo.InitOptions{Debug: log.DebugFlag, PrintStackTrace: log.DebugFlag})

	chainOwner, _ := env.NewKeyPairWithFunds()
	chain := env.NewChain(chainOwner, "iscpchain", dict.Dict{
		root.ParamEVM(evm.FieldChainID): codec.EncodeUint16(uint16(evmParams.ChainID)),
		root.ParamEVM(evm.FieldGenesisAlloc): evmtypes.EncodeGenesisAlloc(evmParams.GetGenesis(core.GenesisAlloc{
			evmtest.FaucetAddress: {Balance: evmtest.FaucetSupply},
		})),
		root.ParamEVM(evm.FieldGasRatio):        evmParams.GasRatio.Bytes(),
		root.ParamEVM(evm.FieldBlockGasLimit):   codec.EncodeUint64(evmParams.BlockGasLimit),
		root.ParamEVM(evm.FieldBlockKeepAmount): codec.EncodeInt32(evmParams.BlockKeepAmount),
	})

	if evmParams.BlockTime > 0 {
		_, err := chain.PostRequestSync(
			solo.NewCallParams(evm.Contract.Name, evm.FuncSetBlockTime.Name,
				evm.FieldBlockTime, evmParams.BlockTime,
			).AddAssetsIotas(1),
			chain.OriginatorPrivateKey,
		)
		log.Check(err)
		go func() {
			const step = 1 * time.Second
			for {
				time.Sleep(step)
				env.AdvanceClockBy(step, 1)
			}
		}()
	}

	signer, _ := env.NewKeyPairWithFunds()

	backend := jsonrpc.NewSoloBackend(env, chain, signer)
	jsonRPCServer.ServeJSONRPC(backend, evmParams.ChainID)
}
