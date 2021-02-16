// +build ignore

package dwfcmd

import (
	"fmt"
	"os"

	"github.com/iotaledger/wasp/tools/wasp-cli/sc/dwf"
	"github.com/iotaledger/wasp/tools/wasp-cli/wallet"
)

func adminCmd(args []string) {
	if len(args) < 1 {
		adminUsage()
	}

	switch args[0] {
	case "deploy":
		log.Check(dwf.Config.Deploy(wallet.Load().SignatureScheme()))

	default:
		adminUsage()
	}
}

func adminUsage() {
	fmt.Printf("Usage: %s dwf admin [deploy]\n", os.Args[0])
	os.Exit(1)
}
