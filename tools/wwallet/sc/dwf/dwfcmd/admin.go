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
	case "deploy":
		check(dwf.Config.Deploy(wallet.Load().SignatureScheme()))

	default:
		adminUsage()
	}
}

func adminUsage() {
	fmt.Printf("Usage: %s dwf admin [deploy]\n", os.Args[0])
	os.Exit(1)
}
