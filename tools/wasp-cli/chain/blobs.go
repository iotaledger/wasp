package chain

import (
	"context"
	"fmt"

	"github.com/spf13/cobra"

	"github.com/iotaledger/wasp/clients/apiclient"
	"github.com/iotaledger/wasp/packages/hashing"
	"github.com/iotaledger/wasp/packages/kv"
	"github.com/iotaledger/wasp/packages/kv/dict"
	"github.com/iotaledger/wasp/tools/wasp-cli/cli/cliclients"
	"github.com/iotaledger/wasp/tools/wasp-cli/cli/config"
	"github.com/iotaledger/wasp/tools/wasp-cli/log"
	"github.com/iotaledger/wasp/tools/wasp-cli/util"
)

var uploadQuorum int

func initUploadFlags(chainCmd *cobra.Command) {
	chainCmd.PersistentFlags().IntVarP(&uploadQuorum, "upload-quorum", "", 3, "quorum for blob upload")
}

func initStoreBlobCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "store-blob <type> <field> <type> <value> ...",
		Short: "Store a blob in the chain",
		Args:  cobra.MinimumNArgs(4),
		Run: func(cmd *cobra.Command, args []string) {
			uploadBlob(cliclients.WaspClientForIndex(), util.EncodeParams(args))
		},
	}
}

func uploadBlob(client *apiclient.APIClient, fieldValues dict.Dict) (hash hashing.HashValue) {
	chainClient := cliclients.ChainClient(client)

	hash, _, _, err := chainClient.UploadBlob(context.Background(), fieldValues)
	log.Check(err)
	log.Printf("uploaded blob to chain -- hash: %s", hash)
	// TODO print receipt?
	return hash
}

func initShowBlobCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "show-blob <hash>",
		Short: "Show a blob in chain",
		Args:  cobra.ExactArgs(1),
		Run: func(cmd *cobra.Command, args []string) {
			hash, err := hashing.HashValueFromHex(args[0])
			log.Check(err)

			client := cliclients.WaspClientForIndex()

			blobInfo, _, err := client.
				CorecontractsApi.
				BlobsGetBlobInfo(context.Background(), config.GetCurrentChainID().String(), hash.Hex()).
				Execute()
			log.Check(err)

			values := dict.New()
			for field := range blobInfo.Fields {
				value, _, err := client.
					CorecontractsApi.
					BlobsGetBlobValue(context.Background(), config.GetCurrentChainID().String(), hash.Hex(), field).
					Execute()

				log.Check(err)
				values.Set(kv.Key(field), []byte(value.ValueData))
			}
			util.PrintDictAsJSON(values)
		},
	}
}

func initListBlobsCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "list-blobs",
		Short: "List blobs in chain",
		Args:  cobra.NoArgs,
		Run: func(cmd *cobra.Command, args []string) {
			client := cliclients.WaspClientForIndex()

			blobsResponse, _, err := client.
				CorecontractsApi.
				BlobsGetAllBlobs(context.Background(), config.GetCurrentChainID().String()).
				Execute()

			log.Check(err)

			log.Printf("Total %d blob(s) in chain %s\n", len(blobsResponse.Blobs), config.GetCurrentChainID())

			header := []string{"hash", "size"}
			rows := make([][]string, len(blobsResponse.Blobs))

			for i, blob := range blobsResponse.Blobs {
				rows[i] = []string{blob.Hash, fmt.Sprintf("%d", blob.Size)}
			}

			log.PrintTable(header, rows)
		},
	}
}
