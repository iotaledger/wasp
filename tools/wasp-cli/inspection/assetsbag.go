package inspection

import (
	"context"
	"fmt"

	"github.com/spf13/cobra"

	"github.com/iotaledger/wasp/clients/iota-go/iotago"
	"github.com/iotaledger/wasp/tools/wasp-cli/cli/cliclients"
	"github.com/iotaledger/wasp/tools/wasp-cli/log"
)

func initAssetsBagCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "assetsbag <AssetsBagID>",
		Short: "Show the content of an AssetsBag",
		Args:  cobra.ExactArgs(1),
		Run: func(cmd *cobra.Command, args []string) {
			bagId, err := iotago.ObjectIDFromHex(args[0])
			log.Check(err)

			ctx := context.Background()
			assetsBag, err := cliclients.L2Client().GetAssetsBagWithBalances(ctx, bagId)
			log.Check(err)

			fmt.Printf("AssetsBag:\n	ID: %s\n	Size: %d\n\n", assetsBag.ID, assetsBag.Size)
			fmt.Println("Balances:")

			for n, c := range assetsBag.Coins {
				fmt.Printf("	%s: %v\n", n, c)
			}

			for n, c := range assetsBag.Objects {
				fmt.Printf("	%s: %v\n", n, c)
			}
		},
	}
}
