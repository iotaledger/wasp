//go:build !noevm
// +build !noevm

package chain

import (
	"github.com/iotaledger/wasp/contracts/native/evm"
	"github.com/iotaledger/wasp/contracts/native/evm/evmchain"
	"github.com/iotaledger/wasp/packages/evm/evmtypes"
	"github.com/iotaledger/wasp/packages/evm/jsonrpc"
	"github.com/iotaledger/wasp/packages/kv/codec"
	"github.com/iotaledger/wasp/packages/kv/dict"
	"github.com/iotaledger/wasp/tools/evm/evmcli"
	"github.com/iotaledger/wasp/tools/wasp-cli/log"
	"github.com/spf13/cobra"
)

func init() {
	plugins = append(plugins, func(chainCmd *cobra.Command) {
		evmCmd := &cobra.Command{
			Use:   "evm <command>",
			Short: "Interact with EVM chains",
			Args:  cobra.NoArgs,
			Run: func(cmd *cobra.Command, args []string) {
				log.Check(cmd.Help())
			},
		}
		chainCmd.AddCommand(evmCmd)

		initEVMDeploy(evmCmd)
		initJSONRPCCommand(evmCmd)
	})
}

func initEVMDeploy(evmCmd *cobra.Command) {
	var deployParams evmcli.DeployParams
	evmDeployCmd := &cobra.Command{
		Use:   "deploy",
		Short: "Deploy the evmchain/evmlight contract (i.e. create a new EVM chain)",
		Run: func(cmd *cobra.Command, args []string) {
			deployContract(deployParams.Name(), deployParams.Description(), deployParams.EVMFlavor().ProgramHash, dict.Dict{
				evm.FieldChainID:      codec.EncodeUint16(uint16(deployParams.ChainID)),
				evm.FieldGenesisAlloc: evmtypes.EncodeGenesisAlloc(deployParams.GetGenesis(nil)),
			})
			log.Printf("%s contract successfully deployed.\n", deployParams.Name())
		},
	}
	evmCmd.AddCommand(evmDeployCmd)

	deployParams.InitFlags(evmDeployCmd)
}

func initJSONRPCCommand(evmCmd *cobra.Command) {
	var jsonRPCServer evmcli.JSONRPCServer
	var chainID int
	var contractName string

	jsonRPCCmd := &cobra.Command{
		Args:  cobra.NoArgs,
		Use:   "jsonrpc",
		Short: "Start a JSON-RPC service to interact with an Ethereum blockchain running on ISCP",
		Long: `Start a JSON-RPC service to interact with an Ethereum blockchain running on ISCP.

By default the server has no unlocked accounts. To send transactions, either:

- use eth_sendRawTransaction
- configure an unlocked account with --account, and use eth_sendTransaction`,
		Run: func(cmd *cobra.Command, args []string) {
			backend := jsonrpc.NewWaspClientBackend(Client())
			jsonRPCServer.ServeJSONRPC(backend, chainID, contractName)
		},
	}

	jsonRPCServer.InitFlags(jsonRPCCmd)
	jsonRPCCmd.Flags().IntVarP(&chainID, "chainid", "", evm.DefaultChainID, "ChainID (used for signing transactions)")
	jsonRPCCmd.Flags().StringVarP(&contractName, "name", "", evmchain.Contract.Name, "evmchain/evmlight contract name")
	evmCmd.AddCommand(jsonRPCCmd)
}
