package trcmd

import (
	"fmt"
	"os"

	"github.com/iotaledger/wasp/tools/wwallet/sc/tr"
	"github.com/iotaledger/wasp/tools/wwallet/wallet"
)

func adminCmd(args []string) {
	if len(args) < 1 {
		adminUsage()
	}

	switch args[0] {
	case "deploy":
		check(tr.Config.Deploy(wallet.Load().SignatureScheme()))

	default:
		adminUsage()
	}
}

func adminUsage() {
	fmt.Printf("Usage: %s tr admin [deploy]\n", os.Args[0])
	os.Exit(1)
}
