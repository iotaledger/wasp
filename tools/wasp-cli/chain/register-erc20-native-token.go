package chain

import (
	"context"
	"fmt"

	"github.com/spf13/cobra"

	"github.com/iotaledger/wasp/v2/clients/chainclient"
	"github.com/iotaledger/wasp/v2/clients/iota-go/iotaclient"
	"github.com/iotaledger/wasp/v2/packages/coin"
	"github.com/iotaledger/wasp/v2/packages/vm/core/evm"
	"github.com/iotaledger/wasp/v2/packages/vm/core/evm/iscmagic"
	"github.com/iotaledger/wasp/v2/tools/wasp-cli/cli/cliclients"
	"github.com/iotaledger/wasp/v2/tools/wasp-cli/log"
	"github.com/iotaledger/wasp/v2/tools/wasp-cli/waspcmd"
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
		RunE: func(cmd *cobra.Command, args []string) error {
			var err error
			node, err = waspcmd.DefaultWaspNodeFallback(node)
			if err != nil {
				return err
			}
			chainAliasName = defaultChainFallback(chainAliasName)
			ctx := context.Background()
			client := cliclients.WaspClientWithVersionCheck(ctx, node)

			coinType, err := coin.TypeFromString(args[0])
			if err != nil {
				return fmt.Errorf("invalid coin type: %s => %v", coinType, err)
			}

			request := evm.FuncRegisterERC20Coin.Message(coinType)
			postRequest(ctx, client, chainAliasName, request, chainclient.PostRequestParams{
				GasBudget: iotaclient.DefaultGasBudget,
			}, withOffLedger)

			log.Printf("ERC20 contract deployed at address %s", iscmagic.ERC20CoinAddress(coinType))
			return nil
		},
	}

	waspcmd.WithWaspNodeFlag(cmd, &node)
	withChainFlag(cmd, &chainAliasName)
	return cmd
}
