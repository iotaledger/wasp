package trcmd

import (
	"fmt"
	"os"
	"time"

	"github.com/iotaledger/wasp/packages/util"
	"github.com/iotaledger/wasp/tools/wasp-cli/sc/tr"
)

func queryCmd(args []string) {
	if len(args) != 1 {
		fmt.Printf("Usage: %s tr query <color>\n", os.Args[0])
		os.Exit(1)
	}

	color, err := util.ColorFromString(args[0])
	check(err)

	tm, err := tr.Client().Query(&color)
	check(err)

	fmt.Printf("Color: %s\n", color)
	fmt.Printf("Supply: %d\n", tm.Supply)
	fmt.Printf("Minted by: %s\n", tm.MintedBy)
	fmt.Printf("Owner: %s\n", tm.Owner)
	fmt.Printf("Created: %s\n", time.Unix(0, tm.Created).UTC())
	fmt.Printf("Updated: %s\n", time.Unix(0, tm.Updated).UTC())
	fmt.Printf("Description: %s\n", tm.Description)
	fmt.Printf("UserDefined: %v\n", tm.UserDefined)
}
