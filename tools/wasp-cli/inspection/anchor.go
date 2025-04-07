package inspection

import (
	"context"
	"fmt"

	"github.com/ethereum/go-ethereum/common/hexutil"
	"github.com/spf13/cobra"

	"github.com/iotaledger/wasp/clients/iota-go/iotago"
	"github.com/iotaledger/wasp/packages/origin"
	"github.com/iotaledger/wasp/packages/transaction"
	"github.com/iotaledger/wasp/tools/wasp-cli/cli/cliclients"
	"github.com/iotaledger/wasp/tools/wasp-cli/log"
)

func initAnchorCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "anchor <AnchorID>",
		Short: "Show the content of an Anchor",
		Args:  cobra.ExactArgs(1),
		Run: func(cmd *cobra.Command, args []string) {
			objectID, err := iotago.ObjectIDFromHex(args[0])
			log.Check(err)

			ctx := context.Background()
			anchor, err := cliclients.L2Client().GetAnchorFromObjectID(ctx, objectID)
			log.Check(err)

			fmt.Printf("Anchor:\n")
			fmt.Printf("\tID: %s\n", anchor.ObjectID)
			fmt.Printf("\tAssetsBag: %s\n", anchor.Object.Assets.Value.ID)
			fmt.Printf("\tStateIndex: %d\n", anchor.Object.StateIndex)
			fmt.Printf("\tStateMetadata: %s\n", hexutil.Encode(anchor.Object.StateMetadata))

			fmt.Print("\nStateMetadata decoded:\n")
			metadata, err := transaction.StateMetadataFromBytes(anchor.Object.StateMetadata)
			if err != nil {
				fmt.Printf("\tCould not decode state metadata: %v\n", err)
				return
			}

			fmt.Printf("\tGasCoinObjectID: %s\n", metadata.GasCoinObjectID)
			fmt.Printf("\tInitDeposit: %d\n", metadata.InitDeposit)
			fmt.Printf("\tL1Commitment: BlockHash:%s, TrieRoot:%s\n",
				metadata.L1Commitment.BlockHash().String(),
				metadata.L1Commitment.TrieRoot().String())
			fmt.Printf("\tPublicUrl: %s\n", metadata.PublicURL)
			fmt.Printf("\tSchemaVersion: %d\n", metadata.SchemaVersion)

			if anchor.Object.StateIndex != 0 {
				fmt.Print("Skipping InitParams, as state index is not 0\n")
				return
			}

			initParams, err := origin.DecodeInitParams(metadata.InitParams)
			if err != nil {
				fmt.Printf("\tCould not decode Init Params! Params: %s\n", metadata.InitParams.String())
				return
			}

			fmt.Print("\n\tInitParams:\n")
			fmt.Printf("\t\tChainOwner: %s\n", initParams.ChainOwner)
			fmt.Printf("\t\tBlockKeepAmount: %d\n", initParams.BlockKeepAmount)
			fmt.Printf("\t\tDeployTestContracts: %v\n", initParams.DeployTestContracts)
			fmt.Printf("\t\tEVM ChainID: %d\n", initParams.EVMChainID)
		},
	}
}
