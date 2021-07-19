package blob

import (
	"io/ioutil"
	"os"

	"github.com/iotaledger/wasp/packages/hashing"
	"github.com/iotaledger/wasp/tools/wasp-cli/config"
	"github.com/iotaledger/wasp/tools/wasp-cli/log"
	"github.com/spf13/cobra"
)

func Init(rootCmd *cobra.Command) {
	rootCmd.AddCommand(blobCmd)

	blobCmd.AddCommand(putBlobCmd)
	blobCmd.AddCommand(getBlobCmd)
	blobCmd.AddCommand(hasBlobCmd)
}

var blobCmd = &cobra.Command{
	Use:   "blob <command>",
	Short: "Interact with the blob cache",
	Args:  cobra.NoArgs,
	Run: func(cmd *cobra.Command, args []string) {
		log.Check(cmd.Help())
	},
}

var putBlobCmd = &cobra.Command{
	Use:   "put <filename>",
	Short: "Store a file in the blob cache",
	Args:  cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		data, err := ioutil.ReadFile(args[0])
		log.Check(err)
		hash, err := config.WaspClient().PutBlob(data)
		log.Check(err)
		log.Printf("Blob uploaded. Hash: %s\n", hash)
	},
}

var getBlobCmd = &cobra.Command{
	Use:   "get <hash>",
	Short: "Get a blob from the blob cache",
	Args:  cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		hash, err := hashing.HashValueFromBase58(args[0])
		log.Check(err)
		data, err := config.WaspClient().GetBlob(hash)
		log.Check(err)
		_, err = os.Stdout.Write(data)
		log.Check(err)
	},
}

var hasBlobCmd = &cobra.Command{
	Use:   "has <hash>",
	Short: "Determine if a blob is in the blob cache",
	Args:  cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		hash, err := hashing.HashValueFromBase58(args[0])
		log.Check(err)
		ok, err := config.WaspClient().HasBlob(hash)
		log.Check(err)
		log.Printf("%v", ok)
	},
}
