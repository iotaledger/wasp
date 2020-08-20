package main

import (
	"fmt"
	"os"
	"strings"

	"github.com/iotaledger/wasp/tools/wasp-client/config"
	"github.com/iotaledger/wasp/tools/wasp-client/config/fa"
	"github.com/iotaledger/wasp/tools/wasp-client/config/fr"
	"github.com/iotaledger/wasp/tools/wasp-client/dashboard"
	"github.com/iotaledger/wasp/tools/wasp-client/fairauction"
	"github.com/iotaledger/wasp/tools/wasp-client/fairroulette"
	"github.com/iotaledger/wasp/tools/wasp-client/wallet"
	"github.com/spf13/pflag"
)

func check(err error) {
	if err != nil {
		fmt.Printf("error: %s\n", err)
		os.Exit(1)
	}
}

var commands = map[string]func([]string){
	"wallet":    wallet.Cmd,
	"set":       config.SetCmd,
	"fr":        fairroulette.Cmd,
	"fa":        fairauction.Cmd,
	"dashboard": dashboard.Cmd,
}

func usage(flags *pflag.FlagSet) {
	cmdNames := make([]string, 0)
	for k := range commands {
		cmdNames = append(cmdNames, k)
	}

	fmt.Printf("Usage: %s [options] [%s]\n", os.Args[0], strings.Join(cmdNames, "|"))
	flags.PrintDefaults()
	os.Exit(1)
}

func main() {
	flags := pflag.NewFlagSet("global flags", pflag.ExitOnError)
	flags.AddFlagSet(config.HookFlags())
	flags.AddFlagSet(wallet.HookFlags())
	flags.AddFlagSet(fr.Config.HookFlags())
	flags.AddFlagSet(fa.Config.HookFlags())
	check(flags.Parse(os.Args[1:]))

	config.Read()

	if flags.NArg() < 1 {
		usage(flags)
	}

	cmd, ok := commands[flags.Arg(0)]
	if !ok {
		usage(flags)
	}
	cmd(flags.Args()[1:])
}
