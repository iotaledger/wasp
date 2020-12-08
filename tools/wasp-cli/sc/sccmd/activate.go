// +build ignore

package sccmd

import (
	"fmt"
	"os"

	"github.com/iotaledger/goshimmer/dapps/valuetransfers/packages/address"
	"github.com/iotaledger/wasp/client/multiclient"
	"github.com/iotaledger/wasp/tools/wasp-cli/config"
)

func activateCmd(args []string) {
	if len(args) != 2 {
		activateUsage()
	}

	scAddress, err := address.FromBase58(args[0])
	log.Check(err)
	committee := parseIntList(args[1])

	log.Check(multiclient.New(config.CommitteeApi(committee)).ActivateChain(&scAddress))
}

func activateUsage() {
	fmt.Printf("Usage: %s sc activate <sc-address> <committee>\n", os.Args[0])
	fmt.Printf("Example: %s sc activate aBcD...wXyZ '0,1,2,3'\n", os.Args[0])
	os.Exit(1)
}
