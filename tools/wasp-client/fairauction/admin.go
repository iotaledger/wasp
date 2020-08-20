package fairauction

import (
	"fmt"
	"os"
	"strconv"

	"github.com/iotaledger/goshimmer/dapps/valuetransfers/packages/transaction"
	"github.com/iotaledger/wasp/tools/wasp-client/config/fa"
	"github.com/iotaledger/wasp/tools/wasp-client/util"
	"github.com/iotaledger/wasp/tools/wasp-client/wallet"
)

func adminCmd(args []string) {
	if len(args) < 1 {
		adminUsage()
	}

	switch args[0] {
	case "init":
		check(fa.Config.InitSC(wallet.Load().SignatureScheme()))

	case "set-owner-margin":
		if len(args) != 2 {
			fa.Config.PrintUsage("admin set-owner-margin <promilles>")
			os.Exit(1)
		}
		p, err := strconv.Atoi(args[1])
		check(err)
		util.WithTransaction(func() (*transaction.Transaction, error) {
			tx, err := fa.Client().SetOwnerMargin(int64(p))
			return tx.Transaction, err
		})

	default:
		adminUsage()
	}
}

func adminUsage() {
	fmt.Printf("Usage: %s fr admin [init|set-owner-margin <promilles>]\n", os.Args[0])
	os.Exit(1)
}
