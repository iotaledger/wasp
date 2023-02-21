package chain

import (
	"strconv"

	"github.com/iotaledger/wasp/clients/chainclient"
	"github.com/iotaledger/wasp/packages/vm/core/evm"
	"github.com/iotaledger/wasp/tools/wasp-cli/util"
	"github.com/iotaledger/wasp/tools/wasp-cli/waspcmd"
	"github.com/spf13/cobra"
)

func initRegisterERC20NativeTokenCmd() *cobra.Command {
	var (
		foundrySerialNumber  uint32
		tokenName            string
		tickerSymbol         string
		tokenDecimals        uint8
		allowance            []string
		transfer             []string
		adjustStorageDeposit bool
		chain                string
		node                 string
	)

	cmd := &cobra.Command{
		Use:   "register-erc20-native-token",
		Short: "Call evm core contract registerERC20Nativetoken entry point",
		Args:  cobra.NoArgs,
		Run: func(cmd *cobra.Command, args []string) {
			node = waspcmd.DefaultWaspNodeFallback(node)
			chain = defaultChainFallback(chain)

			allowanceTokens := util.ParseFungibleTokens(allowance)
			argsAsDict := util.EncodeParams([]string{
				"string", "fs", "uint32", strconv.FormatUint(uint64(foundrySerialNumber), 10),
				"string", "n", "string", tokenName,
				"string", "t", "string", tickerSymbol,
				"string", "d", "uint8", strconv.FormatUint(uint64(tokenDecimals), 10),
			})
			params := chainclient.PostRequestParams{
				Args:      argsAsDict,
				Transfer:  util.ParseFungibleTokens(transfer),
				Allowance: allowanceTokens,
			}
			postRequest(
				node,
				chain,
				evm.Contract.Name,
				evm.FuncRegisterERC20NativeToken.Name,
				params,
				true,
				adjustStorageDeposit,
			)
		},
	}
	waspcmd.WithWaspNodeFlag(cmd, &node)
	withChainFlag(cmd, &chain)

	cmd.Flags().Uint32Var(&foundrySerialNumber, "foundry-sn", 0, "Foundry serial number")
	cmd.Flags().StringVar(&tokenName, "token-name", "", "Token name")
	cmd.Flags().StringVar(&tickerSymbol, "ticker-symbol", "", "Ticker symbol")
	cmd.Flags().Uint8Var(&tokenDecimals, "token-decimals", 0, "Token decimals")
	cmd.Flags().StringSliceVarP(&allowance, "allowance", "l", []string{},
		"include allowance as part of the transaction. Format: <token-id>:<amount>,<token-id>:amount...")
	cmd.Flags().BoolVarP(&adjustStorageDeposit, "adjust-storage-deposit", "s", false, "adjusts the amount of base tokens sent, if it's lower than the min storage deposit required")
	cmd.Flags().StringSliceVarP(&transfer, "transfer", "t", []string{},
		"include a funds transfer as part of the transaction. Format: <token-id>:<amount>,<token-id>:amount...",
	)

	return cmd
}
