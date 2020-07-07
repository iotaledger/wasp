// fr-client allows to use the FairRoulette smart contract from the command line
package main

import (
	"flag"
	"fmt"
	"os"
)

type clientConfig struct {
	waspApi string
}

var config = clientConfig{
	waspApi: "127.0.0.1:9090",
}

func main() {
	globalFlags := flag.NewFlagSet("", flag.ExitOnError)
	scAddress := globalFlags.String("sc", "", "SC Address")
	globalFlags.Parse(os.Args[1:])

	if globalFlags.NArg() < 1 || len(*scAddress) == 0 {
		usage(globalFlags)
	}

	switch globalFlags.Arg(0) {
	case "show":
		dumpState(*scAddress)
	default:
		usage(globalFlags)
	}
}

func check(err error) {
	if err != nil {
		panic(err)
	}
}

func usage(globalFlags *flag.FlagSet) {
	fmt.Printf("Usage: %s [options] [show]\n", os.Args[0])
	globalFlags.PrintDefaults()
	os.Exit(1)
}
