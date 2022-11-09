package chain

import (
	"fmt"

	"github.com/spf13/cobra"

	"github.com/iotaledger/wasp/packages/hashing"
	"github.com/iotaledger/wasp/packages/kv/codec"
	"github.com/iotaledger/wasp/packages/kv/dict"
	"github.com/iotaledger/wasp/packages/vm/core/blob"
	"github.com/iotaledger/wasp/tools/wasp-cli/log"
	"github.com/iotaledger/wasp/tools/wasp-cli/util"
)

var uploadQuorum int

func initUploadFlags(chainCmd *cobra.Command) {
	chainCmd.PersistentFlags().IntVarP(&uploadQuorum, "upload-quorum", "", 3, "quorum for blob upload")
}

var storeBlobCmd = &cobra.Command{
	Use:   "store-blob <type> <field> <type> <value> ...",
	Short: "Store a blob in the chain",
	Args:  cobra.MinimumNArgs(4),
	Run: func(cmd *cobra.Command, args []string) {
		uploadBlob(util.EncodeParams(args))
	},
}

func uploadBlob(fieldValues dict.Dict) (hash hashing.HashValue) {
	hash, _, _, err := Client().UploadBlob(fieldValues)
	log.Check(err)
	log.Printf("uploaded blob to chain -- hash: %s", hash)
	// TODO print receipt?
	return hash
}

var showBlobCmd = &cobra.Command{
	Use:   "show-blob <hash>",
	Short: "Show a blob in chain",
	Args:  cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		hash, err := hashing.HashValueFromHex(args[0])
		log.Check(err)
		fields, err := SCClient(blob.Contract.Hname()).CallView(
			blob.ViewGetBlobInfo.Name,
			dict.Dict{blob.ParamHash: hash.Bytes()},
		)
		log.Check(err)

		values := dict.New()
		for field := range fields {
			value, err := SCClient(blob.Contract.Hname()).CallView(
				blob.ViewGetBlobField.Name,
				dict.Dict{
					blob.ParamHash:  hash.Bytes(),
					blob.ParamField: []byte(field),
				},
			)
			log.Check(err)
			values.Set(field, value[blob.ParamBytes])
		}
		util.PrintDictAsJSON(values)
	},
}

var listBlobsCmd = &cobra.Command{
	Use:   "list-blobs",
	Short: "List blobs in chain",
	Args:  cobra.NoArgs,
	Run: func(cmd *cobra.Command, args []string) {
		ret, err := SCClient(blob.Contract.Hname()).CallView(blob.ViewListBlobs.Name, nil)
		log.Check(err)

		blobs, err := blob.DecodeSizesMap(ret)
		log.Check(err)

		log.Printf("Total %d blob(s) in chain %s\n", len(ret), GetCurrentChainID())

		header := []string{"hash", "size"}
		rows := make([][]string, len(ret))
		i := 0
		for k, size := range blobs {
			hash, err := codec.DecodeHashValue([]byte(k))
			log.Check(err)
			rows[i] = []string{hash.String(), fmt.Sprintf("%d", size)}
			i++
		}
		log.PrintTable(header, rows)
	},
}
