package client

import (
	"fmt"
	"os"

	"github.com/iotaledger/wasp/tools/fairroulette/config"
)

func SetSCAddressCmd(args []string) {
	if len(args) != 1 {
		fmt.Printf("Usage: %s set-address <address>\n", os.Args[0])
		os.Exit(1)
	}
	config.SetSCAddress(args[0])
}
