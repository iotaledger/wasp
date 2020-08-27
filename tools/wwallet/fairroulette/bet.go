package fairroulette

import (
	"fmt"
	"os"
	"strconv"

	"github.com/iotaledger/wasp/packages/sctransaction"
	"github.com/iotaledger/wasp/tools/wwallet/config/fr"
	"github.com/iotaledger/wasp/tools/wwallet/util"
)

func betCmd(args []string) {
	if len(args) != 2 {
		fr.Config.PrintUsage("bet <color> <amount>")
		os.Exit(1)
	}

	color, err := strconv.Atoi(args[0])
	check(err)
	amount, err := strconv.Atoi(args[1])
	check(err)

	util.WithSCRequest(fr.Config, func() (*sctransaction.Transaction, error) {
		return fr.Client().Bet(color, amount)
	})
}

func check(err error) {
	if err != nil {
		fmt.Printf("error: %s\n", err)
		os.Exit(1)
	}
}
