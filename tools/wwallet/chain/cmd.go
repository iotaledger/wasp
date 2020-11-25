package chain

import (
	"fmt"
	"os"
	"strings"

	"github.com/spf13/pflag"
)

func InitCommands(commands map[string]func([]string), flags *pflag.FlagSet) {
	commands["chain"] = chainCmd

	fs := pflag.NewFlagSet("chain", pflag.ExitOnError)
	initDeployFlags(fs)
	initAliasFlags(fs)
	flags.AddFlagSet(fs)
}

var subcmds = map[string]func([]string){
	"list":           listCmd,
	"deploy":         deployCmd,
	"info":           infoCmd,
	"list-contracts": listContractsCmd,
}

func chainCmd(args []string) {
	if len(args) < 1 {
		usage()
	}
	subcmd, ok := subcmds[args[0]]
	if !ok {
		usage()
	}
	subcmd(args[1:])
}

func usage() {
	cmdNames := make([]string, 0)
	for k := range subcmds {
		cmdNames = append(cmdNames, k)
	}

	fmt.Printf("Usage: %s chain [%s]\n", os.Args[0], strings.Join(cmdNames, "|"))
	os.Exit(1)
}

func check(err error) {
	if err != nil {
		fmt.Printf("error: %s\n", err)
		os.Exit(1)
	}
}
