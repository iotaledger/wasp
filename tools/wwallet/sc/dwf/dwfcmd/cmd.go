package dwfcmd

import (
	"fmt"
	"os"

	"github.com/iotaledger/wasp/tools/wwallet/sc/dwf"
	"github.com/spf13/pflag"
)

func InitCommands(commands map[string]func([]string), flags *pflag.FlagSet) {
	commands["dwf"] = cmd
	flags.AddFlagSet(dwf.Config.HookFlags())
}

var subcmds = map[string]func([]string){
	"set":      dwf.Config.HandleSetCmd,
	"admin":    adminCmd,
	"donate":   donateCmd,
	"withdraw": withdrawCmd,
	"status":   statusCmd,
}

func cmd(args []string) {
	dwf.Config.HandleCmd(args, subcmds)
}

func check(err error) {
	if err != nil {
		fmt.Printf("error: %s\n", err)
		os.Exit(1)
	}
}
