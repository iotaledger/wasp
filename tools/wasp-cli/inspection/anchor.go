package inspection

import (
	"context"

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

			log.Printf("Anchor:\n")
			log.Printf("\tID: %s\n", anchor.ObjectID)
			log.Printf("\tAssetsBag: %s\n", anchor.Object.Assets.Value.ID)
			log.Printf("\tStateIndex: %d\n", anchor.Object.StateIndex)
			log.Printf("\tStateMetadata: %s\n", hexutil.Encode(anchor.Object.StateMetadata))

			log.Printf("\nStateMetadata decoded:\n")
			metadata, err := transaction.StateMetadataFromBytes(anchor.Object.StateMetadata)
			if err != nil {
				log.Fatalf("\tCould not decode state metadata: %v\n", err)
			}

			log.Printf("\tGasCoinObjectID: %s\n", metadata.GasCoinObjectID)
			log.Printf("\tInitDeposit: %d\n", metadata.InitDeposit)
			log.Printf("\tL1Commitment: BlockHash:%s, TrieRoot:%s\n",
				metadata.L1Commitment.BlockHash().String(),
				metadata.L1Commitment.TrieRoot().String())
			log.Printf("\tPublicUrl: %s\n", metadata.PublicURL)
			log.Printf("\tSchemaVersion: %d\n", metadata.SchemaVersion)

			if anchor.Object.StateIndex != 0 {
				log.Printf("Skipping InitParams, as state index is not 0\n")
				return
			}

			initParams, err := origin.DecodeInitParams(metadata.InitParams)
			if err != nil {
				log.Fatalf("\tCould not decode Init Params! Params: %s\n", metadata.InitParams.String())
			}

			log.Printf("\n\tInitParams:\n")
			log.Printf("\t\tChainOwner: %s\n", initParams.ChainOwner)
			log.Printf("\t\tBlockKeepAmount: %d\n", initParams.BlockKeepAmount)
			log.Printf("\t\tDeployTestContracts: %v\n", initParams.DeployTestContracts)
			log.Printf("\t\tEVM ChainID: %d\n", initParams.EVMChainID)
		},
	}
}
