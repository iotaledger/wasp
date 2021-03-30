package chain

import (
	"os"
	"strconv"
	"strings"

	"github.com/iotaledger/goshimmer/packages/ledgerstate"
	"github.com/iotaledger/wasp/client/chainclient"
	"github.com/iotaledger/wasp/packages/coretypes"
	"github.com/iotaledger/wasp/packages/coretypes/requestargs"
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

	cb := make(map[ledgerstate.Color]uint64)
	for _, tr := range transfer {
		parts := strings.Split(tr, ":")
		if len(parts) != 2 {
			log.Fatal("Syntax for --transfer: <color>:<amount>,<color:amount>...\nExample: IOTA:100")
		}
		color := colorFromString(parts[0])
		amount, err := strconv.Atoi(parts[1])
		log.Check(err)
		cb[color] = uint64(amount)
	}

	util.WithSCTransaction(GetCurrentChainID(), func() (*ledgerstate.Transaction, error) {
		return SCClient(coretypes.Hn(args[0])).PostRequest(
			args[1],
			chainclient.PostRequestParams{
				Args:     requestargs.New().AddEncodeSimpleMany(util.EncodeParams(args[2:])),
				Transfer: ledgerstate.NewColoredBalances(cb),
			},
		)
	})
}

func colorFromString(s string) ledgerstate.Color {
	if s == ledgerstate.ColorIOTA.String() {
		return ledgerstate.ColorIOTA
	}
	c, err := ledgerstate.ColorFromBase58EncodedString(s)
	log.Check(err)
	return c
}
