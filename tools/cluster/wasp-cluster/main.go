package main

import (
	"flag"
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"os/signal"

	"github.com/iotaledger/hive.go/logger"
	"github.com/iotaledger/wasp/packages/nodeconn"
	"github.com/iotaledger/wasp/packages/parameters"
	"github.com/iotaledger/wasp/packages/util/l1starter"
	"github.com/iotaledger/wasp/tools/cluster"
	"github.com/spf13/pflag"
)

const cmdName = "wasp-cluster"

func check(err error) {
	if err != nil {
		fmt.Printf("[%s] error: %s\n", os.Args[0], err)
		os.Exit(1)
	}
}

func usage(flags *pflag.FlagSet) {
	fmt.Printf("Usage: %s [init <path>|start] [options]\n", os.Args[0])
	flags.PrintDefaults()
	os.Exit(1)
}

//nolint:funlen
func main() {
	commonFlags := pflag.NewFlagSet("common flags", pflag.ExitOnError)

	templatesPath := commonFlags.StringP("templates-path", "t", ".", "Where to find alternative wasp & layer1 config.json templates (optional)")

	waspConfig := cluster.DefaultWaspConfig()

	commonFlags.IntVarP(&waspConfig.NumNodes, "num-nodes", "n", waspConfig.NumNodes, "Amount of wasp nodes")
	commonFlags.IntVarP(&waspConfig.FirstAPIPort, "first-api-port", "a", waspConfig.FirstAPIPort, "First wasp API port")
	commonFlags.IntVarP(&waspConfig.FirstPeeringPort, "first-peering-port", "p", waspConfig.FirstPeeringPort, "First wasp Peering port")
	commonFlags.IntVarP(&waspConfig.FirstNanomsgPort, "first-nanomsg-port", "u", waspConfig.FirstNanomsgPort, "First wasp nanomsg (publisher) port")
	commonFlags.IntVarP(&waspConfig.FirstDashboardPort, "first-dashboard-port", "h", waspConfig.FirstDashboardPort, "First wasp dashboard port")

	l1StarterFlags := flag.NewFlagSet("l1", flag.ExitOnError)
	l1 := l1starter.New(l1StarterFlags)
	commonFlags.AddGoFlagSet(l1StarterFlags)

	if len(os.Args) < 2 {
		usage(commonFlags)
	}

	parseFlags := func(flags *pflag.FlagSet) {
		err := flags.Parse(os.Args[2:])
		check(err)
	}

	if err := logger.InitGlobalLogger(parameters.Init()); err != nil {
		panic(err)
	}

	switch os.Args[1] {
	case "init":
		flags := pflag.NewFlagSet("init", pflag.ExitOnError)
		forceRemove := flags.BoolP("force", "f", false, "Force removing cluster directory if it exists")
		flags.AddFlagSet(commonFlags)
		parseFlags(flags)

		if flags.NArg() != 1 {
			fmt.Printf("Usage: %s init <path> [options]\n", os.Args[0])
			flags.PrintDefaults()
			os.Exit(1)
		}

		if l1.PrivtangleEnabled() {
			fmt.Printf("non-disposable cluster and privtangle are mutually exclusive")
			os.Exit(1)
		}

		l1.StartPrivtangleIfNecessary(log.Printf)
		defer l1.Stop()

		dataPath := flags.Arg(0)
		clusterConfig := cluster.NewConfig(
			waspConfig,
			l1.Config,
		)
		log := logger.NewLogger(cmdName)
		nodeconn.NewL1Client(clusterConfig.L1, log) // indirectly initializes parameters.L1
		err := cluster.New(cmdName, clusterConfig, dataPath, nil, log).InitDataPath(*templatesPath, *forceRemove)
		check(err)

	case "start":
		flags := pflag.NewFlagSet("start", pflag.ExitOnError)
		disposable := flags.BoolP("disposable", "d", false, "If set, run a disposable cluster in a temporary directory (no need for init, automatically removed when stopped)")
		flags.AddFlagSet(commonFlags)
		parseFlags(flags)

		if flags.NArg() > 1 {
			fmt.Printf("Usage: %s start [path] [options]\n", os.Args[0])
			flags.PrintDefaults()
			os.Exit(1) // nolint:gocritic
		}

		var err error

		if !*disposable && l1.PrivtangleEnabled() {
			fmt.Printf("non-disposable cluster and privtangle are mutually exclusive")
			os.Exit(1)
		}

		dataPath := "."
		if flags.NArg() == 1 {
			if *disposable {
				check(fmt.Errorf("[path] and -d are mutually exclusive"))
			}
			dataPath = flags.Arg(0)
		} else if *disposable {
			dataPath, err = ioutil.TempDir(os.TempDir(), cmdName+"-*")
			check(err)
		}

		var clusterConfig *cluster.ClusterConfig
		if !*disposable {
			exists, err := cluster.ConfigExists(dataPath)
			check(err)
			if !exists {
				check(fmt.Errorf("%s/cluster.json not found. Call `%s init` first", dataPath, os.Args[0]))
			}

			clusterConfig, err = cluster.LoadConfig(dataPath)
			check(err)
		} else {
			l1.StartPrivtangleIfNecessary(log.Printf)
			defer l1.Stop()
			clusterConfig = cluster.NewConfig(
				waspConfig,
				l1.Config,
			)
		}

		log := logger.NewLogger(cmdName)
		nodeconn.NewL1Client(clusterConfig.L1, log) // indirectly initializes parameters.L1
		clu := cluster.New(cmdName, clusterConfig, dataPath, nil, log)

		if *disposable {
			check(clu.InitDataPath(*templatesPath, true))
			defer os.RemoveAll(dataPath)
		}

		err = clu.Start(dataPath)
		check(err)
		fmt.Printf("-----------------------------------------------------------------\n")
		fmt.Printf("           The cluster started\n")
		fmt.Printf("-----------------------------------------------------------------\n")

		waitCtrlC()
		clu.Wait()

	default:
		usage(commonFlags)
	}
}

func waitCtrlC() {
	fmt.Printf("[%s] Press CTRL-C to stop\n", os.Args[0])
	c := make(chan os.Signal, 1)
	signal.Notify(c, os.Interrupt)
	<-c
}
