package main

import (
	"bytes"
	"fmt"
	"github.com/iotaledger/wasp/packages/vm/wasmhost"
	"io"
	"io/ioutil"
	"os"
	"os/exec"
	"os/signal"
	"path"
	"strings"

	"github.com/iotaledger/wasp/tools/cluster"
	"github.com/spf13/pflag"
)

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

func main() {
	commonFlags := pflag.NewFlagSet("common flags", pflag.ExitOnError)

	templatesPath := commonFlags.StringP("templates-path", "t", ".", "Where to find alternative wasp & goshimmer config.json templates (optional)")

	config := cluster.DefaultConfig()
	commonFlags.IntVarP(&config.Wasp.NumNodes, "num-nodes", "n", config.Wasp.NumNodes, "Amount of wasp nodes")
	commonFlags.IntVarP(&config.Wasp.FirstApiPort, "first-api-port", "a", config.Wasp.FirstApiPort, "First wasp API port")
	commonFlags.IntVarP(&config.Wasp.FirstPeeringPort, "first-peering-port", "p", config.Wasp.FirstPeeringPort, "First wasp Peering port")
	commonFlags.IntVarP(&config.Wasp.FirstNanomsgPort, "first-nanomsg-port", "u", config.Wasp.FirstNanomsgPort, "First wasp nanomsg (publisher) port")
	commonFlags.IntVarP(&config.Wasp.FirstDashboardPort, "first-dashboard-port", "h", config.Wasp.FirstDashboardPort, "First wasp dashboard port")
	commonFlags.IntVarP(&config.Goshimmer.ApiPort, "goshimmer-api-port", "w", config.Goshimmer.ApiPort, "Goshimmer API port")
	commonFlags.BoolVarP(&config.Goshimmer.Provided, "goshimmer-provided", "g", config.Goshimmer.Provided, "If true, Goshimmer node will not be spawn")

	if len(os.Args) < 2 {
		usage(commonFlags)
	}

	switch os.Args[1] {
	case "init":
		flags := pflag.NewFlagSet("init", pflag.ExitOnError)
		forceRemove := flags.BoolP("force", "f", false, "Force removing cluster directory if it exists")
		flags.AddFlagSet(commonFlags)

		err := flags.Parse(os.Args[2:])
		check(err)

		if flags.NArg() != 1 {
			fmt.Printf("Usage: %s init <path> [options]\n", os.Args[0])
			flags.PrintDefaults()
			os.Exit(1)
		}

		dataPath := flags.Arg(0)
		err = cluster.New("cluster", config).InitDataPath(*templatesPath, dataPath, *forceRemove)
		check(err)

	case "start":
		flags := pflag.NewFlagSet("start", pflag.ExitOnError)
		disposable := flags.BoolP("disposable", "d", false, "If set, run a disposable cluster in a temporary directory (no need for init, automatically removed when stopped)")
		flags.AddFlagSet(commonFlags)

		err := flags.Parse(os.Args[2:])
		check(err)

		if flags.NArg() > 1 {
			fmt.Printf("Usage: %s start [path] [options]\n", os.Args[0])
			flags.PrintDefaults()
			os.Exit(1)
		}

		dataPath := "."
		if flags.NArg() == 1 {
			if *disposable {
				check(fmt.Errorf("[path] and -d are mutually exclusive"))
			}
			dataPath = flags.Arg(0)
		} else if *disposable {
			dataPath, err = ioutil.TempDir(os.TempDir(), "wasp-cluster-*")
			check(err)
		}

		if !*disposable {
			exists, err := cluster.ConfigExists(dataPath)
			check(err)
			if !exists {
				check(fmt.Errorf("%s/cluster.json not found. Call `%s init` first.", dataPath, os.Args[0]))
			}

			config, err = cluster.LoadConfig(dataPath)
			check(err)
		}

		clu := cluster.New("wasp-cluster", config)

		if *disposable {
			check(clu.InitDataPath(*templatesPath, dataPath, true))
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

var scripts = map[string]func(*WaspCli){
	"inc": inccounterScript,
}

type WaspCli struct {
	dir string
}

func (w *WaspCli) Run(args ...string) {
	// -w: wait for requests
	// -d: debug output
	cmd := exec.Command("wasp-cli", append([]string{"-w", "-d"}, args...)...)
	cmd.Dir = w.dir

	stdout := &bytes.Buffer{}
	cmd.Stdout = stdout
	stderr := &bytes.Buffer{}
	cmd.Stderr = stderr

	err := cmd.Run()

	outStr, errStr := stdout.String(), stderr.String()
	if err != nil {
		check(fmt.Errorf(
			"cmd `wasp-cli %s` failed\n%w\noutput:\n%s",
			strings.Join(args, " "),
			err,
			outStr+errStr,
		))
	}
}

func (w *WaspCli) copyFile(fname string) {
	source, err := os.Open(fname)
	check(err)
	defer source.Close()

	dst := path.Join(w.dir, path.Base(fname))
	destination, err := os.Create(dst)
	check(err)
	defer destination.Close()

	_, err = io.Copy(destination, source)
	check(err)
}

func inccounterScript(w *WaspCli) {
	vmtype := "wasmtimevm"
	name := "inccounter"
	description := "inccounter SC"
	file := wasmhost.WasmPath("inccounter_bg.wasm")

	w.copyFile(path.Join("wasm", file))

	w.Run("init")
	w.Run("request-funds")
	w.Run("chain", "deploy", "--chain=chain1", "--committee=0,1,2,3", "--quorum=3")
	w.Run("chain", "deploy-contract", vmtype, name, description, file)
	w.Run("chain", "post-request", name, "increment")
}

func runScript(dataPath string, scriptName string) {
	f := scripts[scriptName]
	if f == nil {
		check(fmt.Errorf("Script %s not found", scriptName))
	}
	f(&WaspCli{dataPath})
}

func waitCtrlC() {
	fmt.Printf("[%s] Press CTRL-C to stop\n", os.Args[0])
	c := make(chan os.Signal, 1)
	signal.Notify(c, os.Interrupt)
	<-c
}
