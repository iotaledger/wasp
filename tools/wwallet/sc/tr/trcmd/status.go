package trcmd

import (
	"fmt"
	"time"

	"github.com/iotaledger/wasp/tools/wwallet/sc/tr"
)

func statusCmd(args []string) {
	status, err := tr.Client().FetchStatus()
	check(err)

	fmt.Printf("%s smart contract status:\n", tr.Config.Name)
	for color, tm := range status.Registry {
		fmt.Printf("  - Color: %s\n", color)
		fmt.Printf("    Supply: %d\n", tm.Supply)
		fmt.Printf("    Minted by: %s\n", tm.MintedBy)
		fmt.Printf("    Owner: %s\n", tm.Owner)
		fmt.Printf("    Created: %s\n", time.Unix(0, tm.Created).UTC())
		fmt.Printf("    Updated: %s\n", time.Unix(0, tm.Updated).UTC())
		fmt.Printf("    Description: %s\n", tm.Description)
		fmt.Printf("    UserDefined: %v\n", tm.UserDefined)
	}
}
