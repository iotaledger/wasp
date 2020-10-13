package sccmd

import (
	"fmt"
	"os"

	"github.com/iotaledger/goshimmer/dapps/valuetransfers/packages/address"
	"github.com/iotaledger/wasp/client/multiclient"
	"github.com/iotaledger/wasp/tools/wwallet/config"
)

func deactivateCmd(args []string) {
	if len(args) != 2 {
		deactivateUsage()
	}

	scAddress, err := address.FromBase58(args[0])
	check(err)
	committee := parseIntList(args[1])

	check(multiclient.New(config.CommitteeApi(committee)).DeactivateSC(&scAddress))
}

func deactivateUsage() {
	fmt.Printf("Usage: %s sc deactivate <sc-address> <committee>\n", os.Args[0])
	fmt.Printf("Example: %s sc deactivate aBcD...wXyZ '0,1,2,3'\n", os.Args[0])
	os.Exit(1)
}
