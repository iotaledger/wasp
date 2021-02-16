// +build ignore

package dwfcmd

import (
	"fmt"
	"os"
	"strconv"

	"github.com/iotaledger/wasp/tools/wasp-cli/sc/dwf"
)

func withdrawCmd(args []string) {
	if len(args) != 1 {
		fmt.Printf("Usage: %s dwf withdraw <amount>\n", os.Args[0])
		os.Exit(1)
	}

	amount, err := strconv.Atoi(args[0])
	log.Check(err)

	_, err = dwf.Client().Withdraw(int64(amount))
	log.Check(err)
}
