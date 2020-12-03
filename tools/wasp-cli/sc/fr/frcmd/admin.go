package frcmd

import (
	"fmt"
	"os"
	"strconv"

	"github.com/iotaledger/wasp/tools/wasp-cli/sc/fr"
	"github.com/iotaledger/wasp/tools/wasp-cli/wallet"
)

func adminCmd(args []string) {
	if len(args) < 1 {
		adminUsage()
	}

	switch args[0] {
	case "deploy":
		check(fr.Config.Deploy(wallet.Load().SignatureScheme()))

	case "set-period":
		if len(args) != 2 {
			fr.Config.PrintUsage("admin set-period <seconds>")
			os.Exit(1)
		}
		s, err := strconv.Atoi(args[1])
		check(err)

		_, err = fr.Client().SetPeriod(s)
		check(err)

	default:
		adminUsage()
	}
}

func adminUsage() {
	fmt.Printf("Usage: %s fr admin [deploy|set-period <seconds>]\n", os.Args[0])
	os.Exit(1)
}
