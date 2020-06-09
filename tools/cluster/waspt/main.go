// waspt is the program for initial setup of the testing environment for the wasp cluster
// Commands:
//    `waspt init` creates directory structure with respective config files
//    `waspt gendksets` starts wasps and generates keys.json file
//    `waspt start` start the cluster and and stops when interrupted by "ctrl-C"
package main

import (
	"flag"
	"fmt"
	"github.com/iotaledger/wasp/tools/cluster/tests/wasptest"
	"os"
	"os/signal"

	"github.com/iotaledger/wasp/tools/cluster"
)

func check(err error) {
	if err != nil {
		fmt.Printf("[cluster] Waspt error: %s. Exitting...\n", err)
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
		fmt.Printf("-----------------------------------------------------------------\n")
		fmt.Printf("           The cluster started\n")
		fmt.Printf("-----------------------------------------------------------------\n")

		waitCtrlC()
		wasps.Wait()

	case "gendksets":
		err = wasps.Start()
		check(err)
		fmt.Printf("-----------------------------------------------------------------\n")
		fmt.Printf("           Generate DKSets\n")
		fmt.Printf("-----------------------------------------------------------------\n")
		err = wasps.GenerateDKSets()
		check(err)
		wasps.Stop()

	case "test0":
		err = wasps.Start()
		check(err)

		err = wasptest.Put3BootupRecords(wasps)
		check(err)

		waitCtrlC()
		wasps.Wait()

	case "test1":
		err = wasps.Start()
		check(err)

		err = wasptest.Put3BootupRecords(wasps)
		check(err)

		err = wasptest.Activate1SC(wasps)
		check(err)

		waitCtrlC()
		wasps.Wait()

	case "test2":
		err = wasps.Start()
		check(err)

		err = wasptest.Put3BootupRecords(wasps)
		check(err)

		err = wasptest.Activate3SC(wasps)
		check(err)

		waitCtrlC()
		wasps.Wait()

	case "test3":
		err = wasps.Start()
		check(err)

		err = wasptest.Put3BootupRecords(wasps)
		check(err)

		err = wasptest.Activate1SC(wasps)
		check(err)

		err = wasptest.Send1Request(wasps)
		check(err)

		waitCtrlC()
		wasps.Wait()
	}
}

func waitCtrlC() {
	fmt.Printf("[waspt] Press CTRL-C to stop\n")
	c := make(chan os.Signal, 1)
	signal.Notify(c, os.Interrupt)
	<-c
}
