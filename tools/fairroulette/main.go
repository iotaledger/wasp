package main

import (
	"fmt"
	"os"
	"strings"

	"github.com/iotaledger/wasp/tools/fairroulette/admin"
	"github.com/iotaledger/wasp/tools/fairroulette/client"
	"github.com/iotaledger/wasp/tools/fairroulette/wallet"
	"github.com/spf13/pflag"
	"github.com/spf13/viper"
)

func check(err error) {
	if err != nil {
		fmt.Printf("error: %s\n", err)
		os.Exit(1)
	}
}

var commands = map[string]func([]string){
	"wallet":      wallet.Cmd,
	"admin":       admin.AdminCmd,
	"set-address": client.SetSCAddressCmd,
	"status":      client.StatusCmd,
	"bet":         client.BetCmd,
}

func usage(globalFlags *pflag.FlagSet) {
	cmdNames := make([]string, 0)
	for k, _ := range commands {
		cmdNames = append(cmdNames, k)
	}

	fmt.Printf("Usage: %s [options] [%s]\n", os.Args[0], strings.Join(cmdNames, "|"))
	globalFlags.PrintDefaults()
	os.Exit(1)
}

func main() {
	globalFlags := pflag.NewFlagSet("global flags", pflag.ExitOnError)
	configPath := globalFlags.StringP("config", "c", "fairroulette.json", "path to fairroulette.json")
	pflag.IntP("address-index", "i", 0, "address index")
	globalFlags.Parse(os.Args[1:])

	viper.SetConfigFile(*configPath)
	viper.ReadInConfig()

	if globalFlags.NArg() < 1 {
		usage(globalFlags)
	}

	cmd, ok := commands[globalFlags.Arg(0)]
	if !ok {
		usage(globalFlags)
	}
	cmd(globalFlags.Args()[1:])
}
