package chain

import (
	"os"

	"github.com/iotaledger/wasp/client/chainclient"
	"github.com/iotaledger/wasp/packages/coretypes"
	"github.com/iotaledger/wasp/packages/sctransaction"
	"github.com/iotaledger/wasp/tools/wasp-cli/log"
	"github.com/iotaledger/wasp/tools/wasp-cli/util"
)

func postRequestCmd(args []string) {
	if len(args) < 2 {
		log.Fatal("Usage: %s chain post-request <name> <funcname> [params]", os.Args[0])
	}
	util.WithSCTransaction(func() (*sctransaction.Transaction, error) {
		return SCClient(coretypes.Hn(args[0])).PostRequest(
			args[1],
			chainclient.PostRequestParams{
				Args: util.EncodeParams(args[2:]),
			},
		)
	})
}
