// +build ignore

package dwfcmd

import (
	"fmt"
	"os"
	"strconv"

	"github.com/iotaledger/wasp/tools/wasp-cli/sc/dwf"
)

func donateCmd(args []string) {
	if len(args) != 2 {
		fmt.Printf("Usage: %s dwf donate <amount> <feedback>\n", os.Args[0])
		os.Exit(1)
	}

	amount, err := strconv.Atoi(args[0])
	log.Check(err)

	feedback := args[1]

	tx, err := dwf.Client().Donate(int64(amount), feedback)
	log.Check(err)
	fmt.Printf("success. Request transaction id: %s\n", tx.ID().String())
}
