// +build ignore

package util

import (
	"fmt"

	"github.com/iotaledger/wasp/client/chainclient"
)

func DumpSCStatus(name string, status *chainclient.SCStatus) {
	fmt.Printf("%s smart contract status:\n", name)
	fmt.Printf("  Program hash: %s\n", status.ProgramHash)
	fmt.Printf("  Description: %s\n", status.Description)
	fmt.Printf("  Owner address: %s\n", status.OwnerAddress.Base58())
	fmt.Printf("  SC address: %s\n", status.SCAddress.Base58())
	fmt.Printf("  Minimum reward: %d\n", status.MinimumReward)
	dumpBalance(status.Balance)
	fmt.Printf("  ----\n")
}

func dumpBalance(bal map[ledgerstate.Color]uint64) {
	fmt.Printf("  SC balance:\n")
	for color, amount := range bal {
		fmt.Printf("    %s: %d\n", color, amount)
	}
}
