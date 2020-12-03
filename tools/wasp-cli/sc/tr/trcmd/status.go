package trcmd

import (
	"fmt"
	"time"

	"github.com/iotaledger/wasp/tools/wasp-cli/sc/tr"
	"github.com/iotaledger/wasp/tools/wasp-cli/util"
)

func statusCmd(args []string) {
	status, err := tr.Client().FetchStatus(true)
	check(err)

	util.DumpSCStatus(tr.Config, status.SCStatus)
	fmt.Printf("  Registry (latest first):\n")
	for _, tm := range status.RegistrySortedByMintTimeDesc {
		fmt.Printf("  - Color: %s\n", tm.Color)
		fmt.Printf("    Supply: %d\n", tm.Supply)
		fmt.Printf("    Minted by: %s\n", tm.MintedBy)
		fmt.Printf("    Owner: %s\n", tm.Owner)
		fmt.Printf("    Created: %s\n", time.Unix(0, tm.Created).UTC())
		fmt.Printf("    Updated: %s\n", time.Unix(0, tm.Updated).UTC())
		fmt.Printf("    Description: %s\n", tm.Description)
		fmt.Printf("    UserDefined: %v\n", tm.UserDefined)
	}
}
