package main

import (
	"encoding/hex"
	"fmt"

	"github.com/ethereum/go-ethereum/crypto"
	"github.com/iotaledger/wasp/contracts/native/evmchain"
	"github.com/iotaledger/wasp/packages/evm/evmtest"
	"github.com/iotaledger/wasp/packages/evm/jsonrpc"
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
		Short: "evmemulator runs an instance of the evmchain contract with Solo as backend",
		Long: fmt.Sprintf(`evmemulator runs an instance of the evmchain contract with Solo as backend.

evmemulator does the following:

- Starts a Solo environment (a framework for running local ISCP chains in-memory)
- Deploys an ISCP chain
- Deploys the evmchain ISCP contract (which runs an Ethereum chain on top of the ISCP chain)
- Starts a JSON-RPC server with the evmchain contract as backend

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
	env := solo.New(solo.NewFakeTestingT("evmproxy"), true, false)

	chainOwner, _ := env.NewKeyPairWithFunds()
	chain := env.NewChain(chainOwner, "iscpchain")
	err := chain.DeployContract(chainOwner, "evmchain", evmchain.Interface.ProgramHash,
		evmchain.FieldGenesisAlloc, evmchain.EncodeGenesisAlloc(deployParams.GetGenesis()),
		evmchain.FieldGasPerIota, deployParams.GasPerIOTA,
	)
	log.Check(err)

	signer, _ := env.NewKeyPairWithFunds()

	backend := jsonrpc.NewSoloBackend(env, chain, signer)
	jsonRPCServer.ServeJSONRPC(backend)
}
