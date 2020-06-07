package main

import (
	"flag"
	"fmt"
	"os"
	"os/signal"

	"github.com/iotaledger/wasp/tools/cluster"
)

func check(err error) {
	if err != nil {
		fmt.Printf("Error: %s\n", err)
		os.Exit(1)
	}
}

func main() {
	globalFlags := flag.NewFlagSet("", flag.ExitOnError)
	configPath := globalFlags.String("config", ".", "Config path")
	dataPath := globalFlags.String("data", "cluster-data", "Data path")
	globalFlags.Parse(os.Args[1:])

	wasps, err := cluster.New(*configPath, *dataPath)
	check(err)

	if globalFlags.NArg() < 1 {
		fmt.Printf("Usage: %s [options] [init|start|gendksets]\n", os.Args[0])
		globalFlags.PrintDefaults()
		return
	}

	switch globalFlags.Arg(0) {

	case "init":
		initFlags := flag.NewFlagSet("init", flag.ExitOnError)
		resetDataPath := initFlags.Bool("r", false, "Reset data path if it exists")
		initFlags.Parse(globalFlags.Args()[1:])
		err = wasps.Init(*resetDataPath)
		check(err)

	case "start":
		err = wasps.Start()
		check(err)

		fmt.Printf("Press CTRL-C to stop\n")
		c := make(chan os.Signal, 1)
		signal.Notify(c, os.Interrupt)
		<-c
		wasps.Wait()

	case "gendksets":
		err = wasps.Start()
		check(err)
		err = wasps.GenerateDKSets()
		check(err)
		wasps.Stop()

	case "origintx":
		// example
		err = wasps.Start()
		check(err)
		err = wasps.CreateOriginTx()
		check(err)
		wasps.Stop()
	}
}
