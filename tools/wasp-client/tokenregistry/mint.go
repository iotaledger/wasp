package tokenregistry

import (
	"fmt"
	"os"
	"strconv"

	"github.com/iotaledger/wasp/packages/vm/examples/tokenregistry/trclient"
	"github.com/iotaledger/wasp/tools/wasp-client/config/tr"
	"github.com/iotaledger/wasp/tools/wasp-client/wallet"
)

func mintCmd(args []string) {
	if len(args) != 2 {
		fmt.Printf("Usage: %s tr mint <description> <amount>\n", os.Args[1])
		os.Exit(1)
	}

	description := args[0]

	amount, err := strconv.Atoi(args[1])
	check(err)

	color, err := tr.Client().MintAndRegister(trclient.MintAndRegisterParams{
		Supply:      int64(amount),
		MintTarget:  wallet.Load().Address(),
		Description: description,
	})
	check(err)

	fmt.Printf("Minted %d tokens of color %s\n", amount, color)
}
