package wallet

import (
	"fmt"
	"os"
	"strconv"

	"github.com/iotaledger/wasp/packages/apilib"
	"github.com/iotaledger/wasp/tools/wwallet/config"
	"github.com/iotaledger/wasp/tools/wwallet/util"
)

func mintCmd(args []string) {
	if len(args) < 1 {
		fmt.Printf("Usage: %s wallet mint <amount>\n", os.Args[0])
		os.Exit(1)
	}

	wallet := Load()

	amount, err := strconv.Atoi(args[0])
	check(err)

	tx, err := apilib.NewColoredTokensTransaction(config.GoshimmerClient(), wallet.SignatureScheme(), int64(amount))
	check(err)

	util.PostTransaction(tx)

	fmt.Printf("Minted %d tokens of color %s\n", amount, tx.ID())
}
