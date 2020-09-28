package frcmd

import (
	"github.com/iotaledger/wasp/tools/wwallet/sc/fr"
	"github.com/spf13/pflag"
)

func InitCommands(commands map[string]func([]string), flags *pflag.FlagSet) {
	commands["fr"] = cmd
	flags.AddFlagSet(fr.Config.HookFlags())
}

var subcmds = map[string]func([]string){
	"set":    fr.Config.HandleSetCmd,
	"admin":  adminCmd,
	"status": statusCmd,
	"bet":    betCmd,
}

func cmd(args []string) {
	fr.Config.HandleCmd(args, subcmds)
}
