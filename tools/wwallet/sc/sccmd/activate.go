package sccmd

import (
	"fmt"
	"os"

	"github.com/iotaledger/goshimmer/dapps/valuetransfers/packages/address"
	"github.com/iotaledger/wasp/packages/apilib"
	"github.com/iotaledger/wasp/tools/wwallet/config"
)

func activateCmd(args []string) {
	if len(args) != 2 {
		activateUsage()
	}

	scAddress, err := address.FromBase58(args[0])
	check(err)
	committee := parseIntList(args[1])

	for _, host := range config.CommitteeApi(committee) {
		check(apilib.ActivateSC(host, &scAddress))
	}
}

func activateUsage() {
	fmt.Printf("Usage: %s sc activate <sc-address> <committee>\n", os.Args[0])
	fmt.Printf("Example: %s sc activate aBcD...wXyZ '0,1,2,3'\n", os.Args[0])
	os.Exit(1)
}
