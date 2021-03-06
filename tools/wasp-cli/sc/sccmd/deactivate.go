// +build ignore

package sccmd

import (
	"fmt"
	"os"

	"github.com/iotaledger/wasp/client/multiclient"
	"github.com/iotaledger/wasp/tools/wasp-cli/config"
)

func deactivateCmd(args []string) {
	if len(args) != 2 {
		deactivateUsage()
	}

	scAddress, err := ledgerstate.AddressFromBase58EncodedString(args[0])
	log.Check(err)
	committee := parseIntList(args[1])

	log.Check(multiclient.New(config.CommitteeApi(committee)).DeactivateChain(&scAddress))
}

func deactivateUsage() {
	fmt.Printf("Usage: %s sc deactivate <sc-address> <committee>\n", os.Args[0])
	fmt.Printf("Example: %s sc deactivate aBcD...wXyZ '0,1,2,3'\n", os.Args[0])
	os.Exit(1)
}
