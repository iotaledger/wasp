package wallet

import (
	"context"
	"time"

	"github.com/iotaledger/wasp/tools/wasp-cli/util"
	"github.com/spf13/cobra"
)

func initMergeCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "merge",
		Short: "Tries to merge all coin objects",
		Args:  cobra.ExactArgs(0),
		Run: func(cmd *cobra.Command, args []string) {
			ctx, cancel := context.WithTimeout(context.Background(), time.Second*1000)
			defer cancel()

			util.TryMergeAllCoins(ctx)
		},
	}
}
