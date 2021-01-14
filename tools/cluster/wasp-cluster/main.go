package main

import (
	"fmt"
	"os"
	"os/signal"

	"github.com/iotaledger/wasp/tools/cluster"
	"github.com/spf13/pflag"
)

func check(err error) {
	if err != nil {
		fmt.Printf("[%s] error: %s\n", os.Args[0], err)
		os.Exit(1)
	}
}

func usage() {
	fmt.Printf("Usage: %s [init <path>|start] [options]\n", os.Args[0])
	os.Exit(1)
}

func main() {
	if len(os.Args) < 2 {
		usage()
	}

	switch os.Args[1] {
	case "init":
		initFlags := pflag.NewFlagSet("init", pflag.ExitOnError)

		config := cluster.DefaultConfig()
		forceRemove := initFlags.BoolP("force", "f", false, "Force removing cluster directory if it exists")
		templatesPath := initFlags.StringP("templates-path", "t", ".", "Where to find alternative wasp & goshimmer config.json templates (optional)")
		initFlags.IntVarP(&config.Wasp.NumNodes, "num-nodes", "n", config.Wasp.NumNodes, "Amount of wasp nodes")
		initFlags.IntVarP(&config.Wasp.FirstApiPort, "first-api-port", "a", config.Wasp.FirstApiPort, "First wasp API port")
		initFlags.IntVarP(&config.Wasp.FirstPeeringPort, "first-peering-port", "p", config.Wasp.FirstPeeringPort, "First wasp Peering port")
		initFlags.IntVarP(&config.Wasp.FirstNanomsgPort, "first-nanomsg-port", "u", config.Wasp.FirstNanomsgPort, "First wasp nanomsg (publisher) port")
		initFlags.IntVarP(&config.Wasp.FirstDashboardPort, "first-dashboard-port", "h", config.Wasp.FirstDashboardPort, "First wasp dashboard port")
		initFlags.IntVarP(&config.Goshimmer.ApiPort, "goshimmer-api-port", "w", config.Goshimmer.ApiPort, "Goshimmer API port")
		initFlags.BoolVarP(&config.Goshimmer.Provided, "goshimmer-provided", "g", config.Goshimmer.Provided, "If true, Goshimmer node will not be spawn")

		err := initFlags.Parse(os.Args[2:])
		check(err)

		if initFlags.NArg() != 1 {
			fmt.Printf("Usage: %s init <path> [options]\n", os.Args[0])
			initFlags.PrintDefaults()
			os.Exit(1)
		}

		dataPath := initFlags.Arg(0)
		err = cluster.New("cluster", config).InitDataPath(*templatesPath, dataPath, *forceRemove)
		check(err)

	case "start":
		startFlags := pflag.NewFlagSet("start", pflag.ExitOnError)

		err := startFlags.Parse(os.Args[2:])
		check(err)

		if startFlags.NArg() > 1 {
			fmt.Printf("Usage: %s start [path] [options]\n", os.Args[0])
			startFlags.PrintDefaults()
			os.Exit(1)
		}

		dataPath := "."
		if startFlags.NArg() == 1 {
			dataPath = startFlags.Arg(0)
		}

		exists, err := cluster.ConfigExists(dataPath)
		check(err)
		if !exists {
			check(fmt.Errorf("%s/cluster.json not found. Call `%s init` first.", dataPath, os.Args[0]))
		}

		config, err := cluster.LoadConfig(dataPath)
		check(err)

		clu := cluster.New("cluster", config)

		err = clu.Start(dataPath)
		check(err)
		fmt.Printf("-----------------------------------------------------------------\n")
		fmt.Printf("           The cluster started\n")
		fmt.Printf("-----------------------------------------------------------------\n")

		waitCtrlC()
		clu.Wait()

	default:
		usage()
	}
}

func waitCtrlC() {
	fmt.Printf("[%s] Press CTRL-C to stop\n", os.Args[0])
	c := make(chan os.Signal, 1)
	signal.Notify(c, os.Interrupt)
	<-c
}
