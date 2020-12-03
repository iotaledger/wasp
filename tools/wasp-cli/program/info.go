package program

import (
	"fmt"
	"os"

	"github.com/iotaledger/wasp/client"
	"github.com/iotaledger/wasp/packages/hashing"
	"github.com/iotaledger/wasp/tools/wasp-cli/config"
)

func infoCmd(args []string) {
	if len(args) != 2 {
		infoUsage()
	}

	hash, err := hashing.HashValueFromBase58(args[0])
	check(err)
	nodes := parseIntList(args[1])

	for _, host := range config.CommitteeApi(nodes) {
		md, err := client.NewWaspClient(host).GetProgramMetadata(&hash)
		check(err)

		fmt.Printf("Node %s:\n", host)
		if md == nil {
			fmt.Printf("  Program not found\n")
		} else {
			fmt.Printf("  Description: %s\n", md.Description)
			fmt.Printf("  VMType: %s\n", md.VMType)
		}
	}
}

func infoUsage() {
	fmt.Printf("Usage: %s program info <program-hash> <nodes>\n", os.Args[0])
	fmt.Printf("Example: %s program info aBcD...wXyZ '0,1,2,3'\n", os.Args[0])
	os.Exit(1)
}
