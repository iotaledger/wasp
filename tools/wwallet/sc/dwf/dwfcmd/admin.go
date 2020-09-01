package dwfcmd

import (
	"fmt"
	"os"

	"github.com/iotaledger/wasp/tools/wwallet/sc/dwf"
	"github.com/iotaledger/wasp/tools/wwallet/wallet"
)

func adminCmd(args []string) {
	if len(args) < 1 {
		adminUsage()
	}

	switch args[0] {
	case "init":
		check(dwf.Config.InitSC(wallet.Load().SignatureScheme()))

	default:
		adminUsage()
	}
}

func adminUsage() {
	fmt.Printf("Usage: %s dwf admin [init]\n", os.Args[0])
	os.Exit(1)
}
