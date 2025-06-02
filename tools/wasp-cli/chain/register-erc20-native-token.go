package chain

import (
	"context"

	"github.com/spf13/cobra"

	"github.com/iotaledger/wasp/clients/chainclient"
	"github.com/iotaledger/wasp/clients/iota-go/iotaclient"
	"github.com/iotaledger/wasp/packages/coin"
	"github.com/iotaledger/wasp/packages/vm/core/evm"
	"github.com/iotaledger/wasp/packages/vm/core/evm/iscmagic"
	"github.com/iotaledger/wasp/tools/wasp-cli/cli/cliclients"
	"github.com/iotaledger/wasp/tools/wasp-cli/log"
	"github.com/iotaledger/wasp/tools/wasp-cli/waspcmd"
)

func initRegisterERC20NativeTokenCmd() *cobra.Command {
	var (
		node           string
		chainAliasName string
		withOffLedger  bool
	)

	cmd := &cobra.Command{
		Use:   "register-erc20-native-token <coinType>",
		Short: "Call evm core contract registerERC20NativeToken entry point",
		Args:  cobra.ExactArgs(1),
		Run: func(cmd *cobra.Command, args []string) {
			node = waspcmd.DefaultWaspNodeFallback(node)
			chainAliasName = defaultChainFallback(chainAliasName)
			ctx := context.Background()
			client := cliclients.WaspClientWithVersionCheck(ctx, node)

			coinType, err := coin.TypeFromString(args[0])
			if err != nil {
				log.Fatalf("invalid coin type: %s => %v", coinType, err)
			}

			request := evm.FuncRegisterERC20Coin.Message(coinType)
			postRequest(ctx, client, chainAliasName, request, chainclient.PostRequestParams{
				GasBudget: iotaclient.DefaultGasBudget,
			}, withOffLedger)

			log.Printf("ERC20 contract deployed at address %s", iscmagic.ERC20CoinAddress(coinType))
		},
	}

	waspcmd.WithWaspNodeFlag(cmd, &node)
	withChainFlag(cmd, &chainAliasName)
	return cmd
}
