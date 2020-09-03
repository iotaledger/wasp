package trcmd

import (
	"fmt"
	"os"

	"github.com/iotaledger/wasp/tools/wwallet/sc/tr"
	"github.com/spf13/pflag"
)

func InitCommands(commands map[string]func([]string), flags *pflag.FlagSet) {
	commands["tr"] = cmd
	flags.AddFlagSet(tr.Config.HookFlags())
}

var subcmds = map[string]func([]string){
	"set":    tr.Config.HandleSetCmd,
	"admin":  adminCmd,
	"status": statusCmd,
	"query":  queryCmd,
	"mint":   mintCmd,
}

func cmd(args []string) {
	tr.Config.HandleCmd(args, subcmds)
}

func check(err error) {
	if err != nil {
		fmt.Printf("error: %s\n", err)
		os.Exit(1)
	}
}
