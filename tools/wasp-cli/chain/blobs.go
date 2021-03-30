package chain

import (
	"fmt"
	"os"

	"github.com/iotaledger/goshimmer/packages/ledgerstate"
	"github.com/iotaledger/wasp/packages/hashing"
	"github.com/iotaledger/wasp/packages/kv/codec"
	"github.com/iotaledger/wasp/packages/kv/dict"
	"github.com/iotaledger/wasp/packages/vm/core/blob"
	"github.com/iotaledger/wasp/tools/wasp-cli/config"
	"github.com/iotaledger/wasp/tools/wasp-cli/log"
	"github.com/iotaledger/wasp/tools/wasp-cli/util"
)

func storeBlobCmd(args []string) {
	if len(args) == 0 {
		log.Fatal("Usage: %s chain store-blob [type field type value ...]", os.Args[0])
	}
	uploadBlob(util.EncodeParams(args), false)
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

func showBlobCmd(args []string) {
	if len(args) != 1 {
		log.Fatal("Usage: %s chain show-blob <hash>", os.Args[0])
	}
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
}

func listBlobsCmd(args []string) {
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
}
