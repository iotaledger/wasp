package chain

import (
	"github.com/iotaledger/wasp/packages/evm/jsonrpc"
	"github.com/iotaledger/wasp/packages/vm/core/evm"
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

		initJSONRPCCommand(evmCmd)
	})
}

func initJSONRPCCommand(evmCmd *cobra.Command) {
	var jsonRPCServer evmcli.JSONRPCServer
	var chainID int

	jsonRPCCmd := &cobra.Command{
		Args:  cobra.NoArgs,
		Use:   "jsonrpc",
		Short: "Start a JSON-RPC service to interact with an Ethereum blockchain running on ISC",
		Long: `Start a JSON-RPC service to interact with an Ethereum blockchain running on ISC.

By default the server has no unlocked accounts. To send transactions, either:

- use eth_sendRawTransaction
- configure an unlocked account with --account, and use eth_sendTransaction`,
		Run: func(cmd *cobra.Command, args []string) {
			backend := jsonrpc.NewWaspClientBackend(Client())
			jsonRPCServer.ServeJSONRPC(backend, chainID)
		},
	}

	jsonRPCServer.InitFlags(jsonRPCCmd)
	jsonRPCCmd.Flags().IntVarP(&chainID, "evm-chainid", "", evm.DefaultChainID, "ChainID (used for signing transactions)")
	evmCmd.AddCommand(jsonRPCCmd)
}
