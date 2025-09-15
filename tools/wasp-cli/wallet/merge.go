package wallet

import (
	"context"
	"time"

	"github.com/spf13/cobra"

	"github.com/iotaledger/wasp/v2/tools/wasp-cli/util"
)

func initMergeCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "merge",
		Short: "Tries to merge all coin objects",
		Args:  cobra.ExactArgs(0),
		RunE: func(cmd *cobra.Command, args []string) error {
			ctx, cancel := context.WithTimeout(context.Background(), time.Second*1000)
			defer cancel()

			return util.TryMergeAllCoins(ctx)
		},
	}
}
