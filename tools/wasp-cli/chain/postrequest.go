package chain

import (
	"os"
	"strconv"
	"strings"

	"github.com/iotaledger/goshimmer/dapps/valuetransfers/packages/balance"
	"github.com/iotaledger/wasp/client/chainclient"
	"github.com/iotaledger/wasp/packages/coretypes"
	"github.com/iotaledger/wasp/packages/coretypes/cbalances"
	"github.com/iotaledger/wasp/packages/coretypes/requestargs"
	"github.com/iotaledger/wasp/packages/sctransaction"
	wasputil "github.com/iotaledger/wasp/packages/util"
	"github.com/iotaledger/wasp/tools/wasp-cli/log"
	"github.com/iotaledger/wasp/tools/wasp-cli/util"
	"github.com/spf13/pflag"
)

var transfer []string

func initPostRequestFlags(flags *pflag.FlagSet) {
	flags.StringSliceVarP(&transfer, "transfer", "", []string{},
		"include a funds transfer as part of the transaction. Format: <color>:<amount>,<color>:amount...",
	)
}

func postRequestCmd(args []string) {
	if len(args) < 2 {
		log.Fatal("Usage: %s chain post-request <name> <funcname> [params]", os.Args[0])
	}

	cb := make(map[balance.Color]int64)
	for _, tr := range transfer {
		parts := strings.Split(tr, ":")
		if len(parts) != 2 {
			log.Fatal("Syntax for --transfer: <color>:<amount>,<color:amount>...\nExample: IOTA:100")
		}
		color, err := wasputil.ColorFromString(parts[0])
		log.Check(err)
		amount, err := strconv.Atoi(parts[1])
		log.Check(err)
		cb[color] = int64(amount)
	}

	util.WithSCTransaction(func() (*sctransaction.Transaction, error) {
		return SCClient(coretypes.Hn(args[0])).PostRequest(
			args[1],
			chainclient.PostRequestParams{
				Args:     requestargs.New().AddEncodeSimpleMany(util.EncodeParams(args[2:])),
				Transfer: cbalances.NewFromMap(cb),
			},
		)
	})
}
