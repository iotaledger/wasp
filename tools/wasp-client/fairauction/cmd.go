package fairauction

import (
	"github.com/iotaledger/wasp/tools/wasp-client/config"
	"github.com/spf13/pflag"
)

var scConfig *config.SCConfig

func HookFlags() *pflag.FlagSet {
	scConfig = config.NewSC("fairauction", "fa")
	return scConfig.Flags
}

var commands = map[string]func([]string){
	"set":    scConfig.HandleSetCmd,
	"status": statusCmd,
}

func Cmd(args []string) {
	scConfig.HandleCmd(args, commands)
}
