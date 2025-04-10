package chain

import (
	"context"

	"github.com/spf13/cobra"

	"github.com/iotaledger/wasp/clients/chainclient"
	"github.com/iotaledger/wasp/clients/iota-go/iotaclient"
	"github.com/iotaledger/wasp/packages/coin"
	"github.com/iotaledger/wasp/packages/parameters"
	"github.com/iotaledger/wasp/packages/vm/core/accounts"
	"github.com/iotaledger/wasp/tools/wasp-cli/cli/cliclients"
	"github.com/iotaledger/wasp/tools/wasp-cli/log"
	"github.com/iotaledger/wasp/tools/wasp-cli/waspcmd"
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
		Run: func(cmd *cobra.Command, args []string) {
			node = waspcmd.DefaultWaspNodeFallback(node)
			chainAliasName = defaultChainFallback(chainAliasName)
			ctx := context.Background()
			client := cliclients.WaspClientWithVersionCheck(ctx, node)

			coinType, err := coin.TypeFromString(args[0])
			if err != nil {
				log.Fatalf("invalid coin type: %s => %v", coinType, err)
			}

			coinInfo, err := cliclients.L1Client().GetCoinMetadata(ctx, args[0])
			log.Check(err)

			totalSupply, err := cliclients.L1Client().GetTotalSupply(ctx, args[0])
			log.Check(err)

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
		},
	}

	waspcmd.WithWaspNodeFlag(cmd, &node)
	withChainFlag(cmd, &chainAliasName)
	return cmd
}
