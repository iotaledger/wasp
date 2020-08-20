package wallet

import (
	"fmt"
	"os"
	"strconv"

	"github.com/iotaledger/wasp/packages/apilib"
	"github.com/iotaledger/wasp/tools/wasp-client/config"
)

func mintCmd(args []string) {
	if len(args) < 1 {
		fmt.Printf("Usage: %s wallet mint <amount>\n", os.Args[1])
		os.Exit(1)
	}

	wallet := Load()

	amount, err := strconv.Atoi(args[0])
	check(err)

	tx, err := apilib.NewColoredTokensTransaction(config.GoshimmerClient(), wallet.SignatureScheme(), int64(amount))
	check(err)

	check(config.GoshimmerClient().PostTransaction(tx))

	fmt.Printf("Minted %d tokens of color %s\n", amount, tx.ID())
}
