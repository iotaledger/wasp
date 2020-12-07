package main

import (
	"fmt"
	"os"
	"strings"

	"github.com/iotaledger/wasp/tools/wasp-cli/chain"
	"github.com/iotaledger/wasp/tools/wasp-cli/config"
	"github.com/iotaledger/wasp/tools/wasp-cli/decode"
	"github.com/iotaledger/wasp/tools/wasp-cli/log"
	"github.com/iotaledger/wasp/tools/wasp-cli/wallet"
	"github.com/spf13/pflag"
)

func usage(commands map[string]func([]string), flags *pflag.FlagSet) {
	cmdNames := make([]string, 0)
	for k := range commands {
		cmdNames = append(cmdNames, k)
	}

	fmt.Printf("Usage: %s [options] [%s]\n", os.Args[0], strings.Join(cmdNames, "|"))
	flags.PrintDefaults()
	os.Exit(1)
}

func main() {
	commands := map[string]func([]string){}
	flags := pflag.NewFlagSet("global flags", pflag.ExitOnError)

	log.InitCommands(commands, flags)
	config.InitCommands(commands, flags)
	wallet.InitCommands(commands, flags)
	chain.InitCommands(commands, flags)
	decode.InitCommands(commands, flags)

	log.Check(flags.Parse(os.Args[1:]))

	config.Read()

	if flags.NArg() < 1 {
		usage(commands, flags)
	}

	cmd, ok := commands[flags.Arg(0)]
	if !ok {
		usage(commands, flags)
	}
	cmd(flags.Args()[1:])
}
