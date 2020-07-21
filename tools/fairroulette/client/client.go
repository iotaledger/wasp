package client

import (
	"fmt"
	"os"

	"github.com/iotaledger/wasp/tools/fairroulette/config"
)

func SetCmd(args []string) {
	if len(args) != 2 {
		fmt.Printf("Usage: %s set <key> <value>\n", os.Args[0])
		os.Exit(1)
	}
	config.Set(args[0], args[1])
}
