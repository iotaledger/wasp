package inspection

import (
	"context"

	"github.com/spf13/cobra"

	"github.com/iotaledger/wasp/v2/clients/iota-go/iotago"
	"github.com/iotaledger/wasp/v2/tools/wasp-cli/cli/cliclients"
	"github.com/iotaledger/wasp/v2/tools/wasp-cli/log"
)

func initAssetsBagCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "assetsbag <AssetsBagID>",
		Short: "Show the content of an AssetsBag",
		Args:  cobra.ExactArgs(1),
		Run: func(cmd *cobra.Command, args []string) {
			bagID, err := iotago.ObjectIDFromHex(args[0])
			log.Check(err)

			ctx := context.Background()
			assetsBag, err := cliclients.L2Client().GetAssetsBagWithBalances(ctx, bagID)
			log.Check(err)

			log.Printf("AssetsBag:\n	ID: %s\n	Size: %d\n\n", assetsBag.ID, assetsBag.Size)
			log.Printf("Balances:\n")

			for t, a := range assetsBag.Coins.Iterate() {
				log.Printf("\t%s: %v\n", t, a)
			}

			for id, t := range assetsBag.Objects.Iterate() {
				log.Printf("\t%s: %v\n", id, t)
			}
		},
	}
}
