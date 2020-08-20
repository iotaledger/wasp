package fairroulette

import (
	"fmt"
	"os"
	"strconv"

	"github.com/iotaledger/goshimmer/dapps/valuetransfers/packages/transaction"
	"github.com/iotaledger/wasp/tools/wasp-client/config/fr"
	"github.com/iotaledger/wasp/tools/wasp-client/util"
	"github.com/iotaledger/wasp/tools/wasp-client/wallet"
)

func adminCmd(args []string) {
	if len(args) < 1 {
		adminUsage()
	}

	switch args[0] {
	case "init":
		check(fr.Config.InitSC(wallet.Load().SignatureScheme()))

	case "set-period":
		if len(args) != 2 {
			fr.Config.PrintUsage("admin set-period <seconds>")
			os.Exit(1)
		}
		s, err := strconv.Atoi(args[1])
		check(err)

		util.WithTransaction(func() (*transaction.Transaction, error) {
			tx, err := fr.Client().SetPeriod(s)
			if err != nil {
				return nil, err
			}
			return tx.Transaction, nil
		})

	default:
		adminUsage()
	}
}

func adminUsage() {
	fmt.Printf("Usage: %s fr admin [init|set-period <seconds>]\n", os.Args[0])
	os.Exit(1)
}
