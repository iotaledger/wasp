// Copyright 2020 IOTA Stiftung
// SPDX-License-Identifier: Apache-2.0

package main

import (
	"encoding/hex"
	"fmt"

	"github.com/ethereum/go-ethereum/core"
	"github.com/ethereum/go-ethereum/crypto"
	"github.com/iotaledger/wasp/contracts/native/evm"
	"github.com/iotaledger/wasp/packages/evm/evmtest"
	"github.com/iotaledger/wasp/packages/evm/evmtypes"
	"github.com/iotaledger/wasp/packages/evm/jsonrpc"
	"github.com/iotaledger/wasp/packages/kv/codec"
	"github.com/iotaledger/wasp/packages/solo"
	"github.com/iotaledger/wasp/tools/evm/evmcli"
	"github.com/iotaledger/wasp/tools/wasp-cli/log"
	"github.com/spf13/cobra"
)

var (
	deployParams  evmcli.DeployParams
	jsonRPCServer evmcli.JSONRPCServer
)

func main() {
	cmd := &cobra.Command{
		Args:  cobra.NoArgs,
		Run:   start,
		Use:   "evmemulator",
		Short: "evmemulator runs an instance of the evmchain/evmlight contract with Solo as backend",
		Long: fmt.Sprintf(`evmemulator runs an instance of the evmchain/evmlight contract with Solo as backend.

evmemulator does the following:

- Starts a Solo environment (a framework for running local ISCP chains in-memory)
- Deploys an ISCP chain
- Deploys the evmchain/evmlight ISCP contract (which runs an Ethereum chain on top of the ISCP chain)
- Starts a JSON-RPC server with the deployed ISCP contract as backend

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

	deployParams.InitFlags(cmd)
	jsonRPCServer.InitFlags(cmd)

	err := cmd.Execute()
	log.Check(err)
}

func start(cmd *cobra.Command, args []string) {
	env := solo.New(nil, log.DebugFlag, log.DebugFlag)

	chainOwner, _ := env.NewKeyPairWithFunds()
	chain := env.NewChain(chainOwner, "iscpchain")
	err := chain.DeployContract(chainOwner, deployParams.Name(), deployParams.EVMFlavor().ProgramHash,
		evm.FieldChainID, codec.EncodeUint16(uint16(deployParams.ChainID)),
		evm.FieldGenesisAlloc, evmtypes.EncodeGenesisAlloc(deployParams.GetGenesis(core.GenesisAlloc{
			evmtest.FaucetAddress: {Balance: evmtest.FaucetSupply},
		})),
		evm.FieldGasPerIota, deployParams.GasPerIOTA,
	)
	log.Check(err)

	signer, _ := env.NewKeyPairWithFunds()

	backend := jsonrpc.NewSoloBackend(env, chain, signer)
	jsonRPCServer.ServeJSONRPC(backend, deployParams.ChainID, deployParams.Name())
}
