package chain

import (
	"github.com/spf13/cobra"

	"github.com/iotaledger/wasp/clients/chainclient"
	"github.com/iotaledger/wasp/packages/vm/core/evm"
	"github.com/iotaledger/wasp/tools/wasp-cli/util"
	"github.com/iotaledger/wasp/tools/wasp-cli/waspcmd"
)

func initRegisterERC20NativeTokenCmd() *cobra.Command {
	return buildPostRequestCmd(
		"register-erc20-native-token",
		"Call evm core contract registerERC20NativeToken entry point",
		evm.Contract.Name,
		evm.FuncRegisterERC20NativeToken.Name,
		func(cmd *cobra.Command) {
			initRegisterERC20NativeTokenParams(cmd)
		},
		func(cmd *cobra.Command) []string {
			return getRegisterERC20NativeTokenArgs(cmd)
		},
	)
}

func initRegisterERC20NativeTokenOnRemoteChainCmd() *cobra.Command {
	return buildPostRequestCmd(
		"register-erc20-native-token-on-remote-chain",
		"Call evm core contract registerERC20NativeTokenOnRemoteChain entry point",
		evm.Contract.Name,
		evm.FuncRegisterERC20NativeTokenOnRemoteChain.Name,
		func(cmd *cobra.Command) {
			initRegisterERC20NativeTokenParams(cmd)
			cmd.Flags().Uint8P("target", "A", 0, "Target chain ID")
		},
		func(cmd *cobra.Command) []string {
			extraArgs := []string{"string", "A", "uint8", cmd.Flag("target").Value.String()}
			return append(getRegisterERC20NativeTokenArgs(cmd), extraArgs...)
		},
	)
}

func buildPostRequestCmd(name, desc, hname, fname string, initFlags func(cmd *cobra.Command), funcArgs func(cmd *cobra.Command) []string) *cobra.Command {
	var (
		chain             string
		node              string
		postrequestParams postRequestParams
	)

	cmd := &cobra.Command{
		Use:   name,
		Short: desc,
		Args:  cobra.NoArgs,
		Run: func(cmd *cobra.Command, args []string) {
			node := waspcmd.DefaultWaspNodeFallback(node)
			chain := defaultChainFallback(chain)

			allowanceTokens := util.ParseFungibleTokens(postrequestParams.allowance)

			params := chainclient.PostRequestParams{
				Args:      util.EncodeParams(funcArgs(cmd)),
				Transfer:  util.ParseFungibleTokens(postrequestParams.transfer),
				Allowance: allowanceTokens,
			}
			postRequest(
				node,
				chain,
				hname,
				fname,
				params,
				postrequestParams.offLedger,
				postrequestParams.adjustStorageDeposit,
			)
		},
	}
	waspcmd.WithWaspNodeFlag(cmd, &node)
	withChainFlag(cmd, &chain)
	postrequestParams.initFlags(cmd)
	initFlags(cmd)

	return cmd
}

func initRegisterERC20NativeTokenParams(cmd *cobra.Command) {
	cmd.Flags().Uint32("foundry-sn", 0, "Foundry serial number")
	cmd.Flags().String("token-name", "", "Token name")
	cmd.Flags().String("ticker-symbol", "", "Ticker symbol")
	cmd.Flags().Uint8("token-decimals", 0, "Token decimals")
}

func getRegisterERC20NativeTokenArgs(cmd *cobra.Command) []string {
	return []string{
		"string", "fs", "uint32", flagValString(cmd, "foundry-sn"),
		"string", "n", "string", flagValString(cmd, "token-name"),
		"string", "t", "string", flagValString(cmd, "ticker-symbol"),
		"string", "d", "uint8", flagValString(cmd, "token-decimals"),
	}
}

func flagValString(cmd *cobra.Command, cmdName string) string {
	return cmd.Flag(cmdName).Value.String()
}
