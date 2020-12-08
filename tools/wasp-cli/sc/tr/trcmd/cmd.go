// +build ignore

package trcmd

import (
	"fmt"
	"os"

	"github.com/iotaledger/wasp/tools/wasp-cli/sc/tr"
)

func InitCommands(commands map[string]func([]string)) {
	commands["tr"] = cmd
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
