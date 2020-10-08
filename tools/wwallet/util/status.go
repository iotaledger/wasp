package util

import (
	"fmt"

	"github.com/iotaledger/goshimmer/dapps/valuetransfers/packages/balance"
	"github.com/iotaledger/wasp/client/scclient"
	"github.com/iotaledger/wasp/tools/wwallet/sc"
)

func DumpSCStatus(sc *sc.Config, status *scclient.SCStatus) {
	fmt.Printf("%s smart contract status:\n", sc.Name)
	fmt.Printf("  Program hash: %s\n", status.ProgramHash)
	fmt.Printf("  Description: %s\n", status.Description)
	fmt.Printf("  Owner address: %s\n", status.OwnerAddress)
	fmt.Printf("  SC address: %s\n", status.SCAddress)
	fmt.Printf("  Minimum reward: %d\n", status.MinimumReward)
	dumpBalance(status.Balance)
	fmt.Printf("  ----\n")
}

func dumpBalance(bal map[balance.Color]int64) {
	fmt.Printf("  SC balance:\n")
	for color, amount := range bal {
		fmt.Printf("    %s: %d\n", color, amount)
	}
}
