package facmd

import (
	"fmt"
	"os"
	"strconv"

	"github.com/iotaledger/wasp/tools/wwallet/sc/fa"
	"github.com/iotaledger/wasp/tools/wwallet/wallet"
)

func adminCmd(args []string) {
	if len(args) < 1 {
		adminUsage()
	}

	switch args[0] {
	case "deploy":
		check(fa.Config.Deploy(wallet.Load().SignatureScheme()))

	case "set-owner-margin":
		if len(args) != 2 {
			fa.Config.PrintUsage("admin set-owner-margin <promilles>")
			os.Exit(1)
		}
		p, err := strconv.Atoi(args[1])
		check(err)
		_, err = fa.Client().SetOwnerMargin(int64(p))
		check(err)

	default:
		adminUsage()
	}
}

func adminUsage() {
	fmt.Printf("Usage: %s fr admin [deploy|set-owner-margin <promilles>]\n", os.Args[0])
	os.Exit(1)
}
