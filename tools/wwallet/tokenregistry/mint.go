package tokenregistry

import (
	"fmt"
	"os"
	"strconv"

	"github.com/iotaledger/wasp/packages/sctransaction"
	"github.com/iotaledger/wasp/packages/vm/examples/tokenregistry/trclient"
	"github.com/iotaledger/wasp/tools/wwallet/config/tr"
	"github.com/iotaledger/wasp/tools/wwallet/util"
	"github.com/iotaledger/wasp/tools/wwallet/wallet"
)

func mintCmd(args []string) {
	if len(args) != 2 {
		fmt.Printf("Usage: %s tr mint <description> <amount>\n", os.Args[0])
		os.Exit(1)
	}

	description := args[0]

	amount, err := strconv.Atoi(args[1])
	check(err)

	tx := util.WithSCRequest(tr.Config, func() (*sctransaction.Transaction, error) {
		return tr.Client().MintAndRegister(trclient.MintAndRegisterParams{
			Supply:      int64(amount),
			MintTarget:  wallet.Load().Address(),
			Description: description,
		})
	})

	fmt.Printf("Minted %d tokens of color %s\n", amount, tx.ID())
}
