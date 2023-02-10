package chain

import (
	"context"
	"fmt"

	"github.com/spf13/cobra"

	iotago "github.com/iotaledger/iota.go/v3"
	"github.com/iotaledger/wasp/clients/apiclient"
	"github.com/iotaledger/wasp/packages/hashing"
	"github.com/iotaledger/wasp/packages/isc"
	"github.com/iotaledger/wasp/packages/kv"
	"github.com/iotaledger/wasp/packages/kv/dict"
	"github.com/iotaledger/wasp/tools/wasp-cli/cli/cliclients"
	"github.com/iotaledger/wasp/tools/wasp-cli/cli/config"
	"github.com/iotaledger/wasp/tools/wasp-cli/log"
	"github.com/iotaledger/wasp/tools/wasp-cli/util"
	"github.com/iotaledger/wasp/tools/wasp-cli/waspcmd"
)

var uploadQuorum int

func initUploadFlags(chainCmd *cobra.Command) {
	chainCmd.PersistentFlags().IntVarP(&uploadQuorum, "upload-quorum", "", 3, "quorum for blob upload")
}

func initStoreBlobCmd() *cobra.Command {
	var node string
	var chain string
	cmd := &cobra.Command{
		Use:   "store-blob <type> <field> <type> <value> ...",
		Short: "Store a blob in the chain",
		Args:  cobra.MinimumNArgs(4),
		Run: func(cmd *cobra.Command, args []string) {
			node = waspcmd.DefaultWaspNodeFallback(node)
			chain = defaultChainFallback(chain)

			chainID := config.GetChain(chain)
			uploadBlob(cliclients.WaspClient(node), chainID, util.EncodeParams(args))
		},
	}
	waspcmd.WithWaspNodeFlag(cmd, &node)
	withChainFlag(cmd, &chain)
	return cmd
}

func uploadBlob(client *apiclient.APIClient, chainID isc.ChainID, fieldValues dict.Dict) (hash hashing.HashValue) {
	chainClient := cliclients.ChainClient(client, chainID)

	hash, _, _, err := chainClient.UploadBlob(context.Background(), fieldValues)
	log.Check(err)
	log.Printf("uploaded blob to chain -- hash: %s", hash)
	// TODO print receipt?
	return hash
}

func initShowBlobCmd() *cobra.Command {
	var node string
	var chain string
	cmd := &cobra.Command{
		Use:   "show-blob <hash>",
		Short: "Show a blob in chain",
		Args:  cobra.ExactArgs(1),
		Run: func(cmd *cobra.Command, args []string) {
			node = waspcmd.DefaultWaspNodeFallback(node)
			chain = defaultChainFallback(chain)

			hash, err := hashing.HashValueFromHex(args[0])
			log.Check(err)

			client := cliclients.WaspClient(node)

			blobInfo, _, err := client.
				CorecontractsApi.
				BlobsGetBlobInfo(context.Background(), config.GetChain(chain).String(), hash.Hex()).
				Execute()
			log.Check(err)

			values := dict.New()
			for field := range blobInfo.Fields {
				value, _, err := client.
					CorecontractsApi.
					BlobsGetBlobValue(context.Background(), config.GetChain(chain).String(), hash.Hex(), field).
					Execute()

				log.Check(err)

				decodedValue, err := iotago.DecodeHex(value.ValueData)
				log.Check(err)

				values.Set(kv.Key(field), []byte(decodedValue))
			}
			util.PrintDictAsJSON(values)
		},
	}
	waspcmd.WithWaspNodeFlag(cmd, &node)
	withChainFlag(cmd, &chain)
	return cmd
}

func initListBlobsCmd() *cobra.Command {
	var node string
	var chain string
	cmd := &cobra.Command{
		Use:   "list-blobs",
		Short: "List blobs in chain",
		Args:  cobra.NoArgs,
		Run: func(cmd *cobra.Command, args []string) {
			node = waspcmd.DefaultWaspNodeFallback(node)
			chain = defaultChainFallback(chain)
			client := cliclients.WaspClient(node)

			blobsResponse, _, err := client.
				CorecontractsApi.
				BlobsGetAllBlobs(context.Background(), config.GetChain(chain).String()).
				Execute()

			log.Check(err)

			log.Printf("Total %d blob(s) in chain %s\n", len(blobsResponse.Blobs), config.GetChain(chain))

			header := []string{"hash", "size"}
			rows := make([][]string, len(blobsResponse.Blobs))

			for i, blob := range blobsResponse.Blobs {
				rows[i] = []string{blob.Hash, fmt.Sprintf("%d", blob.Size)}
			}

			log.PrintTable(header, rows)
		},
	}
	waspcmd.WithWaspNodeFlag(cmd, &node)
	withChainFlag(cmd, &chain)
	return cmd
}
