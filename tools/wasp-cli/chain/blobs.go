package chain

import (
	"fmt"

	"github.com/iotaledger/goshimmer/packages/ledgerstate"
	"github.com/iotaledger/wasp/packages/hashing"
	"github.com/iotaledger/wasp/packages/kv/codec"
	"github.com/iotaledger/wasp/packages/kv/dict"
	"github.com/iotaledger/wasp/packages/vm/core/blob"
	"github.com/iotaledger/wasp/tools/wasp-cli/config"
	"github.com/iotaledger/wasp/tools/wasp-cli/log"
	"github.com/iotaledger/wasp/tools/wasp-cli/util"
	"github.com/spf13/cobra"
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
		uploadBlob(util.EncodeParams(args), false)
	},
}

func uploadBlob(fieldValues dict.Dict, forceWait bool) (hash hashing.HashValue) {
	util.WithSCTransaction(
		GetCurrentChainID(),
		func() (tx *ledgerstate.Transaction, err error) {
			hash, tx, err = Client().UploadBlob(fieldValues, config.CommitteeApi(chainCommittee()), uploadQuorum)
			if err == nil {
				log.Printf("uploaded blob to chain -- hash: %s", hash)
			}
			return
		},
		forceWait)
	return
}

var showBlobCmd = &cobra.Command{
	Use:   "show-blob <hash>",
	Short: "Show a blob in chain",
	Args:  cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		hash := util.ValueFromString("base58", args[0])
		fields, err := SCClient(blob.Interface.Hname()).CallView(blob.FuncGetBlobInfo,
			dict.Dict{
				blob.ParamHash: hash,
			})
		log.Check(err)

		values := dict.New()
		for field := range fields {
			value, err := SCClient(blob.Interface.Hname()).CallView(blob.FuncGetBlobField,
				dict.Dict{
					blob.ParamHash:  hash,
					blob.ParamField: []byte(field),
				})
			log.Check(err)
			values.Set(field, value[blob.ParamBytes])
		}
		util.PrintDictAsJson(values)
	},
}

var listBlobsCmd = &cobra.Command{
	Use:   "list-blobs",
	Short: "List blobs in chain",
	Args:  cobra.NoArgs,
	Run: func(cmd *cobra.Command, args []string) {
		ret, err := SCClient(blob.Interface.Hname()).CallView(blob.FuncListBlobs)
		log.Check(err)

		blobs, err := blob.DecodeSizesMap(ret)
		log.Check(err)

		log.Printf("Total %d blob(s) in chain %s\n", len(ret), GetCurrentChainID())

		header := []string{"hash", "size"}
		rows := make([][]string, len(ret))
		i := 0
		for k, size := range blobs {
			hash, _, err := codec.DecodeHashValue([]byte(k))
			log.Check(err)
			rows[i] = []string{hash.String(), fmt.Sprintf("%d", size)}
			i++
		}
		log.PrintTable(header, rows)
	},
}
