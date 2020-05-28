package main

import (
	"fmt"
	"os"

	"github.com/iotaledger/wasp/tools/cluster"
)

func main() {
	if len(os.Args) < 3 {
		fmt.Printf("usage: %s <config-path> <data-path>\n", os.Args[0])
		os.Exit(1)
	}

	configPath := os.Args[1]
	dataPath := os.Args[2]

	wasps := cluster.New(configPath, dataPath)

	wasps.Init()

	wasps.Start()

	addr := wasps.GenerateNewDistributedKeySet(3)
	fmt.Printf("Generated key set with address %s\n", addr)

	wasps.Stop()
}
