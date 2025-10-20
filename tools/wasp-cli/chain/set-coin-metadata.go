package chain

import (
	"context"
	"fmt"

	"github.com/spf13/cobra"

	"github.com/iotaledger/wasp/v2/clients/chainclient"
	"github.com/iotaledger/wasp/v2/clients/iota-go/iotaclient"
	"github.com/iotaledger/wasp/v2/packages/coin"
	"github.com/iotaledger/wasp/v2/packages/parameters"
	"github.com/iotaledger/wasp/v2/packages/vm/core/accounts"
	"github.com/iotaledger/wasp/v2/tools/wasp-cli/cli/cliclients"
	"github.com/iotaledger/wasp/v2/tools/wasp-cli/waspcmd"
)

func initSetCoinMetadataCmd() *cobra.Command {
	var (
		node           string
		chainAliasName string
		withOffLedger  bool
	)

	cmd := &cobra.Command{
		Use:   "set-coin-metadata <coinType>",
		Short: "Registers coin metadata",
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

			coinInfo, err := cliclients.L1Client().GetCoinMetadata(ctx, args[0])
			if err != nil {
				return err
			}

			totalSupply, err := cliclients.L1Client().GetTotalSupply(ctx, args[0])
			if err != nil {
				return err
			}

			request := accounts.SetCoinMetadata.Message(&parameters.IotaCoinInfo{
				CoinType:    coinType,
				Name:        coinInfo.Name,
				Symbol:      coinInfo.Symbol,
				Description: coinInfo.Description,
				IconURL:     coinInfo.IconUrl,
				Decimals:    coinInfo.Decimals,
				TotalSupply: coin.Value(totalSupply.Value.Uint64()),
			})

			postRequest(ctx, client, chainAliasName, request, chainclient.PostRequestParams{
				GasBudget: iotaclient.DefaultGasBudget,
			}, withOffLedger)
			return nil
		},
	}

	waspcmd.WithWaspNodeFlag(cmd, &node)
	withChainFlag(cmd, &chainAliasName)
	return cmd
}
