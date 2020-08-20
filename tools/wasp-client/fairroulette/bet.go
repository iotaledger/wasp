package fairroulette

import (
	"fmt"
	"os"
	"strconv"

	"github.com/iotaledger/goshimmer/dapps/valuetransfers/packages/transaction"
	"github.com/iotaledger/wasp/tools/wasp-client/config/fr"
	"github.com/iotaledger/wasp/tools/wasp-client/util"
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

	util.WithTransaction(func() (*transaction.Transaction, error) {
		tx, err := fr.Client().Bet(color, amount)
		return tx.Transaction, err
	})
}

func check(err error) {
	if err != nil {
		fmt.Printf("error: %s\n", err)
		os.Exit(1)
	}
}
